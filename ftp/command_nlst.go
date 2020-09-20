package ftp

import (
	"context"
	"fmt"
)

/*
   NAME LIST (NLST)

      This command causes a directory listing to be sent from
      server to user site.  The pathname should specify a
      directory or other system-specific file group descriptor; a
      null argument implies the current directory.  The server
      will return a stream of names of files and no other
      information.  The data will be transferred in ASCII or
      EBCDIC type over the data connection as valid pathname
      strings separated by <CRLF> or <NL>.  (Again the user must
      ensure that the TYPE is correct.)  This command is intended
      to return information that can be used by a program to
      further process the files automatically.  For example, in
      the implementation of a "multiple get" function.
*/

type commandNLST struct{}

func (c commandNLST) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandNLST) Execute(ctx context.Context, s *Session, params []string) error {
	if len(params) == 0 {
		params = append(params, s.currentDir)
	}

	if s.data == nil {
		return s.ReplyStatus(StatusCantOpenDataConnection)
	}

	if s.dataProtected {
		if err := s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for directory listing using TLS/SSL."); err != nil {
			return err
		}
	} else {
		if err := s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for directory listing."); err != nil {
			return err
		}
	}
	defer s.data.Close()
	defer func() {
		s.data = nil
	}()

	user, ok := s.User()
	if !ok {
		return s.ReplyStatus(StatusNotLoggedIn)
	}

	var options, path string

	// check if we have options and set the path eitherway
	if len(params[0]) > 0 && params[0][0] == '-' {
		options = params[0]
		params[0] = ""
	}

	path = s.server.fs.Join(s.currentDir, params)

	// get file list and parse with any options
	finfo, err := s.server.fs.ListDir(path, user)
	if err != nil {
		return s.ReplyError(StatusActionAbortedError, err)
	}

	if len(options) > 0 {
		// handle options
		finfo.SortByName()
	}

	// write it
	n, err := s.data.Write(finfo.Short())
	if err != nil {
		return s.ReplyError(StatusActionAbortedError, err)
	}

	s.ReplyWithMessage(StatusDataClosedOK, fmt.Sprintf("Closing data connection, sent %d bytes", n))

	return nil
}

func init() {
	commandMap["NLST"] = &commandNLST{}
}
