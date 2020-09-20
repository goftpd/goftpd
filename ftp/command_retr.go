package ftp

import (
	"context"
	"fmt"
	"io"
)

/*

   RETRIEVE (RETR)

      This command causes the server-DTP to transfer a copy of the
      file, specified in the pathname, to the server- or user-DTP
      at the other end of the data connection.  The status and
      contents of the file at the server site shall be unaffected.
*/

type commandRETR struct{}

func (c commandRETR) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandRETR) Execute(ctx context.Context, s *Session, params []string) error {
	if len(params) == 0 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	if s.data == nil {
		return s.ReplyStatus(StatusCantOpenDataConnection)
	}

	if s.dataProtected {
		if err := s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for download using TLS/SSL."); err != nil {
			return err
		}
	} else {
		if err := s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for download."); err != nil {
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

	reader, err := s.server.fs.DownloadFile(path, user)
	if err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	// reset seek
	defer func() {
		s.restartPosition = 0
	}()

	// seek reader
	if s.restartPosition > 0 {
		if _, err := reader.Seek(int64(s.restartPosition), io.SeekStart); err != nil {
			return s.ReplyError(StatusActionNotOK, err)
		}
	}

	n, err := io.Copy(s.data, reader)
	if err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	return s.ReplyWithMessage(StatusDataClosedOK, fmt.Sprintf("OK, received %d bytes.", n))
}

func init() {
	commandMap["RETR"] = &commandRETR{}
}
