package ftp

import "github.com/goftpd/goftpd/vfs"

type Command interface {
	Feat() string

	RequireParam() bool
	RequireState() SessionState

	Do(*Session, vfs.VFS, []string) error
}

var commandMap = map[string]Command{}
