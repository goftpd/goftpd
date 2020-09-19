package ftp

import (
	"context"
	"fmt"
	"strings"
)

/*
   AUTHENTICATION/SECURITY MECHANISM (AUTH)

      The argument field is a Telnet string identifying a supported
      mechanism.  This string is case-insensitive.  Values must be
      registered with the IANA, except that values beginning with "X-"
      are reserved for local use.

      If the server does not recognize the AUTH command, it must respond
      with reply code 500.  This is intended to encompass the large
      deployed base of non-security-aware ftp servers, which will
      respond with reply code 500 to any unrecognized command.  If the
      server does recognize the AUTH command but does not implement the
      security extensions, it should respond with reply code 502.

      If the server does not understand the named security mechanism, it
      should respond with reply code 504.

      If the server is not willing to accept the named security
      mechanism, it should respond with reply code 534.

      If the server is not able to accept the named security mechanism,
      such as if a required resource is unavailable, it should respond
      with reply code 431.

      If the server is willing to accept the named security mechanism,
      but requires security data, it must respond with reply code 334.

      If the server is willing to accept the named security mechanism,
      and does not require any security data, it must respond with reply
      code 234.

      If the server is responding with a 334 reply code, it may include
      security data as described in the next section.

      Some servers will allow the AUTH command to be reissued in order
      to establish new authentication.  The AUTH command, if accepted,
      removes any state associated with prior FTP Security commands.
      The server must also require that the user reauthorize (that is,
      reissue some or all of the USER, PASS, and ACCT commands) in this
      case (see section 4 for an explanation of "authorize" in this
      context).
*/

type commandAUTH struct{}

func (c commandAUTH) RequireState() SessionState { return SessionStateNull }

func (c commandAUTH) Execute(ctx context.Context, s *Session, params []string) error {
	if len(params) != 1 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	if strings.ToUpper(params[0]) != "TLS" {
		return s.ReplyWithMessage(
			StatusParameterNotImplemented,
			fmt.Sprintf("Security Mechanism '%s' not supported", params[0]),
		)
	}

	s.ReplyStatus(StatusSecurityExchangeOK)

	if err := s.upgrade(); err != nil {
		return CommandFatalError{err}
	}

	s.state = SessionStateAuth

	return nil
}

func init() {
	commandMap["AUTH"] = &commandAUTH{}
	featSlice = append(featSlice, "AUTH TLS")
}
