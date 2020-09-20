package ftp

import (
	"context"
	"fmt"
)

/*
   PASSWORD (PASS)

      The argument field is a Telnet string specifying the user's
      password.  This command must be immediately preceded by the
      user name command, and, for some sites, completes the user's
      identification for access control.  Since password
      information is quite sensitive, it is desirable in general
      to "mask" it or suppress typeout.  It appears that the
      server has no foolproof way to achieve this.  It is
      therefore the responsibility of the user-FTP process to hide
      the sensitive password information.
*/

type commandPASS struct{}

func (c commandPASS) Feat() string               { return "PASS" }
func (c commandPASS) RequireParam() bool         { return true }
func (c commandPASS) RequireState() SessionState { return SessionStateAuth }

func (c commandPASS) Execute(ctx context.Context, s *Session, params []string) error {
	if len(params) != 1 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	if len(s.loginUser) == 0 {
		return s.ReplyStatus(StatusBadCommandSequence)
	}

	if err := s.ReplyWithArgs(StatusUserLoggedIn, fmt.Sprintf("Welcome back %s!", s.loginUser)); err != nil {
		return err
	}

	s.state = SessionStateLoggedIn

	return nil
}

func init() {
	commandMap["PASS"] = &commandPASS{}
}
