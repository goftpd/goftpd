package ftp

import (
	"crypto/tls"
	"errors"
	"sync"

	"github.com/goftpd/goftpd/vfs"
)

// ServerOpts is used to create a new Server. PublicIP, TLSCertFile,
// TLSKeyFile are all required.
type ServerOpts struct {
	Name         string
	Host         string
	Port         int
	PassivePorts []int

	// required
	PublicIP    string
	TLSCertFile string
	TLSKeyFile  string
}

// Server. Serves stuff.
type Server struct {
	*ServerOpts

	tlsConfig *tls.Config

	fs vfs.VFS

	sessionPool sync.Pool
}

// Create a new Server using the supplied ServerOpts and VFS. Will
// fail if some required options are missing or it's unable to load
// the specified TLS cert/key files.
func NewServer(opts ServerOpts, fs vfs.VFS) (*Server, error) {
	// validation
	if len(opts.PublicIP) == 0 {
		return nil, errors.New("public_ip required")
	}

	// set defaults
	if len(opts.Name) == 0 {
		opts.Name = "Babylon"
	}

	if len(opts.Host) == 0 {
		opts.Host = "::"
	}

	if opts.Port == 0 {
		opts.Port = 2121
	}

	if len(opts.PassivePorts) != 2 {
		opts.PassivePorts = []int{
			20000,
			30000,
		}
	}

	// setup tlsConfig
	tlsConfig := &tls.Config{}

	tlsConfig.NextProtos = []string{"ftp"}

	cert, err := tls.LoadX509KeyPair(opts.TLSCertFile, opts.TLSKeyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig.Certificates = []tls.Certificate{cert}

	s := Server{
		ServerOpts: &opts,
		tlsConfig:  tlsConfig,
		fs:         fs,
		sessionPool: sync.Pool{
			New: func() interface{} {
				return &Session{}
			},
		},
	}

	return &s, nil
}
