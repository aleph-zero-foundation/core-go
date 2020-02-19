package tss

import (
	"crypto/subtle"

	"gitlab.com/alephledger/core-go/pkg/crypto/bn256"
)

// VerifyShare verifies whether the given signature share is correct.
func (ts *ThresholdKey) VerifyShare(share *Share, msg []byte) bool {
	return ts.vks[share.owner].Verify(share.sgn, msg)
}

// VerifySignature verifies whether the given signature is correct.
func (ts *ThresholdKey) VerifySignature(s *Signature, msg []byte) bool {
	return ts.globalVK.Verify(s.sgn, msg)
}

// PolyVerify uses the given polyVerifier to verify if the verification keys form
// a polynomial sequence.
func (ts *ThresholdKey) PolyVerify(pv bn256.PolyVerifier) bool {
	return pv.Verify(ts.vks)
}

// VerifySecretKey checks if the verificationKey and secretKey form a valid pair.
// It returns the incorrect secret key when the pair of keys is invalid or
// nil when the keys are valid.
func (ts *ThresholdKey) VerifySecretKey() *bn256.SecretKey {
	vk := ts.sk.VerificationKey()
	if subtle.ConstantTimeCompare(vk.Marshal(), ts.vks[ts.owner].Marshal()) != 1 {
		return ts.sk
	}
	return nil
}
