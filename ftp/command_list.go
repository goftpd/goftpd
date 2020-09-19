package ftp

import (
	"context"
	"fmt"
	"log"
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

func (c commandLIST) Execute(ctx context.Context, s *Session, params []string) error {
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

	finfo, err := s.server.fs.ListDir(params[0], user)
	if err != nil {
		log.Printf("ERROR ListDir: %s: %s", params[0], err)
		return s.ReplyStatus(StatusActionAbortedError)
	}

	// write it
	n, err := s.data.Write(finfo.Detailed())
	if err != nil {
		log.Printf("ERROR Write: %s", err)
		return s.ReplyStatus(StatusActionAbortedError)
	}

	s.ReplyWithMessage(StatusDataClosedOK, fmt.Sprintf("Closing data connection, sent %d bytes", n))

	return nil
}

func init() {
	commandMap["LIST"] = &commandLIST{}
}
