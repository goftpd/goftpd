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
	"sync"

	"github.com/goftpd/goftpd/acl"
	"github.com/goftpd/goftpd/ftp/cmd"
	"github.com/goftpd/goftpd/script"
	"github.com/goftpd/goftpd/vfs"
	"github.com/spacemonkeygo/openssl"
)

// Session represents an FTP client connection's control
// channel
type Session struct {
	server *Server

	active    bool
	activeMtx sync.Mutex

	control *Control
	data    cmd.DataConn

	//
	sbuilder strings.Builder

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

// Control gets the underlying connection
func (s *Session) Control() net.Conn { return s.control }

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

func (s *Session) User() *acl.User {
	if s.state != cmd.SessionStateLoggedIn {
		return nil
	}

	u, err := s.server.auth.GetUser(s.login)
	if err != nil {
		return nil
	}
	return u
}

// Reset is used by sync.Pool and helps to minimise allocations
func (s *Session) Reset() {
	s.server = nil

	s.activeMtx.Lock()
	s.active = false
	s.activeMtx.Unlock()

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
	s.activeMtx.Lock()
	s.active = false
	s.activeMtx.Unlock()

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
	for idx := range parts {
		if len(parts[idx]) > 0 {
			s.buffer = append(s.buffer, parts[idx])
		}
	}
}

func (s *Session) Flush() error {
	defer func() {
		s.code = 0
		s.buffer = nil
	}()

	if len(s.buffer) == 0 {
		return nil
	}

	s.activeMtx.Lock()
	if !s.active {
		s.activeMtx.Unlock()
		return nil
	}
	s.activeMtx.Unlock()

	s.sbuilder.Reset()

	if _, err := s.sbuilder.WriteString(fmt.Sprintf("%d", s.code)); err != nil {
		return cmd.NewFatalError(err)
	}

	if len(s.buffer) > 1 {
		if _, err := s.sbuilder.WriteString("-"); err != nil {
			return cmd.NewFatalError(err)
		}
	}

	if _, err := s.sbuilder.WriteString(" "); err != nil {
		return cmd.NewFatalError(err)
	}

	for _, p := range s.buffer {
		if len(p) == 0 {
			continue
		}
		if _, err := s.sbuilder.WriteString(p + "\r\n"); err != nil {
			return cmd.NewFatalError(err)
		}
	}

	if len(s.buffer) > 1 {
		if _, err := s.sbuilder.WriteString(fmt.Sprintf("%d End.\r\n", s.code)); err != nil {
			return cmd.NewFatalError(err)
		}
	}

	_, err := s.control.writer.WriteString(s.sbuilder.String())
	if err != nil {
		return cmd.NewFatalError(err)
	}

	if err := s.control.writer.Flush(); err != nil {
		return cmd.NewFatalError(err)
	}

	if len(os.Getenv("DEBUG")) > 0 {
		fmt.Fprintf(os.Stderr, ">>> %s", s.sbuilder.String())
	}

	return nil
}

// Upgrade a sessions underlying connection to use TLS
func (s *Session) Upgrade() error {
	// TODO make this optional?
	conn, err := openssl.Server(s.control, s.server.sslCtx)
	if err != nil {
		return err
	}

	if err := conn.Handshake(); err != nil {
		return err
	}

	s.control = newControl(conn)

	return nil

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
			fmt.Fprintf(&buf, "Handler crashed with error: %v\n", e)

			for i := 1; ; i++ {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				} else {
					fmt.Fprintf(&buf, "\n")
				}
				fmt.Fprintf(&buf, "%v:%v", file, line)
			}

			fmt.Fprintf(os.Stderr, "%s\n", buf.String())
		}
		s.Close()
	}()

	s.control = newControl(conn)
	s.server = server
	s.active = true

	s.ReplyWithMessage(cmd.StatusServiceReady, "Welcome!")
	if err := s.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR flush session welcome: %s\n", err)
		return
	}

	defer s.Close()

	for {
		line, err := s.control.reader.ReadString('\n')
		if err != nil {
			break
		}

		if len(os.Getenv("DEBUG")) > 0 {
			fmt.Fprintf(os.Stderr, "<<< %s", line)
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
			fmt.Fprintf(os.Stderr, "ERROR handleCommand: %s\n", err)
			break
		}
	}
}

// handleCommand takes in the client input in the form of a slice of strings
// and tries to find and execute a command. can return an error
func (session *Session) handleCommand(ctx context.Context, fields []string) error {
	// start := time.Now()

	ftpCommand := strings.ToUpper(fields[0])

	defer func() {
		// log.Printf("%s - DEFER - %s", ftpCommand, time.Since(start))
		if err := session.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR session flush: %s\n", err)
		}
	}()

	// TODO: ugly as sin
	c, ok := cmd.CommandMap[ftpCommand]

	if !ok {

		// if logged in, check if we have a script that uses this command
		if session.State() == cmd.SessionStateLoggedIn {
			err := session.server.se.Do(ctx, fields, script.ScriptHookCommand, session)

			switch err {

			case script.ErrNotExist:
				session.ReplyStatus(cmd.StatusNotImplemented)
				break

			case script.ErrStop:
				break

			case nil:
				break

			default:
				session.Reply(500, "Error in script.")
				return err
			}

			return nil
		}

		session.ReplyStatus(cmd.StatusNotImplemented)

		return nil
	}

	/*
		2020/10/07 07:10:58 TYPE - PRE MAP - 90ns
		2020/10/07 07:10:58 TYPE - PRE STATE - 9.177µs
		2020/10/07 07:10:58 TYPE - PRE USER CHK - 13.05µs
		2020/10/07 07:10:58 TYPE - PRE HOOK - 1.059119ms
		2020/10/07 07:10:58 TYPE - PRE EXEC - 1.067454ms
		2020/10/07 07:10:58 TYPE - PRE POST - 1.071809ms
		2020/10/07 07:10:58 TYPE - DEFER - 1.075274ms
	*/

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

	/*
		if session.State() == cmd.SessionStateLoggedIn {
			user := session.User()
			if user == nil || !user.DeletedAt.IsZero() {
				return errors.New("deleted user")
			}
		}
	*/

	// pre command hook
	if err := session.server.se.Do(ctx, fields, script.ScriptHookPre, session); err != nil {
		if err != script.ErrNotExist {
			if err == script.ErrStop {
				return nil
			}
			session.Reply(500, "Error in script.")
			return err
		}
	}

	if err := c.Execute(ctx, session, fields[1:]); err != nil {
		// check the type of the error, if its a fatal err then
		// return it, otherwise return nil to continue
		if errors.Is(err, cmd.ErrCommandFatal) {
			return err
		}

		return nil
	}

	session.lastCommand = ftpCommand

	// post command hook
	if err := session.server.se.Do(ctx, fields, script.ScriptHookPost, session); err != nil {
		if err != script.ErrNotExist {
			if err == script.ErrStop {
				return nil
			}
			session.Reply(500, "Error in script.")
			return err
		}
	}

	return nil
}
