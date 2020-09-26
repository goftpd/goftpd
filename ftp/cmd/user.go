package cmd

import (
	"context"
)

/*
   USER NAME (USER)

      The argument field is a Telnet string identifying the user.
      The user identification is that which is required by the
      server for access to its file system.  This command will
      normally be the first command transmitted by the user after
      the control connections are made (some servers may require
      this).  Additional identification information in the form of
      a password and/or an account command may also be required by
      some servers.  Servers may allow a new USER command to be
      entered at any point in order to change the access control
      and/or accounting information.  This has the effect of
      flushing any user, password, and account information already
      supplied and beginning the login sequence again.  All
      transfer parameters are unchanged and any file transfer in
      progress is completed under the old access control
      parameters.
*/

type commandUSER struct{}

func (c commandUSER) RequireState() SessionState { return SessionStateAuth }

func (c commandUSER) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) != 1 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	s.SetLogin("")

	if err := s.ReplyStatus(StatusNeedPassword); err != nil {
		return err
	}

	s.SetLogin(params[0])

	return nil
}

func init() {
	CommandMap["USER"] = &commandUSER{}
}
