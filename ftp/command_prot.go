package ftp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goftpd/goftpd/vfs"
)

/*
   DATA CHANNEL PROTECTION LEVEL (PROT)

      The argument is a single Telnet character code specifying the data
      channel protection level.

      This command indicates to the server what type of data channel
      protection the client and server will be using.  The following
      codes are assigned:

         C - Clear
         S - Safe
         E - Confidential
         P - Private

	  The default protection level if no other level is specified is
      Clear.  The Clear protection level indicates that the data channel
      will carry the raw data of the file transfer, with no security
      applied.  The Safe protection level indicates that the data will
      be integrity protected.  The Confidential protection level
      indicates that the data will be confidentiality protected.  The
      Private protection level indicates that the data will be integrity
      and confidentiality protected.

      It is reasonable for a security mechanism not to provide all data
      channel protection levels.  It is also reasonable for a mechanism
      to provide more protection at a level than is required (for
      instance, a mechanism might provide Confidential protection, but
      include integrity-protection in that encoding, due to API or other
      considerations).

      The PROT command must be preceded by a successful protection
      buffer size negotiation.

      If the server does not understand the specified protection level,
      it should respond with reply code 504.

      If the current security mechanism does not support the specified
      protection level, the server should respond with reply code 536.

      If the server has not completed a protection buffer size
      negotiation with the client, it should respond with a 503 reply
      code.

      The PROT command will be rejected and the server should reply 503
      if no previous PBSZ command was issued.

      If the server is not willing to accept the specified protection
      level, it should respond with reply code 534.

      If the server is not able to accept the specified protection
      level, such as if a required resource is unavailable, it should
      respond with reply code 431.

      Otherwise, the server must reply with a 200 reply code to indicate
      that the specified protection level is accepted.
*/

type commandPROT struct{}

func (c commandPROT) Feat() string               { return "PROT" }
func (c commandPROT) RequireParam() bool         { return true }
func (c commandPROT) RequireState() SessionState { return SessionStateAuthenticated }

func (c commandPROT) Do(s *Session, fs vfs.VFS, params []string) error {
	if len(params) != 1 {
		s.Reply(501, "Syntax error in parameters or arguments.")
		return nil
	}

	if s.tlsConfig == nil {
		return errors.New("TLS Config is nil")
	}

	if s.pbsz == nil {
		s.Reply(503, "Send PBSZ first.")
		return nil
	}

	switch strings.ToUpper(params[0]) {
	case "P":
		s.prot = &params[0]

	case "C":
	case "S":
	case "E":
		s.Reply(534, "Only Protection Level P is supported. Please use secure data transfer.")
		return nil

	default:
		s.Reply(504, "Unknown Protection Level.")
		return nil
	}

	s.Reply(200, fmt.Sprintf("Protection Level '%s' accepted.", params[0]))

	s.state = SessionStateAuthenticated

	return nil
}

func init() {
	commandMap["PROT"] = &commandPROT{}
}
