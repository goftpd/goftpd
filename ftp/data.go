package ftp

import (
	"io"
)

type Data interface {
	Host() string
	Port() int

	Kind() string

	BytesRead() int
	BytesWritten() int

	io.Writer
	io.Reader
	io.Closer
}
