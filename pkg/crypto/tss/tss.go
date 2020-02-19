// Package tss implements a threshold signature scheme.
package tss

import (
	"gitlab.com/alephledger/consensus-go/pkg/crypto/encrypt"
	"gitlab.com/alephledger/core-go/pkg/crypto/bn256"
)

// TSS is a set of all raw threshold keys generated by a dealer for all parties.
type TSS struct {
	dealer    uint16
	threshold uint16
	globalVK  *bn256.VerificationKey
	vks       []*bn256.VerificationKey
	sks       []*bn256.SecretKey
}

// ThresholdKey contains encrypted secretKeys of other parties
// and decrypted key of the owner.
type ThresholdKey struct {
	owner     uint16
	dealer    uint16
	threshold uint16
	globalVK  *bn256.VerificationKey
	vks       []*bn256.VerificationKey
	encSKs    []encrypt.CipherText
	sk        *bn256.SecretKey
}

// Share is a share of the coin owned by a process.
type Share struct {
	owner uint16
	sgn   *bn256.Signature
}

// Signature is a result of merging Shares.
type Coin struct {
	sgn *bn256.Signature
}

// Threshold returns the threshold of the given ThresholdCoin.
func (tk *ThresholdKey) Threshold() uint16 {
	return tk.threshold
}
