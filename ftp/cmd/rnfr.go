package cmd

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

func (c commandRNFR) Execute(ctx context.Context, s Session, params []string) error {
	if len(params) == 0 {
		s.ReplyStatus(StatusSyntaxError)
		return nil
	}

	s.SetRenameFrom(params)

	s.ReplyStatus(StatusPendingMoreInfo)
	return nil
}

func init() {
	CommandMap["RNFR"] = &commandRNFR{}
}
