package rmc

import (
	"io"

	"gitlab.com/alephledger/core-go/pkg/network"
)

// Greet sends a greeting to the given conn.
func Greet(w io.Writer, pid uint16, id uint64, msgType byte) error {
	err := network.Greet(w, pid, id)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte{msgType})
	return err
}

// AcceptGreeting accepts a greeting and returns the information it learned from it.
func AcceptGreeting(conn network.Connection) (pid uint16, id uint64, msgType byte, err error) {
	pid, id, err = network.AcceptGreeting(conn)
	if err != nil {
		return
	}
	var data [1]byte
	_, err = conn.Read(data[:])
	msgType = data[0]
	return
}
