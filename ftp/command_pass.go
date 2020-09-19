package ftp

import (
	"context"
	"fmt"
)

/*
 */

type commandPASS struct{}

func (c commandPASS) Feat() string               { return "PASS" }
func (c commandPASS) RequireParam() bool         { return true }
func (c commandPASS) RequireState() SessionState { return SessionStateAuth }

func (c commandPASS) Execute(ctx context.Context, s *Session, params []string) error {
	if len(params) != 1 {
		return s.ReplyStatus(StatusSyntaxError)
	}

	if len(s.loginUser) == 0 {
		return s.ReplyStatus(StatusBadCommandSequence)
	}

	if err := s.ReplyWithArgs(StatusUserLoggedIn, fmt.Sprintf("Welcome back %s!", s.loginUser)); err != nil {
		return err
	}

	s.state = SessionStateLoggedIn

	return nil
}

func init() {
	commandMap["PASS"] = &commandPASS{}
}
