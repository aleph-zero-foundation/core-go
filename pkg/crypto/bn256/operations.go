package bn256

import (
	"crypto/subtle"
	"math/big"

	"github.com/cloudflare/bn256"
)

// AddVerificationKeys returns a sum of the provided verification keys.
// If the first argument is nil, it treats it as a zero.
func AddVerificationKeys(vk1, vk2 *VerificationKey) *VerificationKey {
	if vk1 == nil {
		return vk2
	}
	return &VerificationKey{
		key: *new(bn256.G2).Add(&vk1.key, &vk2.key),
	}
}

// AddSecretKeys returns a sum of the provided secret keys.
// If the first argument is nil, it treats it as a zero.
func AddSecretKeys(sk1, sk2 *SecretKey) *SecretKey {
	if sk1 == nil {
		return sk2
	}
	result := new(big.Int).Add(&sk1.key, &sk2.key)
	result = result.Mod(result, bn256.Order)
	return &SecretKey{
		key: *result,
	}
}

// AddSignatures returns a sum of the provided signatures.
// If the first argument is nil, it treats it as a zero.
func AddSignatures(sgn1, sgn2 *Signature) *Signature {
	if sgn1 == nil {
		return sgn2
	}
	result := new(bn256.G1).Add(&sgn1.G1, &sgn2.G1)
	return &Signature{
		*result,
	}
}

// MulSignature returns the provided signature multiplied by the integer.
// If the first argument is nil, it treats it as a one.
func MulSignature(sgn *Signature, n *big.Int) *Signature {
	if sgn == nil {
		return &Signature{*new(bn256.G1).ScalarBaseMult(n)}
	}
	result := new(bn256.G1).ScalarMult(&sgn.G1, n)
	return &Signature{
		*result,
	}
}

// VerifyKeys checks whether given secretKey and verificationKey forms a vaild pair.
func VerifyKeys(vk *VerificationKey, sk *SecretKey) bool {
	vk2 := sk.VerificationKey()
	return subtle.ConstantTimeCompare(vk.Marshal(), vk2.Marshal()) == 1
}
