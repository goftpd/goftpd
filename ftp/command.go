package ftp

import (
	"context"
	"errors"
)

// Execute should return this if they can no longer continue
var ErrCommandFatal = errors.New("command fatal")

// CommandFatalError is a special error wrapper that unwraps as
// ErrCommandFatal, but its String representation is of the real
// error
type CommandFatalError struct {
	internal error
}

func (e CommandFatalError) Error() string {
	return e.internal.Error()
}

func (e CommandFatalError) Unwrap() error {
	return ErrCommandFatal
}

type Command interface {
	RequireState() SessionState
	Execute(context.Context, *Session, []string) error
}

var commandMap = map[string]Command{}
var featSlice = []string{}
