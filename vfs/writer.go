package vfs

import "io"

// writer is a wrapper to the io.WriteCloser interface
// that lets us call a callback on success. Relies on the caller
// closing the writer. Very easy to make this context aware
type writeCloser struct {
	w              io.WriteCloser
	err            error
	onCloseSuccess func() error
}

// create a new writeCloser
func newWriteCloser(w io.WriteCloser, onCloseSuccess func() error) *writeCloser {
	return &writeCloser{
		w:              w,
		err:            nil,
		onCloseSuccess: onCloseSuccess,
	}
}

// Writer wraps the underlying Write function and saves any errors.
func (w *writeCloser) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if err != nil {
		w.err = err
	}
	return n, err
}

// Close closes the underlying io.WriteCloser and if no errors were
// made, it calls the onSuccess callback
func (w *writeCloser) Close() error {
	if err := w.w.Close(); err != nil {
		return err
	}

	if w.err == nil {
		if err := w.onCloseSuccess(); err != nil {
			return err
		}
	}

	return nil
}
