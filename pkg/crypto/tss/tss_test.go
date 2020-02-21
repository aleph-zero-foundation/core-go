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
		shares       []*Share
		sKeys        []*p2p.SecretKey
		pKeys        []*p2p.PublicKey
		p2pKeys      [][]encrypt.SymmetricKey
	)
	Context("Between small number of processes", func() {
		Describe("Signature shares", func() {
			BeforeEach(func() {
				n, t, dealer = 10, 3, 5

				gtc := NewRandom(n, t)
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
				shares = make([]*Share, n)
				for i := uint16(0); i < n; i++ {
					shares[i] = tcs[i].CreateShare(msg)
				}
			})
			It("should be verified correctly", func() {
				Expect(tcs[2].VerifyShare(shares[1], msg)).To(BeTrue())
				Expect(tcs[2].VerifyShare(shares[1], append(msg, byte(1)))).To(BeFalse())
			})
			It("Should be correctly combined by t-parties", func() {
				c, ok := tcs[0].CombineShares(shares[:t])
				Expect(ok).To(BeTrue())
				Expect(tcs[0].VerifySignature(c, msg)).To(BeTrue())
				Expect(tcs[0].VerifySignature(c, append(msg, byte(1)))).To(BeFalse())
			})
			It("Should be correctly combined by more than t-parties", func() {
				c, ok := tcs[0].CombineShares(shares)
				Expect(ok).To(BeTrue())
				Expect(tcs[0].VerifySignature(c, msg)).To(BeTrue())
				Expect(tcs[0].VerifySignature(c, append(msg, byte(1)))).To(BeFalse())
			})
			It("Should be combined to the same  by two different sets of t-parties", func() {
				c1, ok := tcs[0].CombineShares(shares[:t])
				Expect(ok).To(BeTrue())
				c2, ok := tcs[0].CombineShares(shares[(n - t):])
				Expect(ok).To(BeTrue())
				Expect(c1.Marshal()).To(Equal(c2.Marshal()))
			})
			It("Shouldn't be correctly combined by t-1-parties", func() {
				_, ok := tcs[0].CombineShares(shares[:(t - 1)])
				Expect(ok).To(BeFalse())
			})
			It("Should be marshalled and unmarshalled correctly", func() {
				for i := uint16(0); i < n; i++ {
					csMarshalled := shares[i].Marshal()
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

	Context("WeakThresholdKey", func() {
		var (
			tcs1           []*ThresholdKey
			tcs2           []*ThresholdKey
			wtcs           []*WeakThresholdKey
			shareProviders map[uint16]bool
		)
		BeforeEach(func() {
			n, t = 10, 4

			wtcs = make([]*WeakThresholdKey, n)
			tcs1 = make([]*ThresholdKey, n)
			tcs2 = make([]*ThresholdKey, n)
			shareProviders = map[uint16]bool{}
			for i := uint16(0); i < n; i++ {
				shareProviders[i] = true
			}
			sKeys = make([]*p2p.SecretKey, n)
			pKeys = make([]*p2p.PublicKey, n)
			p2pKeys = make([][]encrypt.SymmetricKey, n)
			for i := uint16(0); i < n; i++ {
				pKeys[i], sKeys[i], _ = p2p.GenerateKeys()
			}
			for i := uint16(0); i < n; i++ {
				p2pKeys[i], _ = p2p.Keys(sKeys[i], pKeys, i)
			}

			gtc1 := NewRandom(n, t)
			tc1, _ := gtc1.Encrypt(p2pKeys[0])
			gtc2 := NewRandom(n, t)
			tc2, _ := gtc2.Encrypt(p2pKeys[1])

			tc1Encoded := tc1.Encode()
			tc2Encoded := tc2.Encode()
			for i := uint16(0); i < n; i++ {
				tcs1[i], _, _ = Decode(tc1Encoded, 0, i, p2pKeys[i][0])
				tcs2[i], _, _ = Decode(tc2Encoded, 1, i, p2pKeys[i][1])
				wtcs[i] = CreateWTK([]*ThresholdKey{tcs1[i], tcs2[i]}, shareProviders)
			}
			msg = []byte("xyz")
		})
		Describe("share providers", func() {
			It("should produce shares for share providers", func() {
				Expect(wtcs[0].CreateShare(msg)).ToNot(BeNil())
			})
			It("should not produce shares for parties that are not share providers", func() {
				delete(shareProviders, 1)
				wtcs[1] = CreateWTK([]*ThresholdKey{tcs1[1], tcs2[1]}, shareProviders)
				Expect(wtcs[1].CreateShare(msg)).To(BeNil())
			})
		})
		Describe("coin shares", func() {
			BeforeEach(func() {
				shares = make([]*Share, n)
				for i := uint16(0); i < n; i++ {
					shares[i] = wtcs[i].CreateShare(msg)
				}
			})
			It("should be the sum of coin shares among single coins", func() {
				shs := SumShares([]*Share{tcs1[0].CreateShare(msg), tcs2[0].CreateShare(msg)})
				Expect(shs.Marshal()).To(Equal(shares[0].Marshal()))
			})
		})
	})
})
