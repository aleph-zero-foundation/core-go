package tcp

import (
	"net"
	"time"

	"gitlab.com/alephledger/core-go/pkg/network"
)

type server struct {
	listener    *net.TCPListener
	remoteAddrs []string
}

// NewServer initializes the network setup for the given local address and the set of remote addresses.
func NewServer(localAddress string, remoteAddresses []string) (network.Server, error) {
	local, err := net.ResolveTCPAddr("tcp", localAddress)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp", local)
	if err != nil {
		return nil, err
	}
	return &server{
		listener:    listener,
		remoteAddrs: remoteAddresses,
	}, nil
}

func (s *server) Listen(timeout time.Duration) (network.Connection, error) {
	s.listener.SetDeadline(time.Now().Add(timeout))
	link, err := s.listener.Accept()
	if err != nil {
		return nil, err
	}
	conn := newConn(link)
	return conn, nil
}

func (s *server) Dial(pid uint16, timeout time.Duration) (network.Connection, error) {
	link, err := net.DialTimeout("tcp", s.remoteAddrs[pid], timeout)
	if err != nil {
		return nil, err
	}
	return newConn(link), nil
}
