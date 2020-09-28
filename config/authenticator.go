package config

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/goftpd/goftpd/acl"
	"github.com/pkg/errors"
)

func (c *Config) ParseAuthenticator() (acl.Authenticator, error) {
	var opts acl.AuthenticatorOpts

	lines, ok := c.lines[NamespaceAuth]
	if !ok {
		return nil, errors.New("no auth options provided")
	}

	if err := c.parse(lines, &opts); err != nil {
		return nil, err
	}

	if len(opts.DB) == 0 {
		opts.DB = "site/config/users.db"
	}

	opt := badger.DefaultOptions(opts.DB)
	// disable badger logger
	opt.Logger = nil

	db, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}

	auth := acl.NewBadgerAuthenticator(db)

	return auth, nil
}
