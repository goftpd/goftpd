package ftp

import "github.com/goftpd/goftpd/vfs"

type Command interface {
	Feat() string
	RequireParam() bool
	RequireAuth() bool
	Do(*Session, vfs.VFS, []string) error
}

var commandMap = map[string]Command{}
