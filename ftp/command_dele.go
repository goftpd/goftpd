package ftp

import (
	"context"
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

func (c commandDELE) Execute(ctx context.Context, s *Session, params []string) error {
	if len(params) == 0 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	path := s.server.fs.Join(s.currentDir, params)

	user, ok := s.User()
	if !ok {
		return s.ReplyStatus(StatusNotLoggedIn)
	}

	if err := s.server.fs.DeleteFile(path, user); err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	return s.ReplyStatus(StatusFileActionOK)
}

func init() {
	commandMap["DELE"] = &commandDELE{}
}
