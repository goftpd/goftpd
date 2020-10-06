package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/goftpd/goftpd/acl"
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

func (c commandPASS) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) != 1 {
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	if len(s.Login()) == 0 {
		s.ReplyStatus(StatusBadCommandSequence)
		return nil
	}

	if !s.Auth().CheckPassword(s.Login(), params[0]) {
		s.SetLogin("")
		s.ReplyStatus(StatusNotLoggedIn)
		return nil
	}

	conn := s.Control()
	laddr := conn.LocalAddr()
	raddr := conn.RemoteAddr()

	if !s.Auth().CheckIP(s.Login(), laddr, raddr) {
		s.SetLogin("")
		s.ReplyStatus(StatusNotLoggedIn)
		return nil
	}

	user, err := s.Auth().GetUser(s.Login())
	if err != nil {
		s.SetLogin("")
		s.ReplyStatus(StatusNotLoggedIn)
		return nil
	}

	if !user.DeletedAt.IsZero() {
		s.SetLogin("")
		s.ReplyStatus(StatusNotLoggedIn)
		return nil
	}

	s.ReplyWithArgs(StatusUserLoggedIn, fmt.Sprintf("Welcome back %s!", s.Login()))

	go func() {
		s.Auth().UpdateUser(s.Login(), func(u *acl.User) error {
			u.LastLoginAt = time.Now()
			return nil
		})
	}()

	s.SetState(SessionStateLoggedIn)

	return nil
}

func init() {
	CommandMap["PASS"] = &commandPASS{}
}
