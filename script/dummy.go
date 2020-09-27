package script

import (
	"context"

	"github.com/goftpd/goftpd/ftp/cmd"
)

type DummyEngine struct{}

func (d *DummyEngine) Do(ctx context.Context, fields []string, hook ScriptHook, s cmd.Session) error {
	return nil
}
