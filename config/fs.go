package config

import (
	"regexp"

	"github.com/dgraph-io/badger/v2"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/goftpd/goftpd/vfs"
	"github.com/pkg/errors"
)

func (c *Config) ParseFS() (vfs.VFS, error) {
	var opts vfs.FilesystemOpts

	lines, ok := c.lines[NamespaceFS]
	if !ok {
		return nil, errors.New("no fs options provided")
	}

	if err := c.parse(lines, &opts); err != nil {
		return nil, err
	}

	if len(opts.Root) == 0 {
		return nil, errors.New("must specify `fs rootpath`")
	}

	if len(opts.ShadowDB) == 0 {
		opts.ShadowDB = "site/config/shadow.db"
	}

	if len(opts.Hide) > 0 {
		re, err := regexp.Compile(opts.Hide)
		if err != nil {
			return nil, errors.WithMessage(err, `"fs hide" regexp is bad`)
		}
		opts.SetHideRE(re)
	}

	ufs := osfs.New(opts.Root)

	opt := badger.DefaultOptions(opts.ShadowDB)
	// disable badger logger
	opt.Logger = nil

	db, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}

	shadowFS := vfs.NewShadowStore(db)

	perms, err := c.ParsePermissions()
	if err != nil {
		return nil, err
	}

	fs, err := vfs.NewFilesystem(&opts, ufs, shadowFS, perms)
	if err != nil {
		return nil, err
	}

	return fs, nil
}
