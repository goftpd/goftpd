package cmd

import (
	"context"
	"errors"
	"io"

	"github.com/goftpd/goftpd/acl"
	"github.com/goftpd/goftpd/vfs"
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

func NewFatalError(err error) CommandFatalError {
	return CommandFatalError{err}
}

type DataConn interface {
	Host() string
	Port() int

	Kind() string

	BytesRead() int
	BytesWritten() int

	io.Writer
	io.Reader
	io.Closer
}

type SessionState int

const (
	SessionStateNull SessionState = iota
	SessionStateAny
	SessionStateAuth
	SessionStateLoggedIn
)

type Session interface {
	// reply
	Reply(int, string)
	ReplyWithMessage(Status, string)
	ReplyWithArgs(Status, ...interface{})
	ReplyError(Status, error)
	ReplyStatus(Status)
	Flush() error

	// TLS
	Upgrade() error

	Close() error

	// filesystem
	FS() vfs.VFS
	Auth() acl.Authenticator

	// data
	Data() DataConn
	ClearData()
	NewPassiveDataConn(context.Context) error
	NewActiveDataConn(context.Context, string) error

	// state
	State() SessionState
	SetState(SessionState)

	SetBinaryMode(bool)
	BinaryMode() bool

	SetDataProtected(bool)
	DataProtected() bool

	SetRestartPosition(int)
	RestartPosition() int

	SetRenameFrom([]string)
	RenameFrom() []string

	SetCWD(string)
	CWD() string

	SetLogin(string)
	Login() string

	User() *acl.User

	LastCommand() string
}

type Command interface {
	RequireState() SessionState
	Execute(context.Context, Session, []string) error
}

var CommandMap = map[string]Command{}
var featSlice = []string{}
