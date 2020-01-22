// Package tcp implements network.Connections that wrap around TCP connections.
package tcp

import (
	"bufio"
	"net"
	"time"

	"gitlab.com/alephledger/core-go/pkg/network"
)

const (
	bufSize = 32000
)

type conn struct {
	link   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	sent   int
	recv   int
}

// newConn creates a Connection object wrapping a particular tcp connection link
func newConn(link net.Conn) network.Connection {
	return &conn{
		link:   link,
		reader: bufio.NewReaderSize(link, bufSize),
		writer: bufio.NewWriterSize(link, bufSize),
	}
}

func (c *conn) Read(b []byte) (int, error) {
	n, err := c.reader.Read(b)
	c.recv += n
	return n, err
}

func (c *conn) Write(b []byte) (int, error) {
	written, n := 0, 0
	var err error
	for written < len(b) {
		n, err = c.writer.Write(b[written:])
		written += n
		if err == bufio.ErrBufferFull {
			err = c.writer.Flush()
		}
		if err != nil {
			break
		}
	}
	c.sent += written
	return written, err
}

func (c *conn) Flush() error {
	return c.writer.Flush()
}

func (c *conn) Close() error {
	err := c.link.Close()
	return err
}

func (c *conn) TimeoutAfter(t time.Duration) {
	c.link.SetDeadline(time.Now().Add(t))
}
