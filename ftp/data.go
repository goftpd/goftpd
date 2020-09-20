package ftp

import (
	"io"
)

type Data interface {
	Host() string
	Port() int

	BytesRead() int
	BytesWritten() int

	io.Writer
	io.Reader
	io.Closer
}
