package tss

import (
	"crypto/subtle"

	"gitlab.com/alephledger/core-go/pkg/crypto/bn256"
)

// VerifyShare verifies whether the given signature share is correct.
func (tk *ThresholdKey) VerifyShare(share *Share, msg []byte) bool {
	return tk.vks[share.owner].Verify(share.sgn, msg)
}

// VerifySignature verifies whether the given signature is correct.
func (tk *ThresholdKey) VerifySignature(s *Signature, msg []byte) bool {
	return tk.globalVK.Verify(s.sgn, msg)
}

// PolyVerify uses the given polyVerifier to verify if the verification keys form
// a polynomial sequence.
func (tk *ThresholdKey) PolyVerify(pv bn256.PolyVerifier) bool {
	return pv.Verify(tk.vks)
}

// VerifySecretKey checks if the verificationKey and secretKey form a valid pair.
// It returns the incorrect secret key when the pair of keys is invalid or
// nil when the keys are valid.
func (tk *ThresholdKey) VerifySecretKey() *bn256.SecretKey {
	vk := tk.sk.VerificationKey()
	if subtle.ConstantTimeCompare(vk.Marshal(), tk.vks[tk.owner].Marshal()) != 1 {
		return tk.sk
	}
	return nil
}
