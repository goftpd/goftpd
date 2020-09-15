package ftp

import (
	"strings"

	"github.com/goftpd/goftpd/vfs"
)

type commandFEAT struct{}

func (c commandFEAT) IsExtension() bool  { return false }
func (c commandFEAT) RequireParam() bool { return false }
func (c commandFEAT) RequireAuth() bool  { return false }

func (c commandFEAT) Do(s *Session, fs vfs.VFS, params []string) error {
	if len(params) > 0 {
		if err := s.Reply(501, "Syntax error in parameters or arguments."); err != nil {
			return err
		}
		return nil
	}

	// this all wants moving to an init or cached
	var count int
	for k := range commandMap {
		if commandMap[k].IsExtension() {
			count++
		}
	}

	if count == 0 {
		s.Reply(211, "No Features.")
		return nil
	}

	b := strings.Builder{}

	b.WriteString("Extensions supported:\n")

	for k := range commandMap {
		if commandMap[k].IsExtension() {
			b.WriteString(" ")
			b.WriteString(k)
			b.WriteString("\n")
		}
	}

	s.Reply(211, b.String())
	return nil
}

func init() {
	commandMap["FEAT"] = &commandFEAT{}
}
