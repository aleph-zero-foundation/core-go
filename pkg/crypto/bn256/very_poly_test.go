package bn256_test

import (
	. "gitlab.com/alephledger/core-go/pkg/crypto/bn256"

	"math/big"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VeryPoly", func() {
	var n, f int
	var pf PolyVerifier
	var values []*VerificationKey
	Context("Poly Verifier", func() {
		BeforeEach(func() {
			n, f = 10, 3
			pf = NewPolyVerifier(n, f)
		})
		Describe("On a sequence of constant values", func() {
			It("should return true", func() {
				values = make([]*VerificationKey, n)
				for i := range values {
					values[i] = NewVerificationKey(big.NewInt(2137))
				}
				Expect(pf.Verify(values)).To(BeTrue())
			})
		})
		Describe("On a sequence x^4 for x=1,2,...,n", func() {
			It("should return false", func() {
				values = make([]*VerificationKey, n)
				for i := range values {
					values[i] = NewVerificationKey(big.NewInt(int64(i * i * i * i)))
				}
				Expect(pf.Verify(values)).To(BeFalse())
			})
		})
		Describe("On a sequence x^3 for x=1,2,...,n", func() {
			It("should return true", func() {
				values = make([]*VerificationKey, n)
				for i := range values {
					values[i] = NewVerificationKey(big.NewInt(int64(i * i * i)))
				}
				Expect(pf.Verify(values)).To(BeTrue())
			})
		})
		Describe("On a sequence of values of some polynomial of degree 3", func() {
			It("should return true", func() {
				values = make([]*VerificationKey, n)
				for i := range values {
					values[i] = NewVerificationKey(big.NewInt(int64(3*i*i*i + 2*i*i + i + 7)))
				}
				Expect(pf.Verify(values)).To(BeTrue())
			})
		})
		Describe("On a sequence of values of some polynomial of degree 4", func() {
			It("should return false", func() {
				values = make([]*VerificationKey, n)
				for i := range values {
					values[i] = NewVerificationKey(big.NewInt(int64(2*i*i*i*i + 3*i*i*i + 2*i*i + i + 7)))
				}
				Expect(pf.Verify(values)).To(BeFalse())
			})
		})
	})
})
