package udp

import (
	"bytes"
	"errors"
	"io"
	"time"

	"gitlab.com/alephledger/core-go/pkg/network"
)

type connIn struct {
	reader io.Reader
	recv   int
}

// newConnIn initializes an incoming UDP "connection" -- wrapping the content of the incoming packet
func newConnIn(packet []byte) network.Connection {
	return &connIn{
		reader: bytes.NewReader(packet),
	}
}

func (c *connIn) Read(b []byte) (int, error) {
	n, err := c.reader.Read(b)
	c.recv += n
	return n, err
}

func (c *connIn) Write(b []byte) (int, error) {
	return 0, errors.New("cannot write to incoming UDP connection")
}

func (c *connIn) Flush() error {
	return errors.New("cannot flush incoming UDP connection")
}

func (c *connIn) Close() error {
	return nil
}

func (c *connIn) TimeoutAfter(t time.Duration) {
	// does nothing as the UDP connIn is non-blocking anyway
}
