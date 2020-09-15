package ftp

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Session represents an FTP client connection's control
// channel
type Session struct {
	control       net.Conn
	controlReader *bufio.Reader
	controlWriter *bufio.Writer

	data net.Conn

	isAuthed bool

	// abstract away?
	currentDir string
}

// Reset is used by sync.Pool and helps to minimise allocations
func (s *Session) Reset() {
	s.control = nil
	s.controlReader = nil
	s.controlWriter = nil
	s.data = nil
	s.isAuthed = false
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

// Reply sends a reply down the control channel if it encounters an error
// it returns
func (s *Session) Reply(code int, message string) error {
	parts := strings.Split(message, "\n")

	b := strings.Builder{}

	if _, err := b.WriteString(fmt.Sprintf("%d", code)); err != nil {
		return err
	}

	if len(parts) > 1 {
		if _, err := b.WriteString("-"); err != nil {
			return err
		}
	} else {
		if _, err := b.WriteString(" "); err != nil {
			return err
		}
	}

	for _, p := range parts {
		if _, err := b.WriteString(p + "\r\n"); err != nil {
			return err
		}
	}

	if len(parts) > 2 {
		if _, err := b.WriteString(fmt.Sprintf("%d End.", code)); err != nil {
			return err
		}
	}

	if _, err := b.WriteString("\r\n"); err != nil {
		return err
	}

	_, err := s.controlWriter.WriteString(b.String())
	if err != nil {
		return err
	}

	if err := s.controlWriter.Flush(); err != nil {
		return err
	}

	return nil
}
