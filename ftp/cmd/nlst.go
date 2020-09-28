package cmd

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

func (c commandNLST) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) == 0 {
		params = append(params, s.CWD())
	}

	if s.Data() == nil {
		s.ReplyStatus(StatusCantOpenDataConnection)
		return nil
	}

	user, ok := s.User()
	if !ok {
		s.ReplyStatus(StatusNotLoggedIn)
		return nil
	}

	var options, path string

	// check if we have options and set the path eitherway
	if len(params[0]) > 0 && params[0][0] == '-' {
		options = params[0]
		if len(params) > 1 {
			params = params[1:]
		} else {
			params = []string{}
		}
	}

	path = s.FS().Join(s.CWD(), params)

	// get file list and parse with any options
	finfo, err := s.FS().ListDir(path, user)
	if err != nil {
		s.ReplyError(StatusActionAbortedError, err)
		return nil
	}

	if len(options) > 0 {
		// handle options
		finfo.SortByName()
	}

	if s.DataProtected() {
		s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for directory listing using TLS/SSL.")
	} else {
		s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for directory listing.")
	}
	defer s.Data().Close()
	defer s.ClearData()

	// write it
	n, err := s.Data().Write(finfo.Short())
	if err != nil {
		s.ReplyError(StatusActionAbortedError, err)
		return nil
	}

	s.Data().Close()

	s.ReplyWithMessage(StatusDataClosedOK, fmt.Sprintf("Closing data connection, sent %d bytes", n))

	return nil
}

func init() {
	CommandMap["NLST"] = &commandNLST{}
}
