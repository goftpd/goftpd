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
		return s.ReplyStatus(StatusSyntaxError)
	}

	if s.Data() == nil {
		return s.ReplyStatus(StatusCantOpenDataConnection)
	}

	path := s.FS().Join(s.CWD(), params)

	user, ok := s.User()
	if !ok {
		return s.ReplyStatus(StatusNotLoggedIn)
	}

	writer, err := s.FS().UploadFile(path, user)
	if err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	if s.DataProtected() {
		if err := s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for upload using TLS/SSL."); err != nil {
			return err
		}
	} else {
		if err := s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for upload."); err != nil {
			return err
		}
	}
	defer s.Data().Close()
	defer s.ClearData()

	n, err := io.Copy(writer, s.Data())
	if err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	s.Data().Close()

	return s.ReplyWithMessage(StatusDataClosedOK, fmt.Sprintf("OK, received %d bytes.", n))
}

func init() {
	CommandMap["STOR"] = &commandSTOR{}
}
