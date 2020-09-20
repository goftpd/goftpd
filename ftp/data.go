package ftp

import (
	"io"
)

type Data interface {
	Host() string
	Port() int

	io.Writer
	io.Reader
	io.Closer
}
