package tests

import (
	"errors"
	"io"
	"net"
	"time"

	"gitlab.com/alephledger/core-go/pkg/network"
)

type connection struct {
	in  *io.PipeReader
	out *io.PipeWriter
}

func (c *connection) Read(buf []byte) (int, error) {
	return c.in.Read(buf)
}

func (c *connection) Write(buf []byte) (int, error) {
	return c.out.Write(buf)
}

func (c *connection) Flush() error {
	return nil
}

func (c *connection) Close() error {
	if err := c.in.CloseWithError(nil); err != nil {
		c.out.CloseWithError(nil)
		return err
	}
	return c.out.CloseWithError(nil)
}

func (c *connection) TimeoutAfter(time.Duration) {}

func (c *connection) Interrupt() error { return nil }

func (c *connection) RemoteAddr() net.Addr { return nil }

// NewConnection creates a pipe simulating a pair of network connections.
func NewConnection() (network.Connection, network.Connection) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()
	return &connection{r1, w2}, &connection{r2, w1}
}

// Server implements network.Server. Afterwards it needs to be closed with Close().
type Server struct {
	dialChans  []chan network.Connection
	listenChan chan network.Connection
	timeout    time.Duration
}

// Dial creates a new connection, pushes one end to the associated dial channel and return the other.
func (s *Server) Dial(k uint16) (network.Connection, error) {
	out, in := NewConnection()
	if int(k) >= len(s.dialChans) {
		return nil, errors.New("unknown host")
	}
	select {
	case s.dialChans[k] <- in:
	case <-time.After(s.timeout):
		return nil, errors.New("Dial timeout")
	}
	return out, nil
}

// Listen picks up a connection from the listen channel
func (s *Server) Listen() (network.Connection, error) {
	select {
	case conn, ok := <-s.listenChan:
		if !ok {
			return nil, errors.New("done")
		}
		return conn, nil
	case <-time.After(s.timeout):
	}
	return nil, errors.New("Listen timeout")
}

// Start mock
func (s *Server) Start() error { return nil }

// Stop this, linter!
func (s *Server) Stop() {}

// CloseNetwork closes all the dial channels.
func CloseNetwork(servers []network.Server) {
	for _, ns := range servers {
		if s, ok := ns.(*Server); ok {
			close(s.listenChan)
		}
	}
}

// NewNetwork returns a slice of interconnected servers that simulate the network of the given size.
func NewNetwork(length int, timeout time.Duration) []network.Server {
	channels := make([]chan network.Connection, length)
	for i := range channels {
		channels[i] = make(chan network.Connection)
	}
	servers := make([]network.Server, length)
	for i := range servers {
		servers[i] = &Server{
			dialChans:  channels,
			listenChan: channels[i],
			timeout:    timeout,
		}
	}
	return servers
}
