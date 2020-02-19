package p2p

import (
	"crypto/subtle"

	"github.com/cloudflare/bn256"
	"gitlab.com/alephledger/consensus-go/pkg/crypto/encrypt"
)

// SharedSecret is a secret shared between two peers.
// It should be revealed, when proving that the other party
// has sent non compliant messages.
type SharedSecret struct {
	secret *bn256.G1
}

// Marshal the shared secret to bytes.
func (ss *SharedSecret) Marshal() []byte {
	return ss.secret.Marshal()
}

// Unmarshal the shared secret from bytes.
func (ss *SharedSecret) Unmarshal(data []byte) (*SharedSecret, error) {
	secret := new(bn256.G1)
	_, err := secret.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	ss.secret = secret
	return ss, nil
}

// NewSharedSecret returns a secret to share with the other party.
func NewSharedSecret(sk1 *SecretKey, pk2 *PublicKey) SharedSecret {
	return SharedSecret{new(bn256.G1).ScalarMult(&pk2.g1, &sk1.key)}
}

// VerifySharedSecret checks whether the shared element comes from the given keys.
func VerifySharedSecret(pk1, pk2 *PublicKey, elem SharedSecret) bool {
	p1 := bn256.Pair(elem.secret, genG2).Marshal()
	p2 := bn256.Pair(&pk1.g1, &pk2.g2).Marshal()
	return subtle.ConstantTimeCompare(p1, p2) == 1
}

// Key returns symmetric key for communication between the peers.
func Key(ss SharedSecret) (encrypt.SymmetricKey, error) {
	return encrypt.NewSymmetricKey(ss.secret.Marshal())
}

// Keys return symmetric keys for communication between the peers.
func Keys(sk1 *SecretKey, pks []*PublicKey, pid uint16) ([]encrypt.SymmetricKey, error) {
	nProc := uint16(len(pks))
	result := make([]encrypt.SymmetricKey, nProc)
	for i := uint16(0); i < nProc; i++ {
		sk, err := Key(NewSharedSecret(sk1, pks[i]))
		if err != nil {
			return nil, err
		}
		result[i] = sk
	}
	return result, nil
}
