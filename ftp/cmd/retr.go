package cmd

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

func (c commandRETR) Execute(ctx context.Context, s Session, params []string) error {
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

	reader, err := s.FS().DownloadFile(path, user)
	if err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	if s.DataProtected() {
		if err := s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for download using TLS/SSL."); err != nil {
			return err
		}
	} else {
		if err := s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for download."); err != nil {
			return err
		}
	}
	defer s.Data().Close()
	defer s.ClearData()

	// reset seek
	defer s.SetRestartPosition(0)

	// seek reader
	if s.RestartPosition() > 0 {
		if _, err := reader.Seek(int64(s.RestartPosition()), io.SeekStart); err != nil {
			return s.ReplyError(StatusActionNotOK, err)
		}
	}

	n, err := io.Copy(s.Data(), reader)
	if err != nil {
		return s.ReplyError(StatusActionNotOK, err)
	}

	s.Data().Close()

	return s.ReplyWithMessage(StatusDataClosedOK, fmt.Sprintf("OK, received %d bytes.", n))
}

func init() {
	CommandMap["RETR"] = &commandRETR{}
}
