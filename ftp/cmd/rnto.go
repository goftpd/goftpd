package cmd

import (
	"context"
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
		return s.ReplyStatus(StatusSyntaxError)
	}

	if s.LastCommand() != "RNFR" {
		return s.ReplyStatus(StatusBadCommandSequence)
	}

	user, ok := s.User()
	if !ok {
		return s.ReplyStatus(StatusNotLoggedIn)
	}

	oldpath := s.FS().Join(s.CWD(), s.RenameFrom())
	newpath := s.FS().Join(s.CWD(), params)

	if err := s.FS().RenameFile(oldpath, newpath, user); err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	return s.ReplyStatus(StatusFileActionOK)
}

func init() {
	CommandMap["RNTO"] = &commandRNTO{}
}
