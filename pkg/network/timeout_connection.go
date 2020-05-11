package network

import (
	"crypto/rand"
	"fmt"
	"time"
)

type timeoutConn struct {
	Connection
	baseTimeout time.Duration
}

// newTimeoutConnection creates an instance of the Connection interface that implements a potentially infinitely
// long timing-out connection. In this implementation, each method has 50% chance of timing-out after (around) baseTimeout.
func newTimeoutConnection(baseTimeout time.Duration, connection Connection) Connection {
	return &timeoutConn{
		Connection:  connection,
		baseTimeout: baseTimeout,
	}
}

// It attempts to call Interrupt using the baseTimeout value as a hint. It implements it by the means of a simple random game
// that uses a fair coin. It starts with 1 as the initial stake and keeps playing till stake <= 0. After each `baseTimeout`
// ticks it toss a fair coin and based on it increases or decreases the current stake.
// Therefore, it calls Interrupt after `baseTimeout` with probability 1/2. It can be proved that it timeouts with probability 1
// and that expected value of the time after it timeouts is `infinity`.
func (tc *timeoutConn) tryTimeout(stopChan chan error) {
	var err error
	defer func() { stopChan <- err }()
	ticker := time.NewTicker(tc.baseTimeout)
	defer ticker.Stop()

	stake := uint64(1)
	var coin [1]byte
	coinToss := func() (bool, error) {
		_, err = rand.Read(coin[:])
		return coin[0]%2 == 0, err
	}
	for stake > 0 {

		select {
		case <-ticker.C:
			var head bool
			head, err = coinToss()
			if err != nil {
				return
			}
			if head {
				stake++
			} else {
				stake--
			}
		case <-stopChan:
			return
		}
	}

	select {
	case <-stopChan:
		return
	default:
	}

	err = tc.Interrupt()
	<-stopChan
}

func (tc *timeoutConn) Read(data []byte) (n int, err error) {
	stop := make(chan error)
	go tc.tryTimeout(stop)
	n, err = tc.Connection.Read(data)
	stop <- nil
	stopErr := <-stop

	if err != nil || stopErr != nil {
		err = fmt.Errorf("error occurred while calling Read on timeoutConn: %v ; %v", err, stopErr)
	}
	return
}

func (tc *timeoutConn) Write(data []byte) (n int, err error) {
	stop := make(chan error)
	go tc.tryTimeout(stop)
	n, err = tc.Connection.Write(data)
	stop <- nil
	stopErr := <-stop
	if err != nil || stopErr != nil {
		err = fmt.Errorf("error occurred while calling Write on timeoutConn: %v ; %v", err, stopErr)
	}
	return
}

func (tc *timeoutConn) Flush() (err error) {
	stop := make(chan error)
	go tc.tryTimeout(stop)
	err = tc.Connection.Flush()
	stop <- nil
	stopErr := <-stop
	if err != nil || stopErr != nil {
		err = fmt.Errorf("error occurred while calling Flush on timeoutConn: %v ; %v", err, stopErr)
	}
	return
}

func (tc *timeoutConn) Close() (err error) {
	err1 := tc.Flush()
	err2 := tc.Connection.Close()
	if err1 != nil || err2 != nil {
		err = fmt.Errorf("error occurred while calling Close on timeoutConn: %v ;  %v", err1, err2)
	}
	return
}

type timeoutConnectionServer struct {
	wrapped Server
	timeout time.Duration
}

// NewTimeoutConnectionServer creates an instance of a network Server that wrapes a given Server by creating instance of
// potentially infinitely long timing-out connections.
func NewTimeoutConnectionServer(server Server, baseTimeout time.Duration) Server {
	return &timeoutConnectionServer{wrapped: server, timeout: baseTimeout}
}

func (tcs *timeoutConnectionServer) Dial(pid uint16) (Connection, error) {
	con, err := tcs.wrapped.Dial(pid)
	if err != nil {
		return nil, err
	}
	return newTimeoutConnection(tcs.timeout, con), nil
}

func (tcs *timeoutConnectionServer) Listen() (Connection, error) {
	con, err := tcs.wrapped.Listen()
	if err != nil {
		return nil, err
	}
	return newTimeoutConnection(tcs.timeout, con), nil
}

func (tcs *timeoutConnectionServer) Stop() {
	tcs.wrapped.Stop()
}
