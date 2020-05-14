package network

import (
	"encoding/binary"
	"io"
)

// Greet sends a greeting to the given conn.
func Greet(w io.Writer, pid uint16, sid uint64) error {
	var data [10]byte
	binary.LittleEndian.PutUint16(data[0:], pid)
	binary.LittleEndian.PutUint64(data[2:], sid)
	_, err := w.Write(data[:])
	return err
}

// AcceptGreeting accepts a greeting and returns the information it learned from it.
func AcceptGreeting(conn io.Reader) (pid uint16, sid uint64, err error) {
	var data [10]byte
	_, err = io.ReadFull(conn, data[:])
	if err != nil {
		return
	}
	pid = binary.LittleEndian.Uint16(data[0:])
	sid = binary.LittleEndian.Uint64(data[2:])
	return
}
