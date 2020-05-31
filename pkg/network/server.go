// Package network defines abstractions used for handling network connections.
package network

import "gitlab.com/alephledger/core-go/pkg/core"

// Server establishes network connections.
type Server interface {
	core.Service
	// Dial connects to a committee member identified by pid and returns the resulting connection or an error.
	Dial(pid uint16) (Connection, error)
	// Listen for an incoming connection for the given time. Returns the connection if successful, otherwise an error.
	Listen() (Connection, error)
}
