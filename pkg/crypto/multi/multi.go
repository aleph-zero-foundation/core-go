// Package multi implements multisignatures on the bn256 curve.
//
// The kind of signatures we implement here is, in general, known to be vulnerable to an attack.
// The attack, however, requires choosing ones public keys based on the public keys of other participants.
// Fortunately, in our situation, we can use a simple protection against it.
// Committee candidates should submit a hash of the public key they are going to use,
// and reveal the public key only as they are elected.
//
// FOR SECURITY REASONS IT IS CRUCIAL THAT EITHER THE ABOVE OR SOME OTHER SOLUTION IS USED.
package multi

import (
	"encoding/binary"
	"errors"
	"sync"

	"gitlab.com/alephledger/core-go/pkg/crypto/bn256"
)

// Signature represents a multisignature associated with a piece of data and keychain.
type Signature struct {
	sync.Mutex
	threshold uint16
	data      []byte
	sgn       *bn256.Signature
	collected map[uint16]bool
}

// NewSignature creates a signature for the given data with the given threshold.
// At first it contains no partial signatures, they have to be aggregated.
func NewSignature(threshold uint16, data []byte) *Signature {
	return &Signature{
		threshold: threshold,
		data:      data,
		collected: map[uint16]bool{},
	}
}

// Aggregate the given signature together with other signatures we received.
// Returns true if the multisignature is complete.
// The signature should be verified earlier.
func (s *Signature) Aggregate(pid uint16, sgnBytes []byte) (bool, error) {
	sgn, err := new(bn256.Signature).Unmarshal(sgnBytes)
	s.Lock()
	defer s.Unlock()
	if s.complete() {
		return true, nil
	}
	if err != nil {
		return s.complete(), err
	}
	if s.collected[pid] {
		return s.complete(), errors.New("second copy of signature")
	}
	s.sgn = bn256.AddSignatures(s.sgn, sgn)
	s.collected[pid] = true
	return s.complete(), nil
}

// Marshal the multisignature to bytes.
// Only marshals the multisignature itself and the list of partial signatures included.
// Should only be called on complete proofs.
func (s *Signature) Marshal() []byte {
	s.Lock()
	defer s.Unlock()
	result := make([]byte, len(s.collected)*2)
	i := 0
	for c := range s.collected {
		binary.LittleEndian.PutUint16(result[i:i+2], uint16(c))
		i += 2
	}
	return append(result, s.sgn.Marshal()...)
}

// MarshaledLength returns how long would a marshaling of this proof be, in bytes.
func (s *Signature) MarshaledLength() int {
	return int(s.threshold)*2 + SignatureLength
}

// Unmarshal the multisignature from bytes.
// The receiver should contain the data and threshold that are the same as for the instance that was marshaled.
// If the unmarshaled signature is incorrect an error is returned.
func (s *Signature) Unmarshal(data []byte) (*Signature, error) {
	s.Lock()
	defer s.Unlock()
	s.collected = map[uint16]bool{}
	for i := 0; i < 2*int(s.threshold); i += 2 {
		c := binary.LittleEndian.Uint16(data[i : i+2])
		s.collected[c] = true
	}
	sgn, err := new(bn256.Signature).Unmarshal(data[2*s.threshold:])
	s.sgn = sgn
	return s, err
}

func (s *Signature) complete() bool {
	return len(s.collected) >= int(s.threshold)
}
