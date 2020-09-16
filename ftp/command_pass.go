package ftp

import (
	"fmt"

	"github.com/goftpd/goftpd/vfs"
)

/*
 */

type commandPASS struct{}

func (c commandPASS) Feat() string               { return "PASS" }
func (c commandPASS) RequireParam() bool         { return true }
func (c commandPASS) RequireState() SessionState { return SessionStateAuthenticated }

func (c commandPASS) Do(s *Session, fs vfs.VFS, params []string) error {
	if len(params) != 1 {
		s.Reply(501, "Syntax error in parameters or arguments.")
		return nil
	}

	if s.username == nil {
		s.Reply(503, "Bad sequence of commands.")
		return nil
	}

	s.Reply(230, fmt.Sprintf("Welcome back %s!", *s.username))

	s.state = SessionStateLoggedIn

	return nil
}

func init() {
	commandMap["PASS"] = &commandPASS{}
}
