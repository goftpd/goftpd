package ftp

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Session struct {
	control       net.Conn
	controlReader *bufio.Reader
	controlWriter *bufio.Writer

	data net.Conn

	// abstract away?
	homeDir    string
	currentDir string
}

func (s *Session) Reply(code int, message string) {
	parts := strings.Split(message, "\n")

	b := strings.Builder{}

	for idx := range parts {
		if idx < len(parts)-1 {
			b.WriteString(fmt.Sprintf("%d-%s\n", code, message))
		} else {
			b.WriteString(fmt.Sprintf("%d %s\r\n", code, message))
		}
	}

	_, err := s.controlWriter.WriteString(b.String())
	if err != nil {
		// TODO: what do we want to do here
		panic(err)
	}
}
