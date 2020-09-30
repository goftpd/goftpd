package cmd

import (
	"context"
	"errors"
	"fmt"
)

/*
   CHANGE TO PARENT DIRECTORY (CDUP)

      This command is a special case of CWD, and is included to
      simplify the implementation of programs for transferring
      directory trees between operating systems having different
      syntaxes for naming the parent directory.  The reply codes
      shall be identical to the reply codes of CWD.  See
      Appendix II for further details.
*/

type commandCDUP struct{}

func (c commandCDUP) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandCDUP) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) != 0 {
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	user := s.User()
	if user == nil {
		return errors.New("no user found")
	}

	path := s.FS().Join(s.CWD(), []string{"../"})

	// acl checks
	_, err := s.FS().ListDir(path, user)
	if err != nil {
		s.ReplyError(StatusActionNotOK, err)
		return nil
	}

	s.SetCWD(path)

	s.ReplyWithMessage(StatusFileActionOK, fmt.Sprintf(`Current Working Dir "%s"`, path))
	return nil
}

func init() {
	CommandMap["CDUP"] = &commandCDUP{}
}
