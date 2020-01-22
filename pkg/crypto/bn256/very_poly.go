package bn256

import (
	"crypto/rand"
	"crypto/subtle"
	"math/big"
	"sync"

	"github.com/cloudflare/bn256"
)

// PolyVerifier is a struct which can verify if the given sequence is a polynomial sequence
// of bounded degree.
type PolyVerifier struct {
	vector []*big.Int
}

// Verify if the given sequence of elems is a polynomial sequence
// with bounded degree.
func (pv *PolyVerifier) Verify(elems []*VerificationKey) bool {
	if len(elems) != len(pv.vector) {
		// wrong number of elems
		return false
	}

	// computation of scalar product <pv.vector, elems>
	scalarProduct := new(bn256.G2)
	summands := make(chan *bn256.G2)

	var wg sync.WaitGroup
	for i, vk := range elems {
		wg.Add(1)
		go func(e *bn256.G2, i int) {
			defer wg.Done()
			summands <- new(bn256.G2).ScalarMult(e, pv.vector[i])
		}(&vk.key, i)
	}
	go func() {
		wg.Wait()
		close(summands)
	}()

	for summand := range summands {
		scalarProduct.Add(scalarProduct, summand)
	}

	// checking if the scalarProduct is the zero element of bn256.G2
	zeroMarshalled := new(bn256.G2).Marshal()
	return subtle.ConstantTimeCompare(zeroMarshalled, scalarProduct.Marshal()) == 1
}

// NewPolyVerifier returns a verifier of polynomial sequences
// of degree at most f and length n.
// We assume 0 <= f <= n-1.
func NewPolyVerifier(n, f int) PolyVerifier {
	// Here are some constants needed for computation of
	// the inverse of the Vandermonde's matrix V(1,2,...,n)
	// The constants depend only on n and should be big integers of length O(nlogn)
	// newton[i][j] is the newton symbol (i choose j)
	// sym[i][j] is the sum_{|S|=j, S in {1,2,...,i}} prod(S)
	// coeff[i][j] is the sum_{S in {1,2,...,n}\{i}, |S|=j} prod(S)
	newton := make([][]*big.Int, n+1)
	for i := 0; i <= n; i++ {
		newton[i] = make([]*big.Int, i+1)
		newton[i][0] = big.NewInt(int64(1))
		newton[i][i] = big.NewInt(int64(1))
		for j := 1; j < i; j++ {
			// newton[i][j] = newton[i-1][j] + newton[i-1][j-1]
			newton[i][j] = big.NewInt(0).Add(newton[i-1][j], newton[i-1][j-1])
		}
	}

	sym := make([][]*big.Int, n+1)
	for i := 0; i <= n; i++ {
		sym[i] = make([]*big.Int, i+1)
		sym[i][0] = big.NewInt(int64(1))
		for j := 1; j <= i; j++ {
			// sym[i][j] = i*sym[i-1][j-1] + sym[i-1][j]
			sym[i][j] = big.NewInt(int64(i))
			sym[i][j] = sym[i][j].Mul(sym[i][j], sym[i-1][j-1])
			if j <= i-1 {
				sym[i][j] = sym[i][j].Add(sym[i][j], sym[i-1][j])
			}
		}
	}

	coeff := make([][]*big.Int, n+1)
	for i := 1; i <= n; i++ {
		coeff[i] = make([]*big.Int, n)
		coeff[i][0] = big.NewInt(int64(1))
		for j := 1; j <= n-1; j++ {
			// coeff[i][j] = sym[n][j] - i*coeff[i][j-1]
			coeff[i][j] = big.NewInt(int64(-i))
			coeff[i][j] = coeff[i][j].Mul(coeff[i][j], coeff[i][j-1])
			coeff[i][j] = coeff[i][j].Add(coeff[i][j], sym[n][j])
		}
	}

	// invV is the inverse of Vandermode's matrix V(1,2,...,n)
	// after the following transformations
	// (1) signs are ignored (all numbers are positive)
	// (2) all numbers are multiplied by n!
	// (3) rows are in reverse order
	invV := make([][]*big.Int, n)
	for i := 0; i < n; i++ {
		invV[i] = make([]*big.Int, n)
		for j := 0; j < n; j++ {
			// invV[i][j] = newton[n-1][j] * coeff[j+1][i]
			invV[i][j] = big.NewInt(int64(1))
			invV[i][j] = invV[i][j].Mul(invV[i][j], newton[n-1][j])
			invV[i][j] = invV[i][j].Mul(invV[i][j], coeff[j+1][i])
		}
	}

	// Our magic vector is a random combination of last n-f-1 rows
	// of the inverse of Vandermonde's matrix V(1,2,...,n)
	magicVector := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		magicVector[i] = big.NewInt(int64(0))
	}
	for i := 0; i < n-f-1; i++ {
		scalar, _ := rand.Int(rand.Reader, bn256.Order)
		for j := 0; j < n; j++ {
			term := big.NewInt(int64(1))
			term.Mul(term, invV[i][j])
			term.Mul(term, scalar)
			if j%2 == 1 {
				// the inverse of the V(1,2,...,n) has a checkboard sign pattern
				term.Neg(term)
			}
			magicVector[j].Add(magicVector[j], term)
		}
	}
	for i := 0; i < n; i++ {
		magicVector[i].Mod(magicVector[i], bn256.Order)
	}
	return PolyVerifier{vector: magicVector}
}
