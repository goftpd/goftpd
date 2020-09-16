package ftp

import "github.com/goftpd/goftpd/vfs"

/*
	         ALLOCATE (ALLO)

            This command may be required by some servers to reserve
            sufficient storage to accommodate the new file to be
            transferred.  The argument shall be a decimal integer
            representing the number of bytes (using the logical byte
            size) of storage to be reserved for the file.  For files
            sent with record or page structure a maximum record or page
            size (in logical bytes) might also be necessary; this is
            indicated by a decimal integer in a second argument field of
            the command.  This second argument is optional, but when
            present should be separated from the first by the three
            Telnet characters <SP> R <SP>.  This command shall be
            followed by a STORe or APPEnd command.  The ALLO command
            should be treated as a NOOP (no operation) by those servers
            which do not require that the maximum size of the file be
            declared beforehand, and those servers interested in only
            the maximum record or page size should accept a dummy value
            in the first argument and ignore it.
*/

type commandALLO struct{}

func (c commandALLO) Feat() string               { return "" }
func (c commandALLO) RequireParam() bool         { return false }
func (c commandALLO) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandALLO) Do(s *Session, fs vfs.VFS, params []string) error {
	if err := s.Reply(202, "No storage allocation necessary."); err != nil {
		return err
	}

	return nil
}

func init() {
	commandMap["ALLO"] = &commandALLO{}
}
