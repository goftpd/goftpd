package ftp

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/big"
	"net"
	"sync"

	"github.com/goftpd/goftpd/vfs"
	"golang.org/x/sync/errgroup"
)

// ServerOpts is used to create a new Server. PublicIP, TLSCertFile,
// TLSKeyFile are all required.
type ServerOpts struct {
	Name         string `goftpd:"sitename_short"`
	LongName     string `goftpd:"sitename_long"`
	Host         string `goftpd:"host"`
	Port         int    `goftpd:"port"`
	PassivePorts []int  `goftpd:"passive_ports"`

	PublicIP string `goftpd:"public_ip"`

	TLSCertFile string `goftpd:"tls_cert_file"`
	TLSKeyFile  string `goftpd:"tls_key_file"`
	tlsConfig   *tls.Config
}

func (o *ServerOpts) SetTLSConfig(t *tls.Config) { o.tlsConfig = t }

// Server. Serves stuff.
type Server struct {
	*ServerOpts

	fs vfs.VFS

	sessionPool sync.Pool

	passivePortsMax *big.Int
	passivePorts    map[int64]struct{}
	passivePortsMtx sync.Mutex
}

// NewServer returns a Server using the supplied ServerOpts and VFS. Will
// fail if some required options are missing or it's unable to load
// the specified TLS cert/key files.
func NewServer(opts *ServerOpts, fs vfs.VFS) (*Server, error) {

	s := Server{
		ServerOpts: opts,
		fs:         fs,
		sessionPool: sync.Pool{
			New: func() interface{} {
				return &Session{}
			},
		},
		passivePorts:    make(map[int64]struct{}, 0),
		passivePortsMax: big.NewInt(int64(opts.PassivePorts[1] - opts.PassivePorts[0])),
	}

	return &s, nil
}

func (s *Server) TLSConfig() *tls.Config {
	return s.tlsConfig
}

// ListenAndServe creates a new tcp listener on the configured Host and Port.
// New connections are buffered down a channel before being given their own
// goroutine. Takes a context and attemps to shutdown on cancellation/deadline
func (s *Server) ListenAndServe(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	addr := net.JoinHostPort(s.Host, fmt.Sprintf("%d", s.Port))

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()

	conns := make(chan net.Conn, 10)

	var errg errgroup.Group

	errg.Go(func() error {
		for {
			conn, err := l.Accept()
			if err != nil {

				// check if this is a cancellation
				select {
				case <-ctx.Done():
					return nil
				default:
				}

				// check if this is temporary
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					continue
				}

				// fatal cancel ctx and return error
				cancel()

				return err
			}

			conns <- conn
		}

		return nil
	})

	errg.Go(func() error {
		for {
			select {
			case c := <-conns:
				go s.handleConnection(ctx, c)

			case <-ctx.Done():

				// attempt to close the listener
				if err := l.Close(); err != nil {
					return err
				}

				return nil
			}
		}

		return nil
	})

	if err := errg.Wait(); err != nil {
		return err
	}

	return nil
}

// handleConnection takes a context and a tcp connection and attempts to
// start a new session
func (server *Server) handleConnection(ctx context.Context, conn net.Conn) {
	session := server.sessionPool.Get().(*Session)
	session.Reset()
	defer server.sessionPool.Put(session)

	session.serve(ctx, server, conn)
}
