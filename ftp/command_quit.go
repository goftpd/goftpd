package ftp

import (
	"context"
)

/*
   LOGOUT (QUIT)

      This command terminates a USER and if file transfer is not
      in progress, the server closes the control connection.  If
      file transfer is in progress, the connection will remain
      open for result response and the server will then close it.
      If the user-process is transferring files for several USERs
      but does not wish to close and then reopen connections for
      each, then the REIN command should be used instead of QUIT.

      An unexpected close on the control connection will cause the
      server to take the effective action of an abort (ABOR) and a
      logout (QUIT).
*/

type commandQUIT struct{}

func (c commandQUIT) Feat() string               { return "QUIT" }
func (c commandQUIT) RequireParam() bool         { return true }
func (c commandQUIT) RequireState() SessionState { return SessionStateAuth }

func (c commandQUIT) Execute(ctx context.Context, s *Session, params []string) error {
	_, ok := s.User()
	if !ok {
		return s.ReplyStatus(StatusNotLoggedIn)
	}

	if s.data != nil {
		if err := s.data.Close(); err != nil {
			return s.ReplyError(StatusCommandUnrecognised, err)
		}
	}

	defer s.Close()

	return s.ReplyStatus(StatusClosingControl)
}

func init() {
	commandMap["QUIT"] = &commandQUIT{}
}
