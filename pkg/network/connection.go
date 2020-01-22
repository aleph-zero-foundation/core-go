package network

import (
	"time"
)

// Connection represents a network connection between two processes.
type Connection interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Flush() error
	Close() error
	TimeoutAfter(t time.Duration)
}
