package cmd

import (
	"context"
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
		return s.ReplyStatus(StatusSyntaxError)
	}

	path := s.FS().Join(s.CWD(), params)

	user, ok := s.User()
	if !ok {
		return s.ReplyStatus(StatusNotLoggedIn)
	}

	if err := s.FS().MakeDir(path, user); err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	return s.ReplyWithArgs(StatusPathCreated, path)
}

func init() {
	CommandMap["MKD"] = &commandMKD{}
}
