package ftp

import (
	"errors"

	"github.com/goftpd/goftpd/vfs"
)

/*
   CLEAR COMMAND CHANNEL (CCC)

      This command does not take an argument.
      It is desirable in some environments to use a security mechanism
      to authenticate and/or authorize the client and server, but not to
      perform any integrity checking on the subsequent commands.  This
      might be used in an environment where IP security is in place,
      insuring that the hosts are authenticated and that TCP streams
      cannot be tampered, but where user authentication is desired.

      If unprotected commands are allowed on any connection, then an
      attacker could insert a command on the control stream, and the
      server would have no way to know that it was invalid.  In order to
      prevent such attacks, once a security data exchange completes
      successfully, if the security mechanism supports integrity, then
      integrity (via the MIC or ENC command, and 631 or 632 reply) must
      be used, until the CCC command is issued to enable non-integrity
      protected control channel messages.  The CCC command itself must
      be integrity protected.

      Once the CCC command completes successfully, if a command is not
      protected, then the reply to that command must also not be
      protected.  This is to support interoperability with clients which
      do not support protection once the CCC command has been issued.

      This command must be preceded by a successful security data
      exchange.

      If the command is not integrity-protected, the server must respond
      with a 533 reply code.

      If the server is not willing to turn off the integrity
      requirement, it should respond with a 534 reply code.

      Otherwise, the server must reply with a 200 reply code to indicate
      that unprotected commands and replies may now be used on the
      command channel.
*/

type commandCCC struct{}

func (c commandCCC) Feat() string               { return "CCC" }
func (c commandCCC) RequireParam() bool         { return true }
func (c commandCCC) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandCCC) Do(s *Session, fs vfs.VFS, params []string) error {
	if len(params) != 0 {
		s.Reply(501, "Syntax error in parameters or arguments.")
		return nil
	}

	if s.tlsConfig == nil {
		return errors.New("TLS Config is nil")
	}

	// boh

	s.Reply(200, "OK.")

	return nil
}

func init() {
	commandMap["CCC"] = &commandCCC{}
}
