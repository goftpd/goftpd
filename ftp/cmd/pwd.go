package cmd

import (
	"context"
	"fmt"
)

/*
   PRINT WORKING DIRECTORY (PWD)

      This command causes the name of the current working
      directory to be returned in the reply.  See Appendix II.
*/

type commandPWD struct{}

func (c commandPWD) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandPWD) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) > 0 {
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	s.ReplyWithMessage(StatusPathCreated, fmt.Sprintf(`"%s" is current directory.`, s.CWD()))
	return nil
}

func init() {
	CommandMap["PWD"] = &commandPWD{}
}
