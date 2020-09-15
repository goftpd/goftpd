package ftp

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/sync/errgroup"
)

// ListenAndServe creates a new tcp listener on the configured Host and Port.
// New connections are buffered down a channel before being given their own
// goroutine. Takes a context and attemps to shutdown on cancellation/deadline
func (s *Server) ListenAndServe(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	addr := net.JoinHostPort(s.Host, fmt.Sprintf("%d", s.Port))

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()

	conns := make(chan net.Conn, 10)

	var errg errgroup.Group

	errg.Go(func() error {
		for {
			conn, err := l.Accept()
			if err != nil {

				// check if this is a cancellation
				select {
				case <-ctx.Done():
					return nil
				default:
				}

				// check if this is temporary
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					continue
				}

				// fatal cancel ctx and return error
				cancel()

				return err
			}

			conns <- conn
		}

		return nil
	})

	errg.Go(func() error {
		for {
			select {
			case c := <-conns:
				go s.handleConnection(ctx, c)

			case <-ctx.Done():

				// attempt to close the listener
				if err := l.Close(); err != nil {
					return err
				}

				return nil
			}
		}

		return nil
	})

	if err := errg.Wait(); err != nil {
		return err
	}

	return nil
}
