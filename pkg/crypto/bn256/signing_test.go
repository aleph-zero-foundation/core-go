package bn256_test

import (
	. "gitlab.com/alephledger/core-go/pkg/crypto/bn256"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Signing", func() {
	var (
		pub  *VerificationKey
		priv *SecretKey
	)
	BeforeEach(func() {
		var err error
		pub, priv, err = GenerateKeys()
		Expect(err).NotTo(HaveOccurred())
	})
	Context("Some data", func() {
		var data []byte
		BeforeEach(func() {
			data = []byte("19890604")
		})
		Describe("When signed", func() {
			var signature *Signature
			BeforeEach(func() {
				signature = priv.Sign(data)
			})
			It("should be successfully verified", func() {
				Expect(pub.Verify(signature, data)).To(BeTrue())
			})
			It("should be successfully verified after marshaling and unmarshaling the signature", func() {
				sgn, err := new(Signature).Unmarshal(signature.Marshal())
				Expect(err).NotTo(HaveOccurred())
				Expect(pub.Verify(sgn, data)).To(BeTrue())
			})
			It("should fail for different data", func() {
				Expect(pub.Verify(signature, []byte("19890535"))).To(BeFalse())
			})
		})
	})
	Context("A key when marshalled and unmarshalled", func() {
		It("Should return the key", func() {
			priv2 := new(SecretKey)
			_, err := priv2.Unmarshal(priv.Marshal())
			Expect(err).NotTo(HaveOccurred())
			Expect(priv2).To(Equal(priv))

			pub2 := new(VerificationKey)
			_, err = pub2.Unmarshal(pub.Marshal())
			Expect(err).NotTo(HaveOccurred())
			// The public keys cannot be compared using Equal,
			// because the same public key (element of G1) can be represented
			// in different ways in memory.
			Expect(pub.Marshal()).To(Equal(pub2.Marshal()))
		})
	})
})
