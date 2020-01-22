package bn256

import (
	"math/big"

	"github.com/cloudflare/bn256"
	"golang.org/x/crypto/sha3"
)

func hash(msg []byte) *bn256.G1 {
	h := make([]byte, 32)
	sha3.ShakeSum128(h, msg)
	bInt := new(big.Int).SetBytes(h)
	// hashing of the form msg => msg * gen is NOT secure
	return new(bn256.G1).ScalarBaseMult(bInt)
}
