package udp

import (
	"context"
	"net"

	"github.com/rs/zerolog"

	"gitlab.com/alephledger/core-go/pkg/network"
)

type server struct {
	dialer      net.Dialer
	ctx         context.Context
	cancel      context.CancelFunc
	listener    *net.UDPConn
	remoteAddrs []string
	log         zerolog.Logger
}

// NewServer initializes the network setup for the given local address and the set of remote addresses.
func NewServer(localAddress string, remoteAddresses []string, log zerolog.Logger) (network.Server, error) {
	local, err := net.ResolveUDPAddr("udp", localAddress)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenUDP("udp", local)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &server{
		listener:    listener,
		remoteAddrs: remoteAddresses,
		ctx:         ctx,
		cancel:      cancel,
		log:         log,
	}, nil
}

func (s *server) Listen() (network.Connection, error) {
	buffer := make([]byte, (1 << 16))
	n, addr, err := s.listener.ReadFromUDP(buffer)
	if err != nil {
		return nil, err
	}
	conn := newConnIn(buffer[:n], addr)
	return conn, nil
}

func (s *server) Dial(pid uint16) (network.Connection, error) {
	var dial net.Dialer
	link, err := dial.DialContext(s.ctx, "udp", s.remoteAddrs[pid])
	if err != nil {
		return nil, err
	}
	return newConnOut(link), nil
}

func (s *server) Stop() {
	s.cancel()
	err := s.listener.Close()
	if err != nil {
		s.log.Err(err).Msg("error occurred while calling Close on the udp-listener")
	}
}
