package ftp

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"strings"
	"time"
)

type activeDataConn struct {
	ctx context.Context

	conn net.Conn

	host string
	port int64

	written int
	read    int
}

func (s *Server) newActiveDataConn(ctx context.Context, param string, dataProtected bool) (*activeDataConn, error) {
	parts := strings.Split(param, ",")

	portOne, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, err
	}

	portTwo, err := strconv.Atoi(parts[5])
	if err != nil {
		return nil, err
	}

	port := int64((portOne * 256) + portTwo)
	host := parts[0] + "." + parts[1] + "." + parts[2] + "." + parts[3]

	d := activeDataConn{
		ctx:  ctx,
		host: host,
		port: port,
	}

	addr := net.JoinHostPort(host, strconv.Itoa(int(port)))

	dialer := net.Dialer{
		Timeout: time.Second * 60,
		// TODO: LocalAddr we probably want to be able to configure this
	}

	d.conn, err = dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	if dataProtected {
		/*
			That is to say, it does not matter which side initiates the
			connection with a connect() call or which side reacts to the
			connection via the accept() call; the FTP client, as defined in
			[RFC-959], is always the TLS client, as defined in [RFC-2246].
		*/
		d.conn = tls.Server(d.conn, s.TLSConfig())
	}

	return &d, nil
}

// Close implements the io.Closer
func (d *activeDataConn) Close() error {
	return d.conn.Close()
}

// Read implements the io.Reader interface and makes it context ware
func (d *activeDataConn) Read(p []byte) (int, error) {
	if err := d.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := d.conn.Read(p)
	d.read += n
	return n, err
}

// Write implements the io.Writer interface and makes it context ware
func (d *activeDataConn) Write(p []byte) (int, error) {
	if err := d.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := d.conn.Write(p)
	d.written += n
	return n, err
}

func (d *activeDataConn) Host() string      { return d.host }
func (d *activeDataConn) Port() int         { return int(d.port) }
func (d *activeDataConn) BytesRead() int    { return d.read }
func (d *activeDataConn) BytesWritten() int { return d.written }
