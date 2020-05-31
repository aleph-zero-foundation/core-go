package network

import (
	"errors"
	"gitlab.com/alephledger/core-go/pkg/core"
)

// Hub represents a set of network servers used to communicate with peers
// Servers are distinguished by bands.
type Hub interface {
	core.Service
	// Dial connects to a committee member identified by pid and returns the resulting connection or an error.
	Dial(pid uint16, band int) (Connection, error)
	// Listen for an incoming connection for the given time. Returns the connection if successful, otherwise an error.
	Listen(band int) (Connection, error)
}

type hub struct {
	servers []Server
}

// NewHub creates a network hub from the given set of servers.
func NewHub(servers []Server) Hub {
	return &hub{servers}
}

// Start all servers
func (hb *hub) Start() error {
	for _, s := range hb.servers {
		err := s.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

// Stop all servers.
func (hb *hub) Stop() {
	for _, s := range hb.servers {
		s.Stop()
	}
}

func (hb *hub) Dial(pid uint16, band int) (Connection, error) {
	if band >= len(hb.servers) {
		return nil, errors.New("Invalid band")
	}
	return hb.servers[band].Dial(pid)
}

func (hb *hub) Listen(band int) (Connection, error) {
	if band >= len(hb.servers) {
		return nil, errors.New("Invalid band")
	}
	return hb.servers[band].Listen()
}
