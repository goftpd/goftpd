package cmd

import (
	"context"
	"fmt"
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

func (c commandCWD) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) == 0 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	user, ok := s.User()
	if !ok {
		return s.ReplyStatus(StatusNotLoggedIn)
	}

	path := s.FS().Join(s.CWD(), params)

	// acl checks
	_, err := s.FS().ListDir(path, user)
	if err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	s.SetCWD(path)

	return s.ReplyWithMessage(StatusFileActionOK, fmt.Sprintf(`Current Working Dir "%s"`, path))
}

func init() {
	CommandMap["CWD"] = &commandCWD{}
}
