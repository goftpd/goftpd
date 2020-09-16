package ftp

import (
	"github.com/goftpd/goftpd/vfs"
)

/*
 */

type commandUSER struct{}

func (c commandUSER) Feat() string               { return "USER" }
func (c commandUSER) RequireParam() bool         { return true }
func (c commandUSER) RequireState() SessionState { return SessionStateAuthenticated }

func (c commandUSER) Do(s *Session, fs vfs.VFS, params []string) error {
	if len(params) != 1 {
		s.Reply(501, "Syntax error in parameters or arguments.")
		return nil
	}

	s.Reply(334, "User name ok, password required.")

	s.username = &params[0]

	return nil
}

func init() {
	commandMap["USER"] = &commandUSER{}
}
