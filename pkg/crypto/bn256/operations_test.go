package bn256_test

import (
	"math/big"

	. "gitlab.com/alephledger/core-go/pkg/crypto/bn256"

	"github.com/cloudflare/bn256"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Operations", func() {
	var (
		vk1, vk2 *VerificationKey
		sk1, sk2 *SecretKey
	)
	BeforeEach(func() {
		var err error
		vk1, sk1, err = GenerateKeys()
		Expect(err).NotTo(HaveOccurred())
		vk2, sk2, err = GenerateKeys()
		Expect(err).NotTo(HaveOccurred())
	})
	Context("On keys and signatures", func() {
		Context("Addition", func() {
			It("Nil added to a key should return the key", func() {
				vk := AddVerificationKeys(nil, vk2)
				Expect(vk).To(Equal(vk2))
				sk := AddSecretKeys(nil, sk2)
				Expect(sk).To(Equal(sk2))
			})
			It("Sum of keys should produce a signature which is sum of signatures produced by single keys", func() {
				vk := AddVerificationKeys(vk1, vk2)
				sk := AddSecretKeys(sk1, sk2)
				data := []byte("asdf")
				Expect(AddSignatures(sk1.Sign(data), sk2.Sign(data)).Marshal()).To(Equal(sk.Sign(data).Marshal()))
				Expect(vk.Verify(sk.Sign(data), data)).To(BeTrue())
			})
		})
		Context("Multiplication", func() {
			Context("Signature multiplied by one", func() {
				It("Should be equal to the signature", func() {
					data := []byte("asdf")
					sgn := sk1.Sign(data)
					Expect(MulSignature(sgn, big.NewInt(1)).Marshal()).To(Equal(sgn.Marshal()))
				})
			})
			Context("Signature multiplied by two", func() {
				It("Should be equal to sum of two copies of the signature", func() {
					data := []byte("asdf")
					sgn := sk1.Sign(data)
					Expect(MulSignature(sgn, big.NewInt(2)).Marshal()).To(Equal(AddSignatures(sgn, sgn).Marshal()))
				})
			})
			Context("Nil signature multiplied by an integer", func() {
				It("Should return ", func() {
					n := big.NewInt(10)
					result := MulSignature(nil, n)
					expected := &Signature{*new(bn256.G1).ScalarBaseMult(n)}
					Expect(result.Marshal()).To(Equal(expected.Marshal()))
				})
			})
		})
	})
})
