// Package persistent implements "virtual connections", many of which utilize the same underlying TCP link.
// Each virtual connection has a unique ID and every piece of data sent through the common TCP link is prefixed with a 12 bytes long header
// consisting of this ID (8 bytes) and the length of the piece of data (4 bytes).
//
// All writes are buffered and the actual network traffic happens only on Flush (which needs to be invoked explicitly) or when the buffer is full.
// Reads are also buffered and they read byte slices from the channel populated by the link supervising the connection.
// Close sends a header with data length set to 0. After closing the connection, calling Write or Flush returns an error, but reading is
// still possible until the underlying channel is depleted.
//
// NOTE: Write() and Flush() are NOT thread safe!
package persistent

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"sync/atomic"
	"time"
)

const (
	headerSize = 12
	// that way allocated slices are 16KB, which should suffice for units with ~15KB data
	bufSize = (1 << 14) - headerSize
)

func parseHeader(header []byte) (uint64, uint32) {
	return binary.LittleEndian.Uint64(header[:8]), binary.LittleEndian.Uint32(header[8:])
}

type chanReader struct {
	ch     chan []byte
	closed chan struct{}
}

func newChanReader(size int, closed chan struct{}) *chanReader {
	return &chanReader{ch: make(chan []byte, size), closed: closed}
}

func (cr *chanReader) Read(b []byte) (int, error) {
	select {
	case buf := <-cr.ch:
		return bytes.NewReader(buf).Read(b)
	case <-cr.closed:
		return 0, errors.New("Read on a closed connection")
	}
}

type conn struct {
	id      uint64
	link    *link
	queue   *chanReader
	reader  *bufio.Reader
	frame   []byte
	buffer  []byte
	bufLen  int
	sent    int
	recv    int
	closed  chan struct{}
	closing int64
}

// newConn creates a Connection with given id that wraps a tcp connection link
func newConn(id uint64, ln *link) *conn {
	closed := make(chan struct{})
	frame := make([]byte, headerSize+bufSize)
	binary.LittleEndian.PutUint64(frame, id)
	queue := newChanReader(32, closed)
	return &conn{
		id:     id,
		link:   ln,
		queue:  queue,
		reader: bufio.NewReaderSize(queue, bufSize),
		frame:  frame,
		buffer: frame[headerSize:],
		closed: closed,
	}
}

func (c *conn) Read(b []byte) (int, error) {
	return c.reader.Read(b)
}

func (c *conn) Write(b []byte) (int, error) {
	if c.isClosed() {
		return 0, errors.New("Write on a closed connection")
	}
	total := 0
	copied := copy(c.buffer[c.bufLen:], b)
	c.bufLen += copied
	total += copied
	for total < len(b) {
		err := c.Flush()
		if err != nil {
			return total, err
		}
		copied = copy(c.buffer[c.bufLen:], b[total:])
		c.bufLen += copied
		total += copied
	}
	return total, nil
}

func (c *conn) Flush() error {
	if c.isClosed() {
		return errors.New("Flush on a closed connection")
	}
	if c.bufLen == 0 {
		return nil
	}
	binary.LittleEndian.PutUint32(c.frame[8:], uint32(c.bufLen))
	_, err := c.link.tcpLink.Write(c.frame[:(headerSize + c.bufLen)])
	if err != nil {
		return err
	}
	c.sent += c.bufLen
	c.bufLen = 0
	return nil
}

func (c *conn) isClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

func (c *conn) Close() error {
	if c.isClosed() {
		return nil
	}
	if atomic.CompareAndSwapInt64(&c.closing, 0, 1) {
		err := c.SendFinished()
		if err != nil {
			return err
		}
		c.Finalize()
		c.erase()
	}
	return nil
}

func (c *conn) LocalClose() {
	if atomic.CompareAndSwapInt64(&c.closing, 0, 1) {
		c.Finalize()
		c.erase()
	}
}

func (c *conn) TimeoutAfter(t time.Duration) {
	if !c.link.IsDead() {
		c.link.tcpLink.SetDeadline(time.Now().Add(t))
	}
	go func() {
		time.Sleep(t)
		c.Close()
	}()
}

func (c *conn) RemoteAddr() net.Addr {
	return c.link.RemoteAddr()
}

func (c *conn) Enqueue(b []byte) {
	select {
	case c.queue.ch <- b:
		c.recv += len(b)
	case <-c.closed:
		return
	}
}

func (c *conn) SendFinished() error {
	header := make([]byte, headerSize)
	binary.LittleEndian.PutUint64(header, c.id)
	binary.LittleEndian.PutUint32(header[8:], 0)
	_, err := c.link.tcpLink.Write(header)
	return err
}

func (c *conn) Finalize() {
	close(c.closed)
}

func (c *conn) erase() {
	c.link.mx.Lock()
	defer c.link.mx.Unlock()
	delete(c.link.conns, c.id)
}
