// Package network defines abstractions used for handling network connections.
package network

// Server establishes network connections.
type Server interface {
	// Dial connects to a committee member identified by pid and returns the resulting connection or an error.
	Dial(pid uint16) (Connection, error)
	// Listen for an incoming connection for the given time. Returns the connection if successful, otherwise an error.
	Listen() (Connection, error)
	// Stop stops this Server and cancels all ongoing executions of Dial and Listen.
	Stop()
}
