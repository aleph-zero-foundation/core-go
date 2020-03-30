// Package rmc implements a reliable multicast for arbitrary data.
//
// This protocol is based on RBC (reliable broadcast), but has slightly different guarantees.
// Crucially a piece of data multicast with RMC with a given id will agree among all who received it, i.e. it is unique.
// The protocol has no hard guarantees pertaining pessimistic message complexity,
// but can be used in tandem with gossip protocols to disseminate data with proofs of uniqueness.
package rmc

import (
	"errors"
	"io"
	"sync"

	"gitlab.com/alephledger/core-go/pkg/crypto/bn256"
	"gitlab.com/alephledger/core-go/pkg/crypto/multi"
)

// RMC is a structure holding all data related to a series of reliable multicasts.
type RMC struct {
	inMx, outMx sync.RWMutex
	keys        *multi.Keychain
	in          map[uint64]*incoming
	out         map[uint64]*instance
}

// New creates a context for executing instances of the reliable multicast.
func New(pubs []*bn256.VerificationKey, priv *bn256.SecretKey) *RMC {
	return &RMC{
		keys: multi.NewKeychain(pubs, priv),
		in:   map[uint64]*incoming{},
		out:  map[uint64]*instance{},
	}
}

// InitiateRaw data signature gathering.
// This should be used only when all participants already know the data,
// and only want to produce a proof that it is agreed between them.
func (rmc *RMC) InitiateRaw(id uint64, data []byte) error {
	_, err := rmc.newRawInstance(id, data)
	return err
}

// AcceptData reads the id from r, followed by the data and signature of the whole thing.
// It verifies that the id matches the provided one, and that the signature was made by pid.
// It returns the data itself, for protocol-independent verification.
func (rmc *RMC) AcceptData(id uint64, pid uint16, r io.Reader) ([]byte, error) {
	in, err := rmc.newIncomingInstance(id, pid)
	if err != nil {
		return nil, err
	}
	return in.AcceptData(r)
}

// SendSignature writes the signature associated with id to w.
// The signature signs the data.
func (rmc *RMC) SendSignature(id uint64, w io.Writer) error {
	ins, err := rmc.get(id)
	if err != nil {
		return err
	}
	return ins.SendSignature(w)
}

// AcceptProof reads a proof from r and verifies it is a proof that id succeeded.
func (rmc *RMC) AcceptProof(id uint64, r io.Reader) error {
	ins, err := rmc.get(id)
	if err != nil {
		return err
	}
	return ins.AcceptProof(r)
}

// SendData writes data concatenated with the id and signed by us to w.
func (rmc *RMC) SendData(id uint64, data []byte, w io.Writer) error {
	if rmc.Status(id) != Unknown {
		out, err := rmc.getOut(id)
		if err != nil {
			return err
		}
		return out.SendData(w)
	}
	out := rmc.newOutgoingInstance(id, data)
	return out.SendData(w)
}

// AcceptSignature reads a signature from r and verifies it represents pid signing the data associated with id.
// It returns true when the signature is exactly threshold-th signature gathered.
func (rmc *RMC) AcceptSignature(id uint64, pid uint16, r io.Reader) (bool, error) {
	ins, err := rmc.get(id)
	if err != nil {
		return false, err
	}
	return ins.AcceptSignature(pid, r)
}

// SendProof writes the proof associated with id to w.
func (rmc *RMC) SendProof(id uint64, w io.Writer) error {
	ins, err := rmc.get(id)
	if err != nil {
		return err
	}
	return ins.SendProof(w)
}

// SendFinished writes the data and proof associated with id to w.
func (rmc *RMC) SendFinished(id uint64, w io.Writer) error {
	ins, err := rmc.get(id)
	if err != nil {
		return err
	}
	return ins.SendFinished(w)
}

// AcceptFinished reads a pair of data and proof from r and verifies it corresponds to a successfully finished RMC.
func (rmc *RMC) AcceptFinished(id uint64, pid uint16, r io.Reader) ([]byte, error) {
	in, err := rmc.getIn(id)
	if err != nil {
		in, _ = rmc.newIncomingInstance(id, pid)
	}
	return in.AcceptFinished(r)
}

// Status returns the state corresponding to id.
func (rmc *RMC) Status(id uint64) Status {
	ins, err := rmc.get(id)
	if err != nil {
		return Unknown
	}
	return ins.Status()
}

// Data returns the raw data corresponding to id.
// If the status differs from Finished, this data might be unreliable!
func (rmc *RMC) Data(id uint64) []byte {
	ins, err := rmc.get(id)
	if err != nil {
		return nil
	}
	return ins.Data()
}

// Proof returns the proof corresponding to id.
// If the status differs from Finished, returns nil.
func (rmc *RMC) Proof(id uint64) *multi.Signature {
	ins, err := rmc.get(id)
	if err != nil {
		return nil
	}
	return ins.proof
}

// Clear removes all information concerning id.
// After a clear the state is Unknown until any further calls with id.
func (rmc *RMC) Clear(id uint64) {
	rmc.inMx.Lock()
	defer rmc.inMx.Unlock()
	rmc.outMx.Lock()
	defer rmc.outMx.Unlock()
	delete(rmc.in, id)
	delete(rmc.out, id)
}

func (rmc *RMC) newIncomingInstance(id uint64, pid uint16) (*incoming, error) {
	result := newIncoming(id, pid, rmc.keys)
	rmc.inMx.Lock()
	defer rmc.inMx.Unlock()
	if result, ok := rmc.in[id]; ok {
		return result, errors.New("duplicate incoming")
	}
	rmc.in[id] = result
	return result, nil
}

func (rmc *RMC) newOutgoingInstance(id uint64, data []byte) *instance {
	result := newOutgoing(id, data, rmc.keys)
	rmc.outMx.Lock()
	defer rmc.outMx.Unlock()
	if result, ok := rmc.out[id]; ok {
		return result
	}
	rmc.out[id] = result
	return result
}

func (rmc *RMC) newRawInstance(id uint64, data []byte) (*instance, error) {
	result := newRaw(id, data, rmc.keys)
	rmc.outMx.Lock()
	defer rmc.outMx.Unlock()
	if result, ok := rmc.out[id]; ok {
		return result, errors.New("duplicate raw")
	}
	rmc.out[id] = result
	return result, nil
}

func (rmc *RMC) getIn(id uint64) (*incoming, error) {
	rmc.inMx.RLock()
	defer rmc.inMx.RUnlock()
	result, ok := rmc.in[id]
	if !ok {
		return nil, errors.New("unknown incoming")
	}
	return result, nil
}

func (rmc *RMC) getOut(id uint64) (*instance, error) {
	rmc.outMx.RLock()
	defer rmc.outMx.RUnlock()
	result, ok := rmc.out[id]
	if !ok {
		return nil, errors.New("unknown outgoing")
	}
	return result, nil
}

func (rmc *RMC) get(id uint64) (*instance, error) {
	if in, err := rmc.getIn(id); err == nil {
		return &in.instance, nil
	}
	if out, err := rmc.getOut(id); err == nil {
		return out, nil
	}
	return nil, errors.New("unknown instance")
}
