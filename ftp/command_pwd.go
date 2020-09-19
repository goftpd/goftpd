package ftp

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

func (c commandPWD) Execute(ctx context.Context, s *Session, params []string) error {
	if len(params) > 0 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	return s.ReplyWithMessage(StatusPathCreated, fmt.Sprintf(`"%s" is current directory.`, s.currentDir))
}

func init() {
	commandMap["PWD"] = &commandPWD{}
}
