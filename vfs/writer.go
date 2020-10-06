package vfs

import (
	"hash"
	"io"
)

// writer is a wrapper to the io.WriteCloser interface
// that lets us call a callback on success. Relies on the caller
// closing the writer. Very easy to make this context aware
type writeCloser struct {
	w              io.WriteCloser
	h              hash.Hash32
	err            error
	onCloseSuccess func(*writeCloser) error
}

// create a new writeCloser
func newWriteCloser(h hash.Hash32, w io.WriteCloser, onCloseSuccess func(w *writeCloser) error) *writeCloser {
	return &writeCloser{
		w:              w,
		h:              h,
		err:            nil,
		onCloseSuccess: onCloseSuccess,
	}
}

// Writer wraps the underlying Write function and saves any errors.
func (w *writeCloser) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if err != nil {
		w.err = err
		return n, err
	}

	_, err = w.h.Write(p)
	if err != nil {
		w.err = err
		return n, err
	}

	return n, nil
}

// Close closes the underlying io.WriteCloser and if no errors were
// made, it calls the onSuccess callback
func (w *writeCloser) Close() error {
	if err := w.w.Close(); err != nil {
		return err
	}

	if w.err == nil {
		if err := w.onCloseSuccess(w); err != nil {
			return err
		}
	}

	return nil
}
