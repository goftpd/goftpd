package cmd

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

func (c commandPASV) Execute(ctx context.Context, s Session, params []string) error {

	// check if we have an existing data conncetion, if so cancel it
	if s.Data() != nil {
		if err := s.Data().Close(); err != nil {
			s.ReplyError(StatusCantOpenDataConnection, err)
			return nil
		}
	}

	// create new passive data connection
	if err := s.NewPassiveDataConn(ctx); err != nil {
		s.ReplyError(StatusCantOpenDataConnection, err)
		return nil
	}

	s.ReplyWithArgs(StatusPassiveMode, c.toString(s.Data()))
	return nil
}

func (c commandPASV) toString(d DataConn) string {
	p1 := d.Port() / 256
	p2 := d.Port() - (p1 * 256)

	parts := strings.Split(d.Host(), ".")
	return fmt.Sprintf("(%s,%s,%s,%s,%d,%d)", parts[0], parts[1], parts[2], parts[3], p1, p2)
}

func init() {
	CommandMap["PASV"] = &commandPASV{}
}
