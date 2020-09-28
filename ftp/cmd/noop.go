package cmd

import (
	"context"
)

/*

   NOOP (NOOP)

      This command does not affect any parameters or previously
      entered commands. It specifies no action other than that the
      server send an OK reply.
*/

type commandNOOP struct{}

func (c commandNOOP) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandNOOP) Execute(ctx context.Context, s Session, params []string) error {
	s.ReplyStatus(StatusOK)
	return nil
}

func init() {
	CommandMap["NOOP"] = &commandNOOP{}
}
