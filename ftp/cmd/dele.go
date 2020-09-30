package cmd

import (
	"context"
	"errors"
)

/*
   DELETE (DELE)

      This command causes the file specified in the pathname to be
      deleted at the server site.  If an extra level of protection
      is desired (such as the query, "Do you really wish to
      delete?"), it should be provided by the user-FTP process.
*/

type commandDELE struct{}

func (c commandDELE) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandDELE) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) == 0 {
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	path := s.FS().Join(s.CWD(), params)

	user := s.User()
	if user == nil {
		return errors.New("no user found")
	}

	if err := s.FS().DeleteFile(path, user); err != nil {
		s.ReplyError(StatusActionNotOK, err)
		return nil
	}

	s.ReplyStatus(StatusFileActionOK)
	return nil
}

func init() {
	CommandMap["DELE"] = &commandDELE{}
}
