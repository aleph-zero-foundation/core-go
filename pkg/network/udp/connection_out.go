// Package udp wraps UDP packets in network.Connections.
//
// Since UDP packets are not real connections the abstraction is not an exact fit.
// Due to that, the connections are either read- or write-only and have other restrictions.
// In particular the write-only connection allows to write only up to 65 507 bytes,
// while the read-only reads the data from only one packet.
package udp

import (
	"errors"
	"net"
	"time"

	"gitlab.com/alephledger/core-go/pkg/network"
)

const udpMaxPacketSize = (1 << 16) - 512

type connOut struct {
	link        net.Conn
	writeBuffer []byte
	sent        int
}

// newConnOut initializes an outgoing UDP "connection"
func newConnOut(link net.Conn) network.Connection {
	return &connOut{
		link:        link,
		writeBuffer: make([]byte, 0),
	}
}

func (c *connOut) Read(b []byte) (int, error) {
	return 0, errors.New("cannot read from outgoing UDP connection")
}

func (c *connOut) Write(b []byte) (int, error) {
	if len(c.writeBuffer)+len(b) > udpMaxPacketSize {
		return 0, errors.New("cannot write as the message length would exceed 65024, did you forget to Flush()?")
	}
	c.writeBuffer = append(c.writeBuffer, b...)
	return len(b), nil
}

func (c *connOut) Flush() error {
	_, err := c.link.Write(c.writeBuffer)
	c.sent += len(c.writeBuffer)
	c.writeBuffer = make([]byte, 0)
	return err
}

func (c *connOut) Close() error {
	err := c.link.Close()
	return err
}

func (c *connOut) TimeoutAfter(t time.Duration) {
	c.link.SetDeadline(time.Now().Add(t))
}

func (c *connOut) RemoteAddr() net.Addr {
	return c.link.RemoteAddr()
}
