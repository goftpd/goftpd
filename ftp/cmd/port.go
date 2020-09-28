package cmd

import (
	"context"
	"fmt"
)

/*
   DATA PORT (PORT)

      The argument is a HOST-PORT specification for the data port
      to be used in data connection.  There are defaults for both
      the user and server data ports, and under normal
      circumstances this command and its reply are not needed.  If
      this command is used, the argument is the concatenation of a
      32-bit internet host address and a 16-bit TCP port address.
      This address information is broken into 8-bit fields and the
      value of each field is transmitted as a decimal number (in
      character string representation).  The fields are separated
      by commas.  A port command would be:

         PORT h1,h2,h3,h4,p1,p2

      where h1 is the high order 8 bits of the internet host
      address.
*/

type commandPORT struct{}

func (c commandPORT) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandPORT) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) != 1 {
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	// check if we have an existing data conncetion, if so cancel it
	if s.Data() != nil {
		if err := s.Data().Close(); err != nil {
			s.ReplyError(StatusCantOpenDataConnection, err)
			return nil
		}
	}

	// create new passive data connection
	if err := s.NewActiveDataConn(ctx, params[0]); err != nil {
		s.ReplyError(StatusCantOpenDataConnection, err)
		return nil
	}

	s.ReplyWithMessage(StatusOK, fmt.Sprintf("Connection established to (%s)", params[0]))
	return nil
}

func init() {
	CommandMap["PORT"] = &commandPORT{}
}
