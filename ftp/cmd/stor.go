package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/goftpd/goftpd/acl"
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

	user := s.User()
	if user == nil {
		return errors.New("no user found")
	}

	writer, err := s.FS().UploadFile(path, user)
	if err != nil {
		s.ReplyError(StatusActionNotOK, err)
		return nil
	}
	defer writer.Close()

	if s.DataProtected() {
		s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for upload using TLS/SSL.")
		if err := s.Flush(); err != nil {
			return err
		}
	} else {
		s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for upload.")
		if err := s.Flush(); err != nil {
			return err
		}
	}
	defer s.Data().Close()
	defer s.ClearData()

	buf := s.FS().GetBuffer()
	defer s.FS().PutBuffer(buf)

	n, err := io.CopyBuffer(writer, s.Data(), *buf)
	if err != nil {
		// delete the file
		if err := s.FS().DeleteFile(path, acl.SuperUser); err != nil {
			return err
		}

		s.ReplyError(StatusActionNotOK, err)
		return nil
	}

	s.Data().Close()

	if n == 0 {
		if err := s.FS().DeleteFile(path, acl.SuperUser); err != nil {
			return err
		}
		s.ReplyWithMessage(StatusActionNotOK, "0 bytes sent.")
		return nil
	}

	// TODO
	// do we want to store for ratio 0?
	if n > 1024 {
		n = n / 1024
		go func() {
			s.Auth().UpdateUser(user.Name, func(u *acl.User) error {
				u.Credits += n * int64(u.Ratio)
				return nil
			})
		}()
	}

	s.ReplyWithMessage(StatusDataClosedOK, fmt.Sprintf("OK, received %d bytes.", n))
	return nil
}

func init() {
	CommandMap["STOR"] = &commandSTOR{}
}
