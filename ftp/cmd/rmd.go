package cmd

import (
	"context"
	"errors"
)

/*
   REMOVE DIRECTORY (RMD)

      This command causes the directory specified in the pathname
      to be removed as a directory (if the pathname is absolute)
      or as a subdirectory of the current working directory (if
      the pathname is relative).  See Appendix II.
*/

type commandRMD struct{}

func (c commandRMD) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandRMD) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) == 0 {
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	path := s.FS().Join(s.CWD(), params)

	user := s.User()
	if user == nil {
		return errors.New("no user found")
	}

	if err := s.FS().DeleteDir(path, user); err != nil {
		s.ReplyError(StatusActionNotOK, err)
		return nil
	}

	s.ReplyStatus(StatusFileActionOK)
	return nil
}

func init() {
	CommandMap["RMD"] = &commandRMD{}
}
