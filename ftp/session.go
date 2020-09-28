package ftp

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/goftpd/goftpd/acl"
	"github.com/goftpd/goftpd/ftp/cmd"
	"github.com/goftpd/goftpd/script"
	"github.com/goftpd/goftpd/vfs"
)

// Session represents an FTP client connection's control
// channel
type Session struct {
	server *Server

	control *Control
	data    cmd.DataConn

	// state
	state           cmd.SessionState
	dataProtected   bool
	binaryMode      bool
	lastCommand     string
	renameFrom      []string
	restartPosition int

	// message state
	code   int
	buffer []string

	// authentication
	login string

	// fs abstract away?
	currentDir string
}

// SetState sets the current state of the session
func (s *Session) SetState(state cmd.SessionState) { s.state = state }

// State shows the current state of the session
func (s *Session) State() cmd.SessionState { return s.state }

// SetBinaryMode sets the current state of the session
func (s *Session) SetBinaryMode(t bool) { s.binaryMode = t }

// BinaryMode shows the current state of the session
func (s *Session) BinaryMode() bool { return s.binaryMode }

// SetDataProtected sets the current state of the session
func (s *Session) SetDataProtected(t bool) { s.dataProtected = t }

// DataProtected shows the current state of the session
func (s *Session) DataProtected() bool { return s.dataProtected }

// SetRestartPosition sets the current state of the session
func (s *Session) SetRestartPosition(t int) { s.restartPosition = t }

// RestartPosition shows the current state of the session
func (s *Session) RestartPosition() int { return s.restartPosition }

// SetRenameFrom sets the current state of the session
func (s *Session) SetRenameFrom(t []string) { s.renameFrom = t }

// CWD gets the current working directory
func (s *Session) CWD() string { return s.currentDir }

// SetCWD sets the current working directory
func (s *Session) SetCWD(t string) { s.currentDir = t }

// LastCommnad returns the last command to be successful
func (s *Session) LastCommand() string { return s.lastCommand }

// RenameFrom shows the current state of the session
func (s *Session) RenameFrom() []string { return s.renameFrom }

// SetLogin sets the current state of the session
func (s *Session) SetLogin(t string) { s.login = t }

// Login shows the current state of the session
func (s *Session) Login() string { return s.login }

func (s *Session) Data() cmd.DataConn { return s.data }
func (s *Session) ClearData()         { s.data = nil }
func (s *Session) NewPassiveDataConn(ctx context.Context) error {
	d, err := s.server.newPassiveDataConn(ctx, s.dataProtected)
	if err != nil {
		return err
	}
	s.data = d
	return nil
}
func (s *Session) NewActiveDataConn(ctx context.Context, params string) error {
	d, err := s.server.newActiveDataConn(ctx, params, s.dataProtected)
	if err != nil {
		return err
	}
	s.data = d
	return nil
}

func (s *Session) FS() vfs.VFS             { return s.server.fs }
func (s *Session) Auth() acl.Authenticator { return s.server.auth }

func (s *Session) User() (*acl.User, bool) {
	u, err := s.server.auth.GetUser(s.login)
	if err != nil {
		return nil, false
	}
	return u, true
}

// Reset is used by sync.Pool and helps to minimise allocations
func (s *Session) Reset() {
	s.server = nil

	s.control = nil
	s.data = nil

	s.state = cmd.SessionStateNull
	s.dataProtected = false
	s.binaryMode = false
	s.lastCommand = ""
	s.renameFrom = []string{}
	s.restartPosition = 0

	s.code = 0
	s.buffer = []string{}

	s.login = ""

	s.currentDir = "/"
}

// Close attempts to gracefully close the control and any running
// data connections
func (s *Session) Close() error {
	if err := s.control.Close(); err != nil {
		return err
	}

	if s.data != nil {
		if err := s.data.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Reply replies with the int and message, mainly for scripts convenience
func (s *Session) Reply(code int, message string) {
	s.reply(code, message)
}

// ReplyStatus replies with the default message for a status code
func (s *Session) ReplyStatus(st cmd.Status) {
	s.reply(st.Code, st.Message)
}

// ReplyStatusArgs replies with the default message for a status code
// but takes args
func (s *Session) ReplyWithArgs(st cmd.Status, args ...interface{}) {
	s.reply(st.Code, fmt.Sprintf(st.Message, args...))
}

// ReplyError replies with the default message for a status code
// but takes args
func (s *Session) ReplyError(st cmd.Status, err error) {
	s.reply(st.Code, fmt.Sprintf("%s (%s)", st.Message, err.Error()))
}

// ReplyWithMessage replies with custom message
func (s *Session) ReplyWithMessage(st cmd.Status, message string) {
	s.reply(st.Code, message)
}

// reply is the underlying code for splitting a message across multiple lines
func (s *Session) reply(code int, message string) {
	parts := strings.Split(message, "\n")

	s.code = code
	s.buffer = append(s.buffer, parts...)
}

func (s *Session) Flush() error {
	defer func() {
		s.code = 0
		s.buffer = nil
	}()

	b := strings.Builder{}

	if _, err := b.WriteString(fmt.Sprintf("%d", s.code)); err != nil {
		return cmd.NewFatalError(err)
	}

	if len(s.buffer) > 1 {
		if _, err := b.WriteString("-"); err != nil {
			return cmd.NewFatalError(err)
		}
	}

	if _, err := b.WriteString(" "); err != nil {
		return cmd.NewFatalError(err)
	}

	for _, p := range s.buffer {
		if _, err := b.WriteString(p + "\r\n"); err != nil {
			return cmd.NewFatalError(err)
		}
	}

	if len(s.buffer) > 2 {
		if _, err := b.WriteString(fmt.Sprintf("%d End.", s.code)); err != nil {
			return cmd.NewFatalError(err)
		}
	}

	if _, err := b.WriteString("\r\n"); err != nil {
		return cmd.NewFatalError(err)
	}

	_, err := s.control.writer.WriteString(b.String())
	if err != nil {
		return cmd.NewFatalError(err)
	}

	if err := s.control.writer.Flush(); err != nil {
		return cmd.NewFatalError(err)
	}

	return nil
}

// Upgrade a sessions underlying connection to use TLS
func (s *Session) Upgrade() error {
	tlsConn := tls.Server(s.control, s.server.TLSConfig())
	if err := tlsConn.Handshake(); err != nil {
		return err
	}

	s.control = newControl(tlsConn)

	return nil
}

// serve takes a connection and fs and parses commands on the control channel
// it traps any panics and attempts to close the session
func (s *Session) serve(ctx context.Context, server *Server, conn net.Conn) {
	defer func() {
		if e := recover(); e != nil {
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "Handler crashed with error: %v", e)

			for i := 1; ; i++ {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				} else {
					fmt.Fprintf(&buf, "\n")
				}
				fmt.Fprintf(&buf, "%v:%v", file, line)
			}

			fmt.Fprintf(os.Stderr, "%s", buf.String())
		}
		s.Close()
	}()

	s.control = newControl(conn)
	s.server = server

	s.ReplyWithMessage(cmd.StatusServiceReady, "Welcome!")
	if err := s.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR flush session welcome: %s", err)
		return
	}

	defer s.Close()

	for {
		line, err := s.control.reader.ReadString('\n')
		if err != nil {
			break
		}

		// check for cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}

		fields := strings.Fields(line)

		if len(fields) == 0 {
			continue
		}

		if err := s.handleCommand(ctx, fields); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR handleCommand: %s", err)
			break
		}

	}
}

// handleCommand takes in the client input in the form of a slice of strings
// and tries to find and execute a command. can return an error
func (session *Session) handleCommand(ctx context.Context, fields []string) error {
	// TODO: ugly as sin
	c, ok := cmd.CommandMap[strings.ToUpper(fields[0])]

	defer func() {
		if err := session.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR session flush: %s", err)
		}

	}()

	if !ok {
		session.ReplyStatus(cmd.StatusNotImplemented)
		return nil
	}

	if session.State() < c.RequireState() {
		switch c.RequireState() {
		case cmd.SessionStateAuth:
			session.ReplyWithMessage(cmd.StatusBadCommandSequence, "Please send AUTH first.")
		case cmd.SessionStateLoggedIn:
			session.ReplyStatus(cmd.StatusNotLoggedIn)
		default:
			session.ReplyStatus(cmd.StatusNotImplemented)
		}
		return nil
	}

	// pre command hook
	if err := session.server.se.Do(ctx, fields, script.ScriptHookPre, session); err != nil {
		return err
	}

	if err := c.Execute(ctx, session, fields[1:]); err != nil {
		// check the type of the error, if its a fatal err then
		// return it, otherwise return nil to continue
		if errors.Is(err, cmd.ErrCommandFatal) {
			return err
		}

		return nil
	}

	session.lastCommand = strings.ToUpper(fields[0])

	// post command hook
	if err := session.server.se.Do(ctx, fields, script.ScriptHookPost, session); err != nil {
		return err
	}

	return nil
}
