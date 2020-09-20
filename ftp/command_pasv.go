package ftp

import (
	"context"
	"fmt"
	"strings"
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

	return s.ReplyWithArgs(StatusPassiveMode, c.toString(s.data))
}

func (c commandPASV) toString(d Data) string {
	p1 := d.Port() / 256
	p2 := d.Port() - (p1 * 256)

	parts := strings.Split(d.Host(), ".")
	return fmt.Sprintf("(%s,%s,%s,%s,%d,%d)", parts[0], parts[1], parts[2], parts[3], p1, p2)
}

func init() {
	commandMap["PASV"] = &commandPASV{}
}
