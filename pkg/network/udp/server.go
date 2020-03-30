package udp

import (
	"net"
	"time"

	"gitlab.com/alephledger/core-go/pkg/network"
)

type server struct {
	listener    *net.UDPConn
	remoteAddrs []string
}

// NewServer initializes the network setup for the given local address and the set of remote addresses.
func NewServer(localAddress string, remoteAddresses []string) (network.Server, error) {
	local, err := net.ResolveUDPAddr("udp", localAddress)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenUDP("udp", local)
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
	buffer := make([]byte, (1 << 16))
	n, addr, err := s.listener.ReadFromUDP(buffer)
	if err != nil {
		return nil, err
	}
	conn := newConnIn(buffer[:n], addr)
	return conn, nil
}

func (s *server) Dial(pid uint16, timeout time.Duration) (network.Connection, error) {
	// can consider setting a timeout here, yet DialUDP is non-blocking, so there should be no need
	link, err := net.Dial("udp", s.remoteAddrs[pid])
	if err != nil {
		return nil, err
	}
	return newConnOut(link), nil
}
