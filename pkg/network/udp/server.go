package udp

import (
	"context"
	"net"

	"github.com/rs/zerolog"

	"gitlab.com/alephledger/core-go/pkg/network"
)

type server struct {
	dialer      net.Dialer
	dialCtx     context.Context
	dialCancel  context.CancelFunc
	listener    *net.UDPConn
	localAddr   string
	remoteAddrs []string
	log         zerolog.Logger
}

// NewServer initializes the network setup for the given local address and the set of remote addresses.
func NewServer(localAddress string, remoteAddresses []string, log zerolog.Logger) network.Server {
	dialCtx, dialCancel := context.WithCancel(context.Background())
	return &server{
		localAddr:   localAddress,
		remoteAddrs: remoteAddresses,
		dialCtx:     dialCtx,
		dialCancel:  dialCancel,
		log:         log,
	}
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
	link, err := dial.DialContext(s.dialCtx, "udp", s.remoteAddrs[pid])
	if err != nil {
		return nil, err
	}
	return newConnOut(link), nil
}

func (s *server) Start() error {
	local, err := net.ResolveUDPAddr("udp", s.localAddr)
	if err != nil {
		return err
	}
	listener, err := net.ListenUDP("udp", local)
	if err != nil {
		return err
	}
	s.listener = listener
	return nil
}

func (s *server) Stop() {
	s.dialCancel()
	err := s.listener.Close()
	if err != nil {
		s.log.Err(err).Msg("error occurred while calling Close on the udp-listener")
	}
}
