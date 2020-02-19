package p2p_test

import (
	"gitlab.com/alephledger/consensus-go/pkg/crypto/encrypt"
	. "gitlab.com/alephledger/consensus-go/pkg/crypto/p2p"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Operations", func() {
	var (
		sk1, sk2 *SecretKey
		pk1, pk2 *PublicKey
	)
	BeforeEach(func() {
		var err error
		pk1, sk1, err = GenerateKeys()
		Expect(err).NotTo(HaveOccurred())
		pk2, sk2, err = GenerateKeys()
		Expect(err).NotTo(HaveOccurred())
	})
	Context("Shared elements", func() {
		var (
			e1, e2   SharedSecret
			k12, k21 encrypt.SymmetricKey
		)
		BeforeEach(func() {
			e1 = NewSharedSecret(sk1, pk2)
			e2 = NewSharedSecret(sk2, pk1)
			k12, _ = Key(e1)
			k21, _ = Key(e2)
		})
		It("Should agree on matching keys", func() {
			msg := []byte("asdf")
			msgEnc, _ := k12.Encrypt(msg)
			msgEncDec, _ := k21.Decrypt(msgEnc)
			Expect(msgEncDec).To(Equal(msg))
		})
		It("Should be verified correctly", func() {
			Expect(VerifySharedSecret(pk1, pk2, e1)).To(BeTrue())
			Expect(VerifySharedSecret(pk2, pk1, e1)).To(BeTrue())
		})
		It("Should not be verified correctly by other key", func() {
			pk3, _, err := GenerateKeys()
			Expect(err).NotTo(HaveOccurred())
			Expect(VerifySharedSecret(pk1, pk3, e1)).To(BeFalse())
		})
	})
})
