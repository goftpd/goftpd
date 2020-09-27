package script

import (
	"context"

	"github.com/goftpd/goftpd/ftp/cmd"
)

type Engine interface {
	Do(context.Context, []string, ScriptHook, cmd.Session) error
}

type ScriptHook string

const (
	ScriptHookPre  ScriptHook = "pre"
	ScriptHookPost            = "post"
)

type ScriptType string

const (
	ScriptTypeTrigger ScriptType = "trigger"
	ScriptTypeEvent              = "event"
)

type Command struct {
	FTPCommand string
	Hook       ScriptHook
	ScriptType ScriptType
	Path       string
}
