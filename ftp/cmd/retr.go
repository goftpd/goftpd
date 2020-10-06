package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/goftpd/goftpd/acl"
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
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	if s.Data() == nil {
		s.ReplyStatus(StatusCantOpenDataConnection)
		return nil
	}

	path := s.FS().Join(s.CWD(), params)

	user := s.User()
	if user == nil {
		return errors.New("no user found")
	}

	reader, size, err := s.FS().DownloadFile(path, user)
	if err != nil {
		s.ReplyError(StatusActionNotOK, err)
		return nil
	}

	var returnCredits bool
	if user.Ratio > 0 {
		if size > 1024 {
			size = size / 1024
			go func() {
				s.Auth().UpdateUser(user.Name, func(u *acl.User) error {
					u.Credits -= size
					return nil
				})
			}()
		}

		defer func() {
			if returnCredits {
				go func() {
					s.Auth().UpdateUser(user.Name, func(u *acl.User) error {
						u.Credits += size
						return nil
					})
				}()
			}
		}()
	}

	if s.DataProtected() {
		s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for download using TLS/SSL.")
		if err := s.Flush(); err != nil {
			returnCredits = true
			return err
		}
	} else {
		s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for download.")
		if err := s.Flush(); err != nil {
			returnCredits = true
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
			s.ReplyError(StatusActionNotOK, err)
			returnCredits = true
			return nil
		}
	}

	buf := s.FS().GetBuffer()
	defer s.FS().PutBuffer(buf)

	n, err := io.CopyBuffer(s.Data(), reader, *buf)
	if err != nil {
		s.ReplyError(StatusActionNotOK, err)
		returnCredits = true
		return nil
	}

	if err := s.Data().Close(); err != nil {
		s.ReplyError(StatusActionNotOK, err)
		// TODO
		// hard to say if we should return credits here
		returnCredits = true
		return nil
	}

	s.ReplyWithMessage(StatusDataClosedOK, fmt.Sprintf("OK, sent %d bytes.", n))
	return nil
}

func init() {
	CommandMap["RETR"] = &commandRETR{}
}
