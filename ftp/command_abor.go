package ftp

import (
	"context"
)

/*
         ABORT (ABOR)

            This command tells the server to abort the previous FTP
            service command and any associated transfer of data.  The
            abort command may require "special action", as discussed in
            the Section on FTP Commands, to force recognition by the
            server.  No action is to be taken if the previous command
            has been completed (including data transfer).  The control
            connection is not to be closed by the server, but the data
            connection must be closed.

            There are two cases for the server upon receipt of this
            command: (1) the FTP service command was already completed,
            or (2) the FTP service command is still in progress.

			In the first case, the server closes the data connection
            (if it is open) and responds with a 226 reply, indicating
            that the abort command was successfully processed.

            In the second case, the server aborts the FTP service in
            progress and closes the data connection, returning a 426
            reply to indicate that the service request terminated
            abnormally.  The server then sends a 226 reply,
            indicating that the abort command was successfully
            processed.

*/

type commandABOR struct{}

func (c commandABOR) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandABOR) Execute(ctx context.Context, s *Session, params []string) error {

	// check if we have an existing data conncetion, if so cancel it
	if s.data != nil {
		if err := s.data.Close(); err != nil {
			return s.ReplyError(StatusCantOpenDataConnection, err)
		}

		if err := s.ReplyStatus(StatusDataCloseAborted); err != nil {
			return err
		}

		// TODO: might be a race condition here
		s.data = nil
	}

	return s.ReplyStatus(StatusDataClosedOK)
}

func init() {
	commandMap["ABOR"] = &commandABOR{}
}
