package cmd

import (
	"context"
	"fmt"
	"io"
)

/*
   STORE (STOR)

      This command causes the server-DTP to accept the data
      transferred via the data connection and to store the data as
      a file at the server site.  If the file specified in the
      pathname exists at the server site, then its contents shall
      be replaced by the data being transferred.  A new file is
      created at the server site if the file specified in the
      pathname does not already exist.
*/

type commandSTOR struct{}

func (c commandSTOR) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandSTOR) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) == 0 {
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	if s.Data() == nil {
		s.ReplyStatus(StatusCantOpenDataConnection)
		return nil
	}

	path := s.FS().Join(s.CWD(), params)

	user, ok := s.User()
	if !ok {
		s.ReplyStatus(StatusNotLoggedIn)
		return nil
	}

	writer, err := s.FS().UploadFile(path, user)
	if err != nil {
		s.ReplyError(StatusActionNotOK, err)
		return nil
	}

	if s.DataProtected() {
		s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for upload using TLS/SSL.")
	} else {
		s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for upload.")
	}
	defer s.Data().Close()
	defer s.ClearData()

	n, err := io.Copy(writer, s.Data())
	if err != nil {
		s.ReplyError(StatusActionNotOK, err)
		return nil
	}

	s.Data().Close()

	s.ReplyWithMessage(StatusDataClosedOK, fmt.Sprintf("OK, received %d bytes.", n))
	return nil
}

func init() {
	CommandMap["STOR"] = &commandSTOR{}
}
