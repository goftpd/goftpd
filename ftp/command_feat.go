package ftp

import (
	"strings"
	"sync"

	"github.com/goftpd/goftpd/vfs"
)

type commandFEAT struct {
	once  sync.Once
	reply string
}

func (c commandFEAT) Feat() string               { return "" }
func (c commandFEAT) RequireParam() bool         { return false }
func (c commandFEAT) RequireState() SessionState { return SessionStateNull }

func (c *commandFEAT) Do(s *Session, fs vfs.VFS, params []string) error {
	if len(params) > 0 {
		if err := s.Reply(501, "Syntax error in parameters or arguments."); err != nil {
			return err
		}
		return nil
	}

	// lets generate the Feat list on the first call
	// and store it for subsequent calls. Also means
	// no globals yay.
	c.once.Do(func() {
		var feats []string

		for k := range commandMap {
			f := commandMap[k].Feat()
			if len(f) > 0 {
				feats = append(feats, f)
			}
		}

		if len(feats) == 0 {
			c.reply = "No Features."
			return
		}

		b := strings.Builder{}

		b.WriteString("Extensions supported:\n")

		for _, f := range feats {
			b.WriteString(" ")
			b.WriteString(f)
			b.WriteString("\n")
		}

		c.reply = b.String()
	})

	s.Reply(211, c.reply)
	return nil
}

func init() {
	commandMap["FEAT"] = &commandFEAT{}
}
