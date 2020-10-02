package script

import (
	"context"

	"github.com/goftpd/goftpd/acl"
	"github.com/goftpd/goftpd/ftp/cmd"
	"github.com/pkg/errors"
)

var (
	ErrNotExist = errors.New("does not exist")
	ErrStop     = errors.New("stop")
)

type Engine interface {
	Do(context.Context, []string, ScriptHook, cmd.Session) error
}

type ScriptHook string

const (
	ScriptHookPre     ScriptHook = "pre"
	ScriptHookPost               = "post"
	ScriptHookCommand            = "command"
)

var stringToScriptHook = map[string]ScriptHook{
	string(ScriptHookPre):     ScriptHookPre,
	string(ScriptHookPost):    ScriptHookPost,
	string(ScriptHookCommand): ScriptHookCommand,
}

type ScriptType string

const (
	ScriptTypeTrigger ScriptType = "trigger"
	ScriptTypeEvent              = "event"
)

var stringToScriptType = map[string]ScriptType{
	string(ScriptTypeTrigger): ScriptTypeTrigger,
	// TODO
	// events dont work as expected as some ftp clients (filezilla) cant parse
	// multi line PASV. also need to be careful about session becomming nil.
	// string(ScriptTypeEvent):   ScriptTypeEvent,
}

type Command struct {
	FTPCommand string
	Hook       ScriptHook
	ScriptType ScriptType
	Path       string
	ACL        *acl.ACL
}
