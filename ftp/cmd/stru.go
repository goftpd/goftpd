package cmd

import (
	"context"
	"strings"
)

/*
         FILE STRUCTURE (STRU)

            The argument is a single Telnet character code specifying
            file structure described in the Section on Data
            Representation and Storage.

            The following codes are assigned for structure:

               F - File (no record structure)
               R - Record structure
               P - Page structure

            The default structure is File.

			 STRU is obsolete. The server should accept STRU F (in any combination of lowercase and uppercase) with code 200, and reject all other STRU attempts with code 504.
*/

type commandSTRU struct{}

func (c commandSTRU) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandSTRU) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) != 1 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	if strings.ToUpper(params[0]) != "F" {
		return s.ReplyStatus(StatusParameterNotImplemented)
	}

	return s.ReplyStatus(StatusOK)
}

func init() {
	CommandMap["STRU"] = &commandSTRU{}
}
