package tss_test

import (
	"gitlab.com/alephledger/core-go/pkg/crypto/bn256"
	"gitlab.com/alephledger/core-go/pkg/crypto/encrypt"
	"gitlab.com/alephledger/core-go/pkg/crypto/p2p"
	. "gitlab.com/alephledger/core-go/pkg/crypto/tss"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("T", func() {
	var (
		n, t, dealer uint16
		msg          []byte
		tcs          []*ThresholdKey
		share        []*Share
		sKeys        []*p2p.SecretKey
		pKeys        []*p2p.PublicKey
		p2pKeys      [][]encrypt.SymmetricKey
	)
	Context("Between small number of processes", func() {
		Describe("Signature shares", func() {
			BeforeEach(func() {
				n, t, dealer = 10, 3, 5

				gtc := NewRandomGlobal(n, t)
				tcs = make([]*ThresholdKey, n)
				sKeys = make([]*p2p.SecretKey, n)
				pKeys = make([]*p2p.PublicKey, n)
				p2pKeys = make([][]encrypt.SymmetricKey, n)
				for i := uint16(0); i < n; i++ {
					pKeys[i], sKeys[i], _ = p2p.GenerateKeys()
				}
				for i := uint16(0); i < n; i++ {
					p2pKeys[i], _ = p2p.Keys(sKeys[i], pKeys, i)
				}
				tc, err := gtc.Encrypt(p2pKeys[dealer])
				Expect(err).NotTo(HaveOccurred())
				tcEncoded := tc.Encode()
				for i := uint16(0); i < n; i++ {
					tcs[i], _, err = Decode(tcEncoded, dealer, i, p2pKeys[i][dealer])
					Expect(err).NotTo(HaveOccurred())
				}
				msg = []byte("xyz")
				share = make([]*Share, n)
				for i := uint16(0); i < n; i++ {
					share[i] = tcs[i].CreateShare(msg)
				}
			})
			It("should be verified correctly", func() {
				Expect(tcs[2].VerifyShare(share[1], msg)).To(BeTrue())
				Expect(tcs[2].VerifyShare(share[1], append(msg, byte(1)))).To(BeFalse())
			})
			It("Should be correctly combined by t-parties", func() {
				c, ok := tcs[0].CombineShares(share[:t])
				Expect(ok).To(BeTrue())
				Expect(tcs[0].VerifySignature(c, msg)).To(BeTrue())
				Expect(tcs[0].VerifySignature(c, append(msg, byte(1)))).To(BeFalse())
			})
			It("Should be correctly combined by more than t-parties", func() {
				c, ok := tcs[0].CombineShares(share)
				Expect(ok).To(BeTrue())
				Expect(tcs[0].VerifySignature(c, msg)).To(BeTrue())
				Expect(tcs[0].VerifySignature(c, append(msg, byte(1)))).To(BeFalse())
			})
			It("Should be combined to the same  by two different sets of t-parties", func() {
				c1, ok := tcs[0].CombineShares(share[:t])
				Expect(ok).To(BeTrue())
				c2, ok := tcs[0].CombineShares(share[(n - t):])
				Expect(ok).To(BeTrue())
				Expect(c1.Marshal()).To(Equal(c2.Marshal()))
			})
			It("Shouldn't be correctly combined by t-1-parties", func() {
				_, ok := tcs[0].CombineShares(share[:(t - 1)])
				Expect(ok).To(BeFalse())
			})
			It("Should be marshalled and unmarshalled correctly", func() {
				for i := uint16(0); i < n; i++ {
					csMarshalled := share[i].Marshal()
					var cs = new(Share)
					err := cs.Unmarshal(csMarshalled)
					Expect(err).NotTo(HaveOccurred())
					Expect(tcs[0].VerifyShare(cs, msg)).To(BeTrue())
				}
			})
		})
	})
	Context("Signature unmarshal", func() {
		Context("On an empty slice", func() {
			It("Should return an error", func() {
				c := new(Signature)
				err := c.Unmarshal([]byte{})
				Expect(err).To(HaveOccurred())
			})
		})
		Context("On a incorrect slice having correct length", func() {
			It("Should return an error", func() {
				c := new(Signature)
				data := make([]byte, bn256.SignatureLength)
				data[0] = 1
				err := c.Unmarshal(data)
				Expect(err).To(HaveOccurred())
			})
		})
		Context("On a correctly marshalled ", func() {
			It("Should work without errors", func() {
				c := new(Signature)

				_, priv, err := bn256.GenerateKeys()
				Expect(err).NotTo(HaveOccurred())
				data := []byte{1, 2, 3}

				err = c.Unmarshal(priv.Sign(data).Marshal())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
