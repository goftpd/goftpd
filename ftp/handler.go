package ftp

import (
	"context"
	"fmt"
	"net"
	"os"
)

// handleConnection takes a context and a tcp connection and attempts to
// start a new session
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	fmt.Fprintln(os.Stdout, "New connection.")
}
