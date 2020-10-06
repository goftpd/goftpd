package ftp

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type passiveDataConn struct {
	ctx context.Context

	conn net.Conn

	dataProtected bool

	host string
	port int64

	onClose func()

	written int
	read    int

	err error

	sync.Mutex
}

func (s *Server) newPassiveDataConn(ctx context.Context, dataProtected bool) (*passiveDataConn, error) {
	var count int
	for {
		if count > 1000 {
			break
		}
		count++

		n, err := rand.Int(rand.Reader, s.passivePortsMax)
		if err != nil {
			return nil, err
		}

		s.passivePortsMtx.Lock()
		_, ok := s.passivePorts[n.Int64()]

		// we keep the lock open so we dont
		// have to worry about race conditions on
		// setting the value in the map

		if ok {
			s.passivePortsMtx.Unlock()
			continue
		} else {
			s.passivePorts[n.Int64()] = struct{}{}
		}

		s.passivePortsMtx.Unlock()

		port := n.Int64() + int64(s.PassivePorts[0])

		// if we want to support none tls, do it here
		var ln net.Listener

		addr := net.JoinHostPort(s.BindIP, strconv.Itoa(int(port)))

		if dataProtected {
			ln, err = tls.Listen("tcp", addr, s.tlsConfig)
		} else {
			ln, err = net.Listen("tcp", addr)
		}

		// check listen error
		if err != nil {
			if isErrorAddressAlreadyInUse(err) {
				continue
			}
			return nil, err
		}

		dc := passiveDataConn{
			ctx:           ctx,
			host:          s.PublicIP,
			port:          port,
			dataProtected: dataProtected,
			onClose: func() {
				s.passivePortsMtx.Lock()
				delete(s.passivePorts, port)
				s.passivePortsMtx.Unlock()
			},
		}

		go dc.Accept(ctx, ln)

		return &dc, nil
	}

	return nil, errors.New("unable to find a data port")

}

// Close implements the io.Closer interface and also allows us
// to call our onClose fn that will cleanup server state
func (d *passiveDataConn) Close() error {

	d.Lock()
	defer d.Unlock()

	if d.conn != nil {
		if err := d.conn.Close(); err != nil {
			return err
		}
	}

	if d.onClose != nil {
		d.onClose()
	}

	return nil
}

// Accept makes passiveDataConn context aware as well as concurrent. It
// locks the socket to prevent races on the underlying conn
func (d *passiveDataConn) Accept(ctx context.Context, ln net.Listener) {
	d.Lock()
	defer d.Unlock()

	// always close the listener
	defer ln.Close()

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*60))
	defer cancel()

	// make accept context aware
	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	d.conn, d.err = ln.Accept()

	if d.dataProtected {
		// handshake
		if err := d.conn.(*tls.Conn).Handshake(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR HANDSHAKE PASV: %s", err)
		}
	}
}

// Read implements the io.Reader interface as well as providing us
// with an early return for any accept errors
func (d *passiveDataConn) Read(p []byte) (int, error) {
	if err := d.ctx.Err(); err != nil {
		return 0, err
	}

	d.Lock()
	defer d.Unlock()

	if d.err != nil {
		return 0, d.err
	}

	n, err := d.conn.Read(p)
	d.read += n

	return n, err
}

// Write implements the io.Writer interface as well as providing us
// with an early return for any accept errors
func (d *passiveDataConn) Write(p []byte) (int, error) {
	if err := d.ctx.Err(); err != nil {
		return 0, err
	}

	d.Lock()
	defer d.Unlock()

	if d.err != nil {
		return 0, d.err
	}

	n, err := d.conn.Write(p)
	d.written += n

	return n, err
}

// isErrorAddressAlreadyInUse checks to see if this is a bind to port issue
func isErrorAddressAlreadyInUse(err error) bool {
	errOpError, ok := err.(*net.OpError)
	if !ok {
		return false
	}

	errSyscallError, ok := errOpError.Err.(*os.SyscallError)
	if !ok {
		return false
	}

	errErrno, ok := errSyscallError.Err.(syscall.Errno)
	if !ok {
		return false
	}

	if errErrno == syscall.EADDRINUSE {
		return true
	}

	const WSAEADDRINUSE = 10048
	if runtime.GOOS == "windows" && errErrno == WSAEADDRINUSE {
		return true
	}

	return false
}

func (d *passiveDataConn) Host() string      { return d.host }
func (d *passiveDataConn) Port() int         { return int(d.port) }
func (d *passiveDataConn) BytesRead() int    { return d.read }
func (d *passiveDataConn) BytesWritten() int { return d.written }
func (d *passiveDataConn) Kind() string      { return "Passive" }
