package ftp

import (
	"fmt"
	"os"
)

type FTPError struct {
	Code    int
	Message string
}

func (e *FTPError) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Message)
}

func newFTPError(code int, message string) *FTPError {
	return &FTPError{
		Code:    code,
		Message: message,
	}
}

// handleError is a utility function for handline FTPErrors
func (s *Session) handleError(err error) {
	if e, ok := err.(*FTPError); ok {
		s.Reply(e.Code, e.Message)
		return
	}

	fmt.Fprintf(os.Stderr, "SESSION NONE FTPERROR: %v", err)

	s.Reply(500, "Unknown error occured.")
}
