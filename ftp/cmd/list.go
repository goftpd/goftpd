package cmd

import (
	"context"
	"fmt"
)

/*
         LIST (LIST)

            This command causes a list to be sent from the server to the
            passive DTP.  If the pathname specifies a directory or other
            group of files, the server should transfer a list of files
            in the specified directory.  If the pathname specifies a
            file then the server should send current information on the
            file.  A null argument implies the user's current working or
            default directory.  The data transfer is over the data
            connection in type ASCII or type EBCDIC.  (The user must
			ensure that the TYPE is appropriately ASCII or EBCDIC).
            Since the information on a file may vary widely from system
            to system, this information may be hard to use automatically
            in a program, but may be quite useful to a human user.
*/

type commandLIST struct{}

func (c commandLIST) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandLIST) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) == 0 {
		params = append(params, s.CWD())
	}

	if s.Data() == nil {
		return s.ReplyStatus(StatusCantOpenDataConnection)
	}

	user, ok := s.User()
	if !ok {
		return s.ReplyStatus(StatusNotLoggedIn)
	}

	var options, path string

	// check if we have options and set the path eitherway
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
		return s.ReplyError(StatusActionAbortedError, err)
	}

	if len(options) > 0 {
		// handle options
		finfo.SortByName()
	}

	if s.DataProtected() {
		if err := s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for directory listing using TLS/SSL."); err != nil {
			return err
		}
	} else {
		if err := s.ReplyWithMessage(StatusTransferStatusOK, "Opening connection for directory listing."); err != nil {
			return err
		}
	}
	defer s.Data().Close()
	defer s.ClearData()

	// write it
	n, err := s.Data().Write(finfo.Detailed())
	if err != nil {
		return s.ReplyError(StatusActionAbortedError, err)
	}

	s.Data().Close()

	s.ReplyWithMessage(StatusDataClosedOK, fmt.Sprintf("Closing data connection, sent %d bytes", n))

	return nil
}

func init() {
	CommandMap["LIST"] = &commandLIST{}
}
