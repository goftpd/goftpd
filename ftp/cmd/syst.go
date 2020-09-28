package cmd

import (
	"context"
)

/*
   This command is used to find out the type of operating
   system at the server.  The reply shall have as its first
   word one of the system names listed in the current version
   of the Assigned Numbers document [4].
*/

type commandSYST struct{}

func (c commandSYST) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandSYST) Execute(ctx context.Context, s Session, params []string) error {
	s.ReplyWithMessage(StatusSystemType, "UNIX Type: L8")
	return nil
}

func init() {
	CommandMap["SYST"] = &commandSYST{}
}
