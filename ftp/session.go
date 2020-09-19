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
)

type SessionState int

const (
	SessionStateNull SessionState = iota
	SessionStateAny
	SessionStateAuth
	SessionStateLoggedIn
)

// Session represents an FTP client connection's control
// channel
type Session struct {
	server *Server

	control *Control
	data    Data

	// state
	state         SessionState
	dataProtected bool
	binaryMode    bool

	// authentication
	loginUser string

	// fs abstract away?
	currentDir string
}

// State shows the current state of the session
func (s *Session) State() SessionState { return s.state }

type User struct {
	name   string
	groups []string
}

func (u User) Name() string { return u.name }
func (u User) PrimaryGroup() string {
	if len(u.groups) == 0 {
		return "nobody"
	}
	return u.groups[0]
}
func (u User) Groups() []string { return u.groups }

func (s *Session) User() (acl.User, bool) {
	if len(s.loginUser) > 0 {
		return User{s.loginUser, []string{s.loginUser}}, true
	}
	return nil, false
}

// Reset is used by sync.Pool and helps to minimise allocations
func (s *Session) Reset() {
	s.server = nil

	s.control = nil
	s.data = nil

	s.state = SessionStateNull
	s.dataProtected = false
	s.binaryMode = false

	s.loginUser = ""

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

// ReplyStatus replies with the default message for a status code
func (s *Session) ReplyStatus(st Status) error {
	return s.reply(st.Code, st.Message)
}

// ReplyStatusArgs replies with the default message for a status code
// but takes args
func (s *Session) ReplyWithArgs(st Status, args ...interface{}) error {
	return s.reply(st.Code, fmt.Sprintf(st.Message, args...))
}

// ReplyWithMessage replies with custom message
func (s *Session) ReplyWithMessage(st Status, message string) error {
	return s.reply(st.Code, message)
}

// reply is the underlying code for splitting a message across multiple lines
func (s *Session) reply(code int, message string) error {
	parts := strings.Split(message, "\n")

	b := strings.Builder{}

	if _, err := b.WriteString(fmt.Sprintf("%d", code)); err != nil {
		return CommandFatalError{err}
	}

	if len(parts) > 1 {
		if _, err := b.WriteString("-"); err != nil {
			return CommandFatalError{err}
		}
	}

	if _, err := b.WriteString(" "); err != nil {
		return CommandFatalError{err}
	}

	for _, p := range parts {
		if _, err := b.WriteString(p + "\r\n"); err != nil {
			return CommandFatalError{err}
		}
	}

	if len(parts) > 2 {
		if _, err := b.WriteString(fmt.Sprintf("%d End.", code)); err != nil {
			return CommandFatalError{err}
		}
	}

	if _, err := b.WriteString("\r\n"); err != nil {
		return CommandFatalError{err}
	}

	_, err := s.control.writer.WriteString(b.String())
	if err != nil {
		return CommandFatalError{err}
	}

	if err := s.control.writer.Flush(); err != nil {
		return CommandFatalError{err}
	}

	return nil
}

// upgrade a sessions underlying connection to use TLS
func (s *Session) upgrade() error {
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

	s.ReplyWithMessage(StatusServiceReady, "Welcome!")

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
	cmd, ok := commandMap[strings.ToUpper(fields[0])]

	if !ok {
		return session.ReplyStatus(StatusNotImplemented)
	}

	if session.State() < cmd.RequireState() {
		switch cmd.RequireState() {
		case SessionStateAuth:
			return session.ReplyWithMessage(StatusBadCommandSequence, "Please send AUTH first.")
		case SessionStateLoggedIn:
			return session.ReplyStatus(StatusNotLoggedIn)

		}
		return session.ReplyStatus(StatusNotImplemented)
	}

	// pre command hook
	if err := cmd.Execute(ctx, session, fields[1:]); err != nil {
		// check the type of the error, if its a fatal err then
		// return it, otherwise return nil to continue
		if errors.Is(err, ErrCommandFatal) {
			return err
		}

		return nil
	}

	// post command hook

	return nil
}
