package tss

import (
	"encoding/binary"
	"errors"
	"sync"

	"gitlab.com/alephledger/core-go/pkg/crypto/bn256"
)

// Marshal returns byte representation of the given signature share in the following form
// (1) owner, 2 bytes as uint16
// (2) signature
func (sh *Share) Marshal() []byte {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data[:2], sh.owner)
	data = append(data, sh.sgn.Marshal()...)
	return data
}

// Unmarshal reads a signature share from its byte representation.
func (sh *Share) Unmarshal(data []byte) error {
	if len(data) < 2 {
		return errors.New("given data is too short")
	}
	owner := binary.LittleEndian.Uint16(data[:2])
	sgn := data[2:]
	sh.owner = owner
	decSgn, err := new(bn256.Signature).Unmarshal(sgn)
	if err != nil {
		return err
	}
	sh.sgn = decSgn
	return nil
}

// Marshal returns byte representation of the given signature.
func (s *Signature) Marshal() []byte {
	return s.sgn.Marshal()
}

// Unmarshal creates a signature from its byte representation.
func (s *Signature) Unmarshal(data []byte) error {
	if len(data) != bn256.SignatureLength {
		return errors.New("unmarshalling of signature failed. Wrong data length")
	}
	sgn := new(bn256.Signature)
	sgn, err := sgn.Unmarshal(data)
	if err != nil {
		return err
	}
	s.sgn = sgn
	return nil
}

// CreateShare creates a Share for given process and nonce.
func (tk *ThresholdKey) CreateShare(msg []byte) *Share {
	return &Share{
		owner: tk.owner,
		sgn:   tk.sk.Sign(msg),
	}
}

// CombineShares combines the given shares into a Signature.
// It returns a Signature and a bool value indicating whether the combining was successful or not.
func (tk *ThresholdKey) CombineShares(shares []*Share) (*Signature, bool) {
	if uint16(len(shares)) > tk.threshold {
		shares = shares[:tk.threshold]
	}
	if tk.threshold != uint16(len(shares)) {
		return nil, false
	}
	var points []int64
	for _, sh := range shares {
		points = append(points, int64(sh.owner))
	}

	summands := make(chan *bn256.Signature)

	var wg sync.WaitGroup
	for _, sh := range shares {
		wg.Add(1)
		go func(ch *Share) {
			defer wg.Done()
			summands <- bn256.MulSignature(ch.sgn, lagrange(points, int64(ch.owner)))
		}(sh)
	}
	go func() {
		wg.Wait()
		close(summands)
	}()

	var sum *bn256.Signature
	for elem := range summands {
		sum = bn256.AddSignatures(sum, elem)
	}

	return &Signature{sgn: sum}, true
}

// CreateShare creates a Share for given process and message if the holder of
// the weak threshold key is a share provider. Else it returns nil.
func (wtk *WeakThresholdKey) CreateShare(msg []byte) *Share {
	if !wtk.shareProviders[wtk.owner] {
		return nil
	}
	return &Share{
		owner: wtk.owner,
		sgn:   wtk.sk.Sign(msg),
	}
}
