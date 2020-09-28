package cmd

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

func (c commandAPPE) Execute(ctx context.Context, s Session, params []string) error {
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

	if s.DataProtected() {
		s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for upload using TLS/SSL.")
	} else {
		s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for upload.")
	}
	defer s.Data().Close()
	defer s.ClearData()

	writer, err := s.FS().ResumeUploadFile(path, user)
	if err != nil {
		s.ReplyError(StatusActionNotOK, err)
		return nil
	}

	n, err := io.Copy(writer, s.Data())
	if err != nil {
		s.ReplyError(StatusActionNotOK, err)
		return nil
	}

	s.ClearData()

	s.ReplyWithMessage(StatusDataClosedOK, fmt.Sprintf("OK, received %d bytes.", n))
	return nil
}

func init() {
	CommandMap["APPE"] = &commandAPPE{}
}
