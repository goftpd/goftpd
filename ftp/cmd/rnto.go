package cmd

import (
	"context"
	"errors"
)

/*
	RENAME TO (RNTO)

   	This command specifies the new pathname of the file
   	specified in the immediately preceding "rename from"
   	command.  Together the two commands cause a file to be
   	renamed.
*/

type commandRNTO struct{}

func (c commandRNTO) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandRNTO) Execute(ctx context.Context, s Session, params []string) error {
	defer s.SetRenameFrom(nil)

	if len(params) == 0 {
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	if s.LastCommand() != "RNFR" {
		s.ReplyStatus(StatusBadCommandSequence)
		return nil
	}

	user := s.User()
	if user == nil {
		return errors.New("no user found")
	}

	oldpath := s.FS().Join(s.CWD(), s.RenameFrom())
	newpath := s.FS().Join(s.CWD(), params)

	if err := s.FS().RenameFile(oldpath, newpath, user); err != nil {
		s.ReplyError(StatusActionNotOK, err)
		return nil
	}

	s.ReplyStatus(StatusFileActionOK)
	return nil
}

func init() {
	CommandMap["RNTO"] = &commandRNTO{}
}
