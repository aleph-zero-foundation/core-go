package tcp

import (
	"github.com/rs/zerolog"

	"gitlab.com/alephledger/core-go/pkg/network"
)

// NewHub builds a hub of TCP network servers operating over provided set of addresses.
// Loggers to use with each server should be provided as logs  and should have the same length as addresses.
func NewHub(pid uint16, addresses [][]string, logs []zerolog.Logger) network.Hub {
	servers := make([]network.Server, len(addresses))
	for i := range servers {
		servers[i] = NewServer(addresses[i][pid], addresses[i], logs[i])
	}
	return network.NewHub(servers)

}
