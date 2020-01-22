package tests

import (
	"errors"
	"io"
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
	if err := c.in.CloseWithError(errors.New("")); err != nil {
		c.out.CloseWithError(errors.New(""))
		return err
	}
	return c.out.CloseWithError(errors.New(""))
}

func (c *connection) TimeoutAfter(time.Duration) {}

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
}

// Dial creates a new connection, pushes one end to the associated dial channel and return the other.
func (s *Server) Dial(k uint16, timeout time.Duration) (network.Connection, error) {
	out, in := NewConnection()
	if int(k) >= len(s.dialChans) {
		return nil, errors.New("unknown host")
	}
	select {
	case s.dialChans[k] <- in:
	case <-time.After(timeout):
		return nil, errors.New("Dial timeout")
	}
	return out, nil
}

// Listen picks up a connection from the listen channel
func (s *Server) Listen(timeout time.Duration) (network.Connection, error) {
	select {
	case conn, ok := <-s.listenChan:
		if !ok {
			return nil, errors.New("done")
		}
		return conn, nil
	case <-time.After(timeout):
	}
	return nil, errors.New("Listen timeout")
}

// CloseNetwork closes all the dial channels.
func CloseNetwork(servers []network.Server) {
	for _, ns := range servers {
		if s, ok := ns.(*Server); ok {
			close(s.listenChan)
		}
	}
}

// NewNetwork returns a slice of interconnected servers that simulate the network of the given size.
func NewNetwork(length int) []network.Server {
	channels := make([]chan network.Connection, length)
	for i := range channels {
		channels[i] = make(chan network.Connection)
	}
	servers := make([]network.Server, length)
	for i := range servers {
		servers[i] = &Server{
			dialChans:  channels,
			listenChan: channels[i]}
	}
	return servers
}
