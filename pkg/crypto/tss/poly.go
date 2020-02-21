package tss

import (
	"math/big"

	"gitlab.com/alephledger/core-go/pkg/crypto/bn256"
)

func lagrange(points []int64, x int64) *big.Int {
	num := big.NewInt(int64(1))
	den := big.NewInt(int64(1))
	for _, p := range points {
		if p == x {
			continue
		}
		num.Mul(num, big.NewInt(0-p-1))
		den.Mul(den, big.NewInt(x-p))
	}
	den.ModInverse(den, bn256.Order)
	num.Mul(num, den)
	num.Mod(num, bn256.Order)
	return num
}

func poly(coeffs []*big.Int, x *big.Int) *big.Int {
	ans := big.NewInt(int64(0))
	for _, c := range coeffs {
		ans.Mul(ans, x)
		ans.Add(ans, c)
		ans.Mod(ans, bn256.Order)
	}
	return ans
}
