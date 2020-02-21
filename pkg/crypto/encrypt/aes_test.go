package encrypt_test

import (
	"bytes"
	"encoding/binary"
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "gitlab.com/alephledger/core-go/pkg/crypto/encrypt"
)

var _ = Describe("Encryption", func() {

	var (
		sk  SymmetricKey
		ct  CipherText
		err error
	)

	Describe("small", func() {

		BeforeEach(func() {
			sk, err = NewSymmetricKey([]byte("2137"))
		})
		Describe("Checking enc/dec", func() {

			var msg []byte

			BeforeEach(func() {
				msg = make([]byte, 8)
				binary.LittleEndian.PutUint64(msg, rand.Uint64())
				ct, err = sk.Encrypt(msg)
			})

			It("Should decrypt correctly", func() {
				Expect(err).To(BeNil())
				dmsg, err := sk.Decrypt(ct)
				Expect(err).To(BeNil())
				Expect(bytes.Equal(msg, dmsg)).To(BeTrue())
			})

			It("Should return false for forged ciphertext", func() {
				ct[0]++
				dmsg, err := sk.Decrypt(ct)
				Expect(err).NotTo(BeNil())
				Expect(bytes.Equal(msg, dmsg)).To(BeFalse())
			})
		})
	})
})
