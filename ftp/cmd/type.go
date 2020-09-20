package cmd

import (
	"context"
	"strings"
)

/*
         REPRESENTATION TYPE (TYPE)

            The argument specifies the representation type as described
            in the Section on Data Representation and Storage.  Several
            types take a second parameter.  The first parameter is
            denoted by a single Telnet character, as is the second
            Format parameter for ASCII and EBCDIC; the second parameter
            for local byte is a decimal integer to indicate Bytesize.
            The parameters are separated by a <SP> (Space, ASCII code
            32).

            The following codes are assigned for type:

                         \    /
               A - ASCII |    | N - Non-print
                         |-><-| T - Telnet format effectors
               E - EBCDIC|    | C - Carriage Control (ASA)
                         /    \
               I - Image

               L <byte size> - Local byte Byte size

			The default representation type is ASCII Non-print.  If the
            Format parameter is changed, and later just the first
            argument is changed, Format then returns to the Non-print
            default.

*/

type commandTYPE struct{}

func (c commandTYPE) RequireState() SessionState { return SessionStateLoggedIn }

func (c commandTYPE) Execute(ctx context.Context, s Session, params []string) error {
	switch strings.ToUpper(params[0]) {
	case "A":
		// TODO:
		// according to https://cr.yp.to/ftp/type.html we should check for second
		// param
		s.SetBinaryMode(false)
	case "I":
		s.SetBinaryMode(true)
	case "L":
		// TODO:
		// according to https://cr.yp.to/ftp/type.html we should check for second
		// param
		s.SetBinaryMode(true)

	default:
		return s.ReplyStatus(StatusSyntaxError)
	}

	return s.ReplyStatus(StatusOK)
}

func init() {
	CommandMap["TYPE"] = &commandTYPE{}
}
