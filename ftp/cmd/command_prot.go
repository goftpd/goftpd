package cmd

import (
	"context"
	"fmt"
	"strings"
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

func (c commandPROT) RequireState() SessionState { return SessionStateAuth }

func (c commandPROT) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) != 1 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	switch strings.ToUpper(params[0]) {
	case "P":
		s.SetDataProtected(true)

	case "C":
		s.SetDataProtected(false)

	case "S":
		fallthrough
	case "E":
		return s.ReplyWithArgs(StatusBadProtectionLevel, params[0])

	default:
		return s.ReplyStatus(StatusParameterNotImplemented)
	}

	return s.ReplyWithMessage(StatusOK, fmt.Sprintf("Protection Level '%s' accepted.", params[0]))
}

func init() {
	CommandMap["PROT"] = &commandPROT{}
}
