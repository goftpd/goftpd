package ftp

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

func (c commandSTOR) Execute(ctx context.Context, s *Session, params []string) error {
	if len(params) == 0 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	if s.data == nil {
		return s.ReplyStatus(StatusCantOpenDataConnection)
	}

	if s.dataProtected {
		if err := s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for upload using TLS/SSL."); err != nil {
			return err
		}
	} else {
		if err := s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for upload."); err != nil {
			return err
		}
	}
	defer s.data.Close()
	defer func() {
		s.data = nil
	}()

	path := s.server.fs.Join(s.currentDir, params)

	user, ok := s.User()
	if !ok {
		return s.ReplyStatus(StatusNotLoggedIn)
	}

	writer, err := s.server.fs.UploadFile(path, user)
	if err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	n, err := io.Copy(writer, s.data)
	if err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	return s.ReplyWithMessage(StatusDataClosedOK, fmt.Sprintf("OK, received %d bytes.", n))
}

func init() {
	commandMap["STOR"] = &commandSTOR{}
}
