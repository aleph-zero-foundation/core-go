package rmc

import (
	"gitlab.com/alephledger/core-go/pkg/network"
)

// Greet sends a greeting to the given conn.
func Greet(conn network.Connection, pid uint16, id uint64, msgType byte) error {
	err := network.Greet(conn, pid, id)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte{msgType})
	if err != nil {
		return err
	}
	return conn.Flush()
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
