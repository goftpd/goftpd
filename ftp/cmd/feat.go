package cmd

import (
	"context"
	"strings"
	"sync"
)

type commandFEAT struct {
	once  sync.Once
	reply string
}

func (c commandFEAT) Feat() string               { return "" }
func (c commandFEAT) RequireParam() bool         { return false }
func (c commandFEAT) RequireState() SessionState { return SessionStateNull }

func (c *commandFEAT) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) > 0 {
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	// lets generate the Feat list on the first call
	// and store it for subsequent calls. Also means
	// no globals yay.
	c.once.Do(func() {
		if len(featSlice) == 0 {
			c.reply = "No Features."
			return
		}

		b := strings.Builder{}

		b.WriteString("Extensions supported:\n")

		for _, f := range featSlice {
			b.WriteString(" ")
			b.WriteString(f)
			b.WriteString("\n")
		}

		c.reply = b.String()
	})

	s.ReplyWithMessage(StatusSystemStatus, c.reply)
	return nil
}

func init() {
	CommandMap["FEAT"] = &commandFEAT{}
	featSlice = append(featSlice, "UTF8")
}
