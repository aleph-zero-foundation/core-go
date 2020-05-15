// Package tcp implements network.Connections that wrap around TCP connections.
package tcp

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"gitlab.com/alephledger/core-go/pkg/network"
)

const (
	bufSize = 32000
)

type conn struct {
	*bufio.Reader
	*bufio.Writer
	link net.Conn
}

// newConn creates a Connection object wrapping a particular tcp connection link.
func newConn(link net.Conn) network.Connection {
	return &conn{
		link:   link,
		Reader: bufio.NewReaderSize(link, bufSize),
		Writer: bufio.NewWriterSize(link, bufSize),
	}
}

func (c *conn) Read(b []byte) (int, error) {
	err := c.link.SetReadDeadline(time.Time{})
	if err != nil {
		return 0, err
	}
	return c.Reader.Read(b)
}

func (c *conn) Write(b []byte) (int, error) {
	err := c.link.SetWriteDeadline(time.Time{})
	if err != nil {
		return 0, err
	}
	return c.Writer.Write(b)
}

func (c *conn) Flush() error {
	err := c.link.SetWriteDeadline(time.Time{})
	if err != nil {
		return err
	}
	return c.Writer.Flush()
}

func (c *conn) Close() error {
	err1 := c.Flush()
	err2 := c.link.Close()
	if err1 != nil || err2 != nil {
		return fmt.Errorf("error occurred while closing connection: %v ; %v", err1, err2)
	}
	return nil
}

func (c *conn) Interrupt() error {
	return c.link.SetDeadline(time.Now())
}

func (c *conn) RemoteAddr() net.Addr {
	return c.link.RemoteAddr()
}
