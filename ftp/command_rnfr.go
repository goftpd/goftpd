package ftp

import (
	"context"
)

/*
   RENAME FROM (RNFR)

      This command specifies the old pathname of the file which is
      to be renamed.  This command must be immediately followed by
      a "rename to" command specifying the new file pathname.
*/

type commandRNFR struct{}

func (c commandRNFR) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandRNFR) Execute(ctx context.Context, s *Session, params []string) error {
	if len(params) == 0 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	s.renameFrom = params

	return s.ReplyStatus(StatusPendingMoreInfo)
}

func init() {
	commandMap["RNFR"] = &commandRNFR{}
}
