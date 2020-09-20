package ftp

import (
	"context"
)

/*
   PASSIVE (PASV)

      This command requests the server-DTP to "listen" on a data
      port (which is not its default data port) and to wait for a
      connection rather than initiate one upon receipt of a
      transfer command.  The response to this command includes the
      host and port address this server is listening on.
*/

type commandPASV struct{}

func (c commandPASV) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandPASV) Execute(ctx context.Context, s *Session, params []string) error {

	// check if we have an existing data conncetion, if so cancel it
	if s.data != nil {
		if err := s.data.Close(); err != nil {
			return s.ReplyError(StatusCantOpenDataConnection, err)
		}
	}

	// create new passive data connection
	data, err := s.server.newPassiveDataConn(ctx, s.dataProtected)
	if err != nil {
		return s.ReplyError(StatusCantOpenDataConnection, err)
	}

	s.data = data

	return s.ReplyWithArgs(StatusPassiveMode, s.data.String())
}

func init() {
	commandMap["PASV"] = &commandPASV{}
}
