package ftp

import (
	"context"
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

func (c commandRMD) Execute(ctx context.Context, s *Session, params []string) error {
	if len(params) == 0 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	path := s.server.fs.Join(s.currentDir, params)

	user, ok := s.User()
	if !ok {
		return s.ReplyStatus(StatusNotLoggedIn)
	}

	if err := s.server.fs.DeleteDir(path, user); err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	return s.ReplyStatus(StatusFileActionOK)
}

func init() {
	commandMap["RMD"] = &commandRMD{}
}
