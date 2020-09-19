package ftp

import (
	"context"
	"fmt"
	"path"
	"strings"
)

/*
   CHANGE WORKING DIRECTORY (CWD)

      This command allows the user to work with a different
      directory or dataset for file storage or retrieval without
      altering his login or accounting information.  Transfer
      parameters are similarly unchanged.  The argument is a
      pathname specifying a directory or other system dependent
      file group designator.
*/

type commandCWD struct{}

func (c commandCWD) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandCWD) Execute(ctx context.Context, s *Session, params []string) error {
	if len(params) != 1 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	// clean?
	if !strings.HasPrefix(params[0], "/") {
		params[0] = path.Join(s.currentDir, params[0])
	}

	user, ok := s.User()
	if !ok {
		return s.ReplyStatus(StatusNotLoggedIn)
	}

	// acl checks
	_, err := s.server.fs.ListDir(params[0], user)
	if err != nil {
		return s.ReplyStatus(StatusActionNotOK)
	}

	s.currentDir = params[0]

	return s.ReplyWithMessage(StatusFileActionOK, fmt.Sprintf(`Current Working Dir "%s"`, params[0]))
}

func init() {
	commandMap["CWD"] = &commandCWD{}
}
