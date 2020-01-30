// Package network defines abstractions used for handling network connections.
package network

import "time"

// Server establishes network connections.
type Server interface {
	// Dial connects to a committee member identified by pid and returns the resulting connection or an error.
	Dial(pid uint16, timeout time.Duration) (Connection, error)
	// Listen for an incoming connection for the given time. Returns the connection if successful, otherwise an error.
	Listen(timeout time.Duration) (Connection, error)
}
