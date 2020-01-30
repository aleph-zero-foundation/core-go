package multi

import (
	"crypto/subtle"

	"gitlab.com/alephledger/core-go/pkg/crypto/bn256"
)

// SignatureLength is the length of signatures created by this package.
const SignatureLength = bn256.SignatureLength

// Keychain represents the set of keys used for the multisigning procedure.
type Keychain struct {
	pubs []*bn256.VerificationKey
	priv *bn256.SecretKey
	pid  uint16
}

// NewKeychain creates a new keychain using the provided keys.
func NewKeychain(pubs []*bn256.VerificationKey, priv *bn256.SecretKey) *Keychain {
	ourPub := priv.VerificationKey().Marshal()
	var pid uint16
	for id, p := range pubs {
		if subtle.ConstantTimeCompare(p.Marshal(), ourPub) == 1 {
			pid = uint16(id)
			break
		}
	}
	return &Keychain{
		pubs: pubs,
		priv: priv,
		pid:  pid,
	}
}

// Verify checks whether the slice of bytes consists of some data followed by a correct signature by pid.
func (k *Keychain) Verify(pid uint16, data []byte) bool {
	if len(data) < SignatureLength {
		return false
	}
	dataEnd := len(data) - SignatureLength
	signature, err := new(bn256.Signature).Unmarshal(data[dataEnd:])
	if err != nil {
		return false
	}
	return k.pubs[pid].Verify(signature, data[:dataEnd])
}

// Sign returns a signature for the provided data.
func (k *Keychain) Sign(data []byte) []byte {
	return k.priv.Sign(data).Marshal()
}

// MultiVerify verifies whether the provided multisignature contains correctly signed data.
func (k *Keychain) MultiVerify(s *Signature) bool {
	if !s.complete() {
		return false
	}
	var multiKey *bn256.VerificationKey
	for c := range s.collected {
		multiKey = bn256.AddVerificationKeys(multiKey, k.pubs[c])
	}
	return multiKey.Verify(s.sgn, s.data)
}

// Pid of the owner of the private key on this keychain.
func (k *Keychain) Pid() uint16 {
	return k.pid
}

// Length of the keychain, i.e. how many public keys there are.
func (k *Keychain) Length() uint16 {
	return uint16(len(k.pubs))
}
