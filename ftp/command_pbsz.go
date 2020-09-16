package ftp

import (
	"errors"
	"strconv"

	"github.com/goftpd/goftpd/vfs"
)

/*
   PROTECTION BUFFER SIZE (PBSZ)

      The argument is a decimal integer representing the maximum size,
      in bytes, of the encoded data blocks to be sent or received during
      file transfer.  This number shall be no greater than can be
      represented in a 32-bit unsigned integer.

      This command allows the FTP client and server to negotiate a
      maximum protected buffer size for the connection.  There is no
      default size; the client must issue a PBSZ command before it can
      issue the first PROT command.

      The PBSZ command must be preceded by a successful security data
      exchange.

      If the server cannot parse the argument, or if it will not fit in
      32 bits, it should respond with a 501 reply code.

      If the server has not completed a security data exchange with the
      client, it should respond with a 503 reply code.

      Otherwise, the server must reply with a 200 reply code.  If the
      size provided by the client is too large for the server, it must
      use a string of the form "PBSZ=number" in the text part of the
      reply to indicate a smaller buffer size.  The client and the
      server must use the smaller of the two buffer sizes if both buffer
      sizes are specified.
*/

type commandPBSZ struct{}

func (c commandPBSZ) Feat() string               { return "PBSZ" }
func (c commandPBSZ) RequireParam() bool         { return true }
func (c commandPBSZ) RequireState() SessionState { return SessionStateAuthenticated }

func (c commandPBSZ) Do(s *Session, fs vfs.VFS, params []string) error {
	if len(params) != 1 {
		s.Reply(501, "Syntax error in parameters or arguments.")
		return nil
	}

	if s.tlsConfig == nil {
		return errors.New("TLS Config is nil")
	}

	size, err := strconv.Atoi(params[0])
	if err != nil {
		s.Reply(501, "Syntax error in parameters or arguments.")
		return nil
	}

	// cant find any reference to what buffer this is

	s.Reply(200, "OK.")

	s.pbsz = &size

	return nil
}

func init() {
	commandMap["PBSZ"] = &commandPBSZ{}
}
