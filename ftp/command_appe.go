package ftp

import (
	"context"
	"fmt"
	"io"
)

/*
   APPEND (with create) (APPE)

      This command causes the server-DTP to accept the data
      transferred via the data connection and to store the data in
      a file at the server site.  If the file specified in the
      pathname exists at the server site, then the data shall be
      appended to that file; otherwise the file specified in the
      pathname shall be created at the server site.
*/

type commandAPPE struct{}

func (c commandAPPE) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandAPPE) Execute(ctx context.Context, s *Session, params []string) error {
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

	writer, err := s.server.fs.ResumeUploadFile(path, user)
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
	commandMap["APPE"] = &commandAPPE{}
}
