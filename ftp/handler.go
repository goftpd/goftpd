package ftp

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/goftpd/goftpd/vfs"
)

// handleConnection takes a context and a tcp connection and attempts to
// start a new session
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	session := s.sessionPool.Get().(*Session)
	session.Reset()
	defer s.sessionPool.Put(session)

	session.serve(ctx, conn, s.fs)
}

// serve takes a connection and fs and parses commands on the control channel
// it traps any panics and attempts to close the session
func (s *Session) serve(ctx context.Context, conn net.Conn, fs vfs.VFS) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintf(os.Stderr, "SESSION PANIC: %v", e)
		}
		s.Close()
	}()

	s.control = conn
	s.controlReader = bufio.NewReader(conn)
	s.controlWriter = bufio.NewWriter(conn)

	s.Reply(220, "Welcome!")

	defer s.Close()

	for {
		line, err := s.controlReader.ReadString('\n')
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

		cmd, err := s.getCommand(fields[0], fields[1:])
		if err != nil {
			s.handleError(err)
			continue
		}

		if err := cmd.Do(s, fs, fields[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "SESSION ERROR: %v", err)
			break
		}
	}
}

// get command looks to see if there is a command registered under the cmdStr, it also
// does some generic validation such as checking params requirements and auth requirements
// returns Command and or
func (s *Session) getCommand(cmdStr string, params []string) (Command, error) {

	cmd, ok := commandMap[strings.ToUpper(cmdStr)]

	if !ok {
		return nil, newFTPError(502, "Command not implemented.")
	}

	if cmd.RequireParam() && len(params) == 0 {
		return nil, newFTPError(501, "Syntax error in parameters or arguments.")
	}

	if cmd.RequireAuth() && !s.isAuthed {
		return nil, newFTPError(530, "Not logged in.")
	}

	return cmd, nil
}
