package bn256

import (
	"github.com/cloudflare/bn256"
)

func hash(msg []byte) *bn256.G1 {
	return bn256.HashG1(msg, []byte("az-sig"))
}
