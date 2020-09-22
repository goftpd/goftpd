package config

import (
	"crypto/tls"

	"github.com/goftpd/goftpd/ftp"
	"github.com/pkg/errors"
)

func (c *Config) ParseServerOpts() (*ftp.ServerOpts, error) {
	var opts ftp.ServerOpts

	lines, ok := c.lines[NamespaceServer]
	if !ok {
		return nil, errors.New("no server options provided")
	}

	if err := c.parse(lines, &opts); err != nil {
		return nil, err
	}

	// validation
	if len(opts.PublicIP) == 0 {
		return nil, errors.New("public_ip required")
	}

	// set defaults
	if len(opts.Name) == 0 {
		opts.Name = "go"
	}

	if len(opts.LongName) == 0 {
		opts.Name = "goftpd"
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

	if opts.PassivePorts[0] >= opts.PassivePorts[1] {
		return nil, errors.New("Passvive Ports must be in order: min,max")
	}

	// setup tlsConfig
	tlsConfig := &tls.Config{}

	tlsConfig.NextProtos = []string{"ftp"}

	cert, err := tls.LoadX509KeyPair(opts.TLSCertFile, opts.TLSKeyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig.Certificates = []tls.Certificate{cert}

	opts.SetTLSConfig(tlsConfig)

	return &opts, nil

}
