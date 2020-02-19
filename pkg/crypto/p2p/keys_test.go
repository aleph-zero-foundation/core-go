package p2p_test

import (
	. "gitlab.com/alephledger/core-go/pkg/crypto/p2p"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Signing", func() {
	var (
		pk *PublicKey
		sk *SecretKey
	)
	BeforeEach(func() {
		var err error
		pk, sk, err = GenerateKeys()
		Expect(err).NotTo(HaveOccurred())
	})
	Context("A key when marshalled and unmarshalled", func() {
		It("Should return the key", func() {
			sk2 := new(SecretKey)
			_, err := sk2.Unmarshal(sk.Marshal())
			Expect(err).NotTo(HaveOccurred())
			Expect(sk2).To(Equal(sk))

			pk2 := new(PublicKey)
			_, err = pk2.Unmarshal(pk.Marshal())
			Expect(err).NotTo(HaveOccurred())
			Expect(pk.Marshal()).To(Equal(pk2.Marshal()))
		})
	})
	Context("When encoded and decoded", func() {
		It("Should return the same keys", func() {
			encPK := pk.Encode()
			decPK, err := DecodePublicKey(encPK)
			Expect(err).NotTo(HaveOccurred())
			Expect(pk.Marshal()).To(Equal(decPK.Marshal()))

			encSK := sk.Encode()
			decSK, err := DecodeSecretKey(encSK)
			Expect(err).NotTo(HaveOccurred())
			Expect(sk).To(Equal(decSK))
		})
	})
	Context("Verification of public key", func() {
		It("Should return true on a correct key", func() {
			Expect(pk.Verify()).To(BeTrue())
		})
	})
})
