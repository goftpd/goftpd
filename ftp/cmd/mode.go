package cmd

import (
	"context"
	"strings"
)

/*

         TRANSFER MODE (MODE)

            The argument is a single Telnet character code specifying
            the data transfer modes described in the Section on
            Transmission Modes.

            The following codes are assigned for transfer modes:

               S - Stream
               B - Block
               C - Compressed

            The default transfer mode is Stream.

			 MODE is obsolete. The server should accept MODE S (in any combination of lowercase and uppercase) with code 200, and reject all other MODE attempts with code 504.
*/

type commandMODE struct{}

func (c commandMODE) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandMODE) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) != 1 {
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	if strings.ToUpper(params[0]) != "S" {
		s.ReplyStatus(StatusParameterNotImplemented)
		return nil
	}

	s.ReplyStatus(StatusOK)
	return nil
}

func init() {
	CommandMap["MODE"] = &commandMODE{}
}
