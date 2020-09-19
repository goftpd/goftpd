package ftp

import (
	"bufio"
	"net"
)

type Control struct {
	net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func newControl(conn net.Conn) *Control {
	c := Control{
		Conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}

	return &c
}
