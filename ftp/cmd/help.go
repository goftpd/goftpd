package cmd

import (
	"context"
	"strings"
	"sync"
)

/*
   HELP (HELP)

      This command shall cause the server to send helpful
      information regarding its implementation status over the
      control connection to the user.  The command may take an
      argument (e.g., any command name) and return more specific
      information as a response.  The reply is type 211 or 214.
      It is suggested that HELP be allowed before entering a USER
      command. The server may use this reply to specify
      site-dependent parameters, e.g., in response to HELP SITE.
*/

type commandHELP struct {
	once  sync.Once
	reply string
}

func (c commandHELP) Feat() string               { return "" }
func (c commandHELP) RequireParam() bool         { return false }
func (c commandHELP) RequireState() SessionState { return SessionStateNull }

func (c *commandHELP) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) > 0 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	c.once.Do(func() {
		b := strings.Builder{}
		b.WriteString("Commands Supported:\n")

		for c := range CommandMap {
			b.WriteString(" ")
			b.WriteString(c)
			b.WriteString("\n")
		}

		c.reply = b.String()

		// we could add a HelpInfo to the command interface so that we could reply to individual
		// help commands and support params > 0
	})

	return s.ReplyWithMessage(StatusSystemStatus, c.reply)
}

func init() {
	CommandMap["HELP"] = &commandHELP{}
}
