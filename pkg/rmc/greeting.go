package rmc

import (
	"encoding/binary"
	"io"

	"gitlab.com/alephledger/core-go/pkg/network"
)

// Greet sends a greeting to the given conn.
func Greet(conn network.Connection, pid uint16, id uint64, msgType byte) error {
	var data [11]byte
	binary.LittleEndian.PutUint16(data[0:], pid)
	binary.LittleEndian.PutUint64(data[2:], id)
	data[10] = msgType
	_, err := conn.Write(data[:])
	if err != nil {
		return err
	}
	return conn.Flush()
}

// AcceptGreeting accepts a greeting and returns the information it learned from it.
func AcceptGreeting(conn network.Connection) (pid uint16, id uint64, msgType byte, err error) {
	var data [11]byte
	_, err = io.ReadFull(conn, data[:])
	if err != nil {
		return
	}
	pid = binary.LittleEndian.Uint16(data[0:])
	id = binary.LittleEndian.Uint64(data[2:])
	msgType = data[10]
	return
}
