package encrypt_test

import (
	"bytes"
	"encoding/binary"
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "gitlab.com/alephledger/consensus-go/pkg/crypto/encrypt"
)

var _ = Describe("Encryption", func() {

	var (
		ek  EncryptionKey
		dk  DecryptionKey
		ct  CipherText
		err error
	)

	Describe("small", func() {

		BeforeEach(func() {
			ek, dk, _ = GenerateKeys()
		})
		Describe("Checking CTEq", func() {
			var msg1, msg2 []byte
			var ct2 CipherText
			BeforeEach(func() {
				msg1 = make([]byte, 8)
				binary.LittleEndian.PutUint64(msg1, rand.Uint64())
				ct, err = ek.Encrypt(msg1)
				Expect(err).NotTo(HaveOccurred())

				msg2 = make([]byte, 8)
				binary.LittleEndian.PutUint64(msg2, rand.Uint64())
				ct2, err = ek.Encrypt(msg2)
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should check equalty correctly", func() {
				Expect(CTEq(ct, ct)).To(BeTrue())
				Expect(CTEq(ct, ct2)).To(BeFalse())
			})
		})

		Describe("Checking enc/dec", func() {

			var msg []byte

			BeforeEach(func() {
				msg = make([]byte, 8)
				binary.LittleEndian.PutUint64(msg, rand.Uint64())
				ct, err = ek.Encrypt(msg)
			})

			It("Should decrypt correctly", func() {
				Expect(err).To(BeNil())
				dmsg, err := dk.Decrypt(ct)
				Expect(err).To(BeNil())
				Expect(bytes.Equal(msg, dmsg)).To(BeTrue())
			})

			It("Should return false for forged ciphertext", func() {
				ct[0]++
				dmsg, err := dk.Decrypt(ct)
				Expect(err).NotTo(BeNil())
				Expect(bytes.Equal(msg, dmsg)).To(BeFalse())
			})
		})
		Describe("Checking encoding", func() {
			var (
				ekText string
				dkText string
			)
			BeforeEach(func() {
				ekText = ek.Encode()
				dkText = dk.Encode()
			})

			It("Should decode correctly", func() {
				ekd, err := NewEncryptionKey(ekText)
				Expect(err).To(BeNil())
				Expect(eqE(ek, ekd)).To(BeTrue())

				dkd, err := NewDecryptionKey(dkText)
				Expect(err).To(BeNil())
				Expect(eqD(dk, dkd)).To(BeTrue())
			})
			It("Should throw an error for malformed data", func() {
				ekText = "#" + ekText[1:]
				_, err := NewEncryptionKey(ekText)
				Expect(err).NotTo(BeNil())

				dkText = "#" + dkText[1:]
				_, err = NewDecryptionKey(dkText)
				Expect(err).NotTo(BeNil())
			})
		})
	})
})

func eqE(ek1, ek2 EncryptionKey) bool {
	return ek1.Encode() == ek2.Encode()
}

func eqD(dk1, dk2 DecryptionKey) bool {
	return dk1.Encode() == dk2.Encode()
}
