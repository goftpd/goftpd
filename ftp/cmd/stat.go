package cmd

import (
	"context"
	"fmt"
)

/*
         STATUS (STAT)

            This command shall cause a status response to be sent over
            the control connection in the form of a reply.  The command
            may be sent during a file transfer (along with the Telnet IP
            and Synch signals--see the Section on FTP Commands) in which
            case the server will respond with the status of the
            operation in progress, or it may be sent between file
            transfers.  In the latter case, the command may have an
            argument field.  If the argument is a pathname, the command
            is analogous to the "list" command except that data shall be
			transferred over the control connection.  If a partial
            pathname is given, the server may respond with a list of
            file names or attributes associated with that specification.
            If no argument is given, the server should return general
            status information about the server FTP process.  This
            should include current values of all transfer parameters and
            the status of connections.
*/

const statBaseMessage string = `FTP server status:
Logged in as %s
TYPE: %s, FORM: Nonprint; STRUcture: File; transfer MODE: Stream; Protection: %s
%s
`

type commandSTAT struct{}

func (c commandSTAT) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandSTAT) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) == 0 {
		return c.NoParams(s)
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

	s.ReplyWithMessage(
		StatusSystemStatus,
		fmt.Sprintf(
			"Status of \"%s\":\n%s",
			path,
			finfo.Detailed(),
		),
	)
	return nil
}

func (c commandSTAT) NoParams(s Session) error {
	data := s.Data()

	// check if we have an existing data conncetion, if so cancel it
	dataMessage := "No data connection"
	if data != nil {
		dataMessage = fmt.Sprintf(
			"%s Data Connection. Written %d bytes Read %d bytes.",
			data.Kind(),
			data.BytesWritten(),
			data.BytesRead(),
		)
	}

	user, ok := s.User()
	if !ok {
		s.ReplyStatus(StatusNotLoggedIn)
		return nil
	}

	dataType := "ASCII"
	if s.BinaryMode() {
		dataType = "Binary"
	}

	protection := "Clear"
	if s.DataProtected() {
		protection = "Protected"
	}

	msg := fmt.Sprintf(statBaseMessage, user.Name, dataType, protection, dataMessage)

	s.ReplyWithMessage(StatusSystemStatus, msg)
	return nil
}

func init() {
	CommandMap["STAT"] = &commandSTAT{}
}
