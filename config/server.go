package config

import (
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

	return &opts, nil

}
