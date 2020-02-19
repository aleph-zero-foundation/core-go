// Package p2p implements functions for generating symmetric keys for peer to peer communication.
package p2p

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/cloudflare/bn256"
)

// SecretKey is a secret key used to generate p2p keys.
type SecretKey struct {
	key big.Int
}

// PublicKey is a public key used to generate p2p keys.
type PublicKey struct {
	g1 bn256.G1
	g2 bn256.G2
}

var genG1 = new(bn256.G1).ScalarBaseMult(big.NewInt(int64(1)))
var genG2 = new(bn256.G2).ScalarBaseMult(big.NewInt(int64(1)))

// GenerateKeys randomly.
func GenerateKeys() (*PublicKey, *SecretKey, error) {
	secret, err := rand.Int(rand.Reader, bn256.Order)
	if err != nil {
		return nil, nil, err
	}
	sk := NewSecretKey(secret)
	return sk.PublicKey(), sk, nil
}

// NewSecretKey returns a secret key with the specified secret.
func NewSecretKey(secret *big.Int) *SecretKey {
	return &SecretKey{
		key: *secret,
	}
}

// NewPublicKey returns a public key for the specified secret.
func NewPublicKey(secret *big.Int) *PublicKey {
	return &PublicKey{
		g1: *new(bn256.G1).ScalarBaseMult(secret),
		g2: *new(bn256.G2).ScalarBaseMult(secret),
	}
}

// Verify verifies if the public key is correct.
func (pk *PublicKey) Verify() bool {
	p1 := bn256.Pair(&pk.g1, genG2).Marshal()
	p2 := bn256.Pair(genG1, &pk.g2).Marshal()

	return subtle.ConstantTimeCompare(p1, p2) == 1
}

// Marshal the public key.
func (pk *PublicKey) Marshal() []byte {
	g1Marshalled := pk.g1.Marshal()
	g2Marshalled := pk.g2.Marshal()

	result := make([]byte, 4+len(g1Marshalled)+len(g2Marshalled))
	binary.LittleEndian.PutUint32(result[:4], uint32(len(g1Marshalled)))
	copy(result[4:], g1Marshalled)
	copy(result[(4+len(g1Marshalled)):], g2Marshalled)
	return result
}

// Unmarshal the public key.
func (pk *PublicKey) Unmarshal(data []byte) (*PublicKey, error) {
	if len(data) < 4 {
		return nil, errors.New("data too short")
	}
	g1Len := int(binary.LittleEndian.Uint32(data[:4]))

	if len(data) < 4+g1Len {
		return nil, errors.New("data too short")
	}

	_, err := pk.g1.Unmarshal(data[4:(4 + g1Len)])
	if err != nil {
		return nil, err
	}

	_, err = pk.g2.Unmarshal(data[(4 + g1Len):])
	if err != nil {
		return nil, err
	}

	return pk, nil
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

// PublicKey returns the public key associated with this secret key.
func (sk *SecretKey) PublicKey() *PublicKey {
	return NewPublicKey(&sk.key)
}

// Encode encodes given SecretKey into a base64 string.
func (sk *SecretKey) Encode() string {
	return base64.StdEncoding.EncodeToString(sk.Marshal())
}

// Encode encodes given PublicKey into a base64 string.
func (pk *PublicKey) Encode() string {
	return base64.StdEncoding.EncodeToString(pk.Marshal())
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

// DecodePublicKey decodes a verification key encoded as a base64 string.
func DecodePublicKey(enc string) (*PublicKey, error) {
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return nil, err
	}
	pk, err := new(PublicKey).Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return pk, nil
}
