package cmd

import (
	"context"
	"errors"
)

/*
   MAKE DIRECTORY (MKD)

      This command causes the directory specified in the pathname
      to be created as a directory (if the pathname is absolute)
      or as a subdirectory of the current working directory (if
      the pathname is relative).  See Appendix II.
*/

type commandMKD struct{}

func (c commandMKD) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandMKD) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) == 0 {
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	path := s.FS().Join(s.CWD(), params)

	user := s.User()
	if user == nil {
		return errors.New("no user found")
	}

	if err := s.FS().MakeDir(path, user); err != nil {
		s.ReplyError(StatusActionNotOK, err)
		return nil
	}

	s.ReplyWithArgs(StatusPathCreated, path)
	return nil
}

func init() {
	CommandMap["MKD"] = &commandMKD{}
}
