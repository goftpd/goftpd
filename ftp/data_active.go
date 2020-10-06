package ftp

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spacemonkeygo/openssl"
)

type activeDataConn struct {
	ctx context.Context

	tlsConfig *tls.Config
	sslCtx    *openssl.Ctx

	conn net.Conn

	host string
	port int64

	written int
	read    int

	sync.Mutex
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
		ctx:    ctx,
		host:   host,
		port:   port,
		sslCtx: s.sslCtx,
	}

	if dataProtected {
		d.tlsConfig = s.TLSConfig()
	}

	return &d, nil
}

// Connect attempts to connect to the underlying connection
func (d *activeDataConn) connect() error {

	d.Lock()
	defer d.Unlock()
	if d.conn != nil {
		return nil
	}

	addr := net.JoinHostPort(d.host, strconv.Itoa(int(d.port)))

	dialer := net.Dialer{
		Timeout: time.Second * 60,
		// TODO: LocalAddr we probably want to be able to configure this
	}

	var err error

	d.conn, err = dialer.DialContext(d.ctx, "tcp", addr)
	if err != nil {
		return err
	}

	if d.tlsConfig != nil {
		// TODO make this optional?
		d.conn, err = openssl.Server(d.conn, d.sslCtx)
		if err != nil {
			return err
		}

		/*
			That is to say, it does not matter which side initiates the
			connection with a connect() call or which side reacts to the
			connection via the accept() call; the FTP client, as defined in
			[RFC-959], is always the TLS client, as defined in [RFC-2246].
		*/
		// d.conn = tls.Server(d.conn, d.tlsConfig)
	}

	return nil
}

// Close implements the io.Closer
func (d *activeDataConn) Close() error {
	d.Lock()
	defer d.Unlock()

	if d.conn == nil {
		return nil
	}
	return d.conn.Close()
}

// Read implements the io.Reader interface and makes it context ware
func (d *activeDataConn) Read(p []byte) (int, error) {

	if err := d.ctx.Err(); err != nil {
		return 0, err
	}

	if err := d.connect(); err != nil {
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

	if err := d.connect(); err != nil {
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
func (d *activeDataConn) Kind() string      { return "Active" }
