// Package bn256 is a wrapper around github.com/cloudflare/bn256.
//
// In addition to generating and using keypairs for signing, it also contains functions
// needed to implement more involved cryptography, like threshold signatures and multisignatures.
package bn256

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"math/big"

	"github.com/cloudflare/bn256"
)

// VerificationKey can verify the validity of signatures.
type VerificationKey struct {
	key bn256.G2
}

// SecretKey can be used to sign data.
type SecretKey struct {
	key big.Int
}

// Signature confirms some information.
type Signature struct {
	bn256.G1
}

// Marshal the signature to bytes.
func (s *Signature) Marshal() []byte {
	return s.G1.Marshal()
}

// Unmarshal a signature from bytes.
func (s *Signature) Unmarshal(data []byte) (*Signature, error) {
	_, err := s.G1.Unmarshal(data)
	return s, err
}

var gen = new(bn256.G2).ScalarBaseMult(big.NewInt(int64(1)))

// GenerateKeys randomly.
func GenerateKeys() (*VerificationKey, *SecretKey, error) {
	secret, err := rand.Int(rand.Reader, Order)
	if err != nil {
		return nil, nil, err
	}
	sk := NewSecretKey(secret)
	return sk.VerificationKey(), sk, nil
}

// NewSecretKey returns a secret key with the specified secret.
func NewSecretKey(secret *big.Int) *SecretKey {
	return &SecretKey{
		key: *secret,
	}
}

// NewVerificationKey returns a verification key for the specified secret.
func NewVerificationKey(secret *big.Int) *VerificationKey {
	return &VerificationKey{
		key: *new(bn256.G2).ScalarBaseMult(secret),
	}
}

// Verify returns true if the provided signature is valid for msg.
func (vk *VerificationKey) Verify(s *Signature, msg []byte) bool {
	p1 := bn256.Pair(&s.G1, gen).Marshal()
	// hashing of the form msg => msg * gen is NOT secure
	p2 := bn256.Pair(hash(msg), &vk.key).Marshal()

	return subtle.ConstantTimeCompare(p1, p2) == 1
}

// Marshal the verification key.
func (vk *VerificationKey) Marshal() []byte {
	return vk.key.Marshal()
}

// Unmarshal the verification key.
func (vk *VerificationKey) Unmarshal(data []byte) (*VerificationKey, error) {
	_, err := vk.key.Unmarshal(data)
	return vk, err
}

// Sign returns a signature of msg.
func (sk *SecretKey) Sign(msg []byte) *Signature {
	return &Signature{*new(bn256.G1).ScalarMult(hash(msg), &sk.key)}
}

// Marshal the secret key.
func (sk *SecretKey) Marshal() []byte {
	return sk.key.Bytes()
}

// Unmarshal the secret key.
func (sk *SecretKey) Unmarshal(data []byte) (*SecretKey, error) {
	sk.key.SetBytes(data)
	return sk, nil
}

// VerificationKey returns the verification key associated with this secret key.
func (sk *SecretKey) VerificationKey() *VerificationKey {
	return NewVerificationKey(&sk.key)
}

// Encode encodes given SecretKey into a base64 string
func (sk *SecretKey) Encode() string {
	return base64.StdEncoding.EncodeToString(sk.Marshal())
}

// Encode encodes given VerificationKey into a base64 string
func (vk *VerificationKey) Encode() string {
	return base64.StdEncoding.EncodeToString(vk.Marshal())
}

// DecodeSecretKey decodes a secret key encoded as a base64 string.
func DecodeSecretKey(enc string) (*SecretKey, error) {
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return nil, err
	}
	sk, err := new(SecretKey).Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return sk, nil
}

// DecodeVerificationKey decodes a verification key encoded as a base64 string.
func DecodeVerificationKey(enc string) (*VerificationKey, error) {
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return nil, err
	}
	vk, err := new(VerificationKey).Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return vk, nil
}
