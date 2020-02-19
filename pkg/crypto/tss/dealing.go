package tss

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"math/big"
	"sync"

	"gitlab.com/alephledger/core-go/pkg/crypto/bn256"
	"gitlab.com/alephledger/core-go/pkg/crypto/encrypt"
)

// New returns a Threshold Signature Scheme based on given slice of coefficients.
func New(nProc uint16, coeffs []*big.Int) *TSS {
	threshold := uint16(len(coeffs))
	secret := coeffs[threshold-1]

	globalVK := bn256.NewVerificationKey(secret)

	var wg sync.WaitGroup
	var sks = make([]*bn256.SecretKey, nProc)
	var vks = make([]*bn256.VerificationKey, nProc)

	for i := uint16(0); i < nProc; i++ {
		wg.Add(1)
		go func(ind uint16) {
			defer wg.Done()
			secret := poly(coeffs, big.NewInt(int64(ind+1)))
			sks[ind] = bn256.NewSecretKey(secret)
			vks[ind] = bn256.NewVerificationKey(secret)
		}(i)
	}
	wg.Wait()

	return &TSS{
		threshold: threshold,
		globalVK:  globalVK,
		vks:       vks,
		sks:       sks,
	}
}

// NewRandom generates a random polynomial of degree thereshold - 1 and builds a TSS based on the polynomial.
func NewRandomGlobal(nProc, threshold uint16) *TSS {
	var coeffs = make([]*big.Int, threshold)
	for i := uint16(0); i < threshold; i++ {
		c, _ := rand.Int(rand.Reader, bn256.Order)
		coeffs[i] = c
	}
	return New(nProc, coeffs)
}

// Encrypt encrypts secretKeys of the given TSS
// using given a set of encryptionKeys and returns an (unowned)ThresholdSignature.
func (tss *TSS) Encrypt(encryptionKeys []encrypt.SymmetricKey) (*ThresholdKey, error) {
	nProc := uint16(len(encryptionKeys))
	encSKs := make([]encrypt.CipherText, nProc)

	for i := uint16(0); i < nProc; i++ {
		encSK, err := encryptionKeys[i].Encrypt(tss.sks[i].Marshal())
		if err != nil {
			return nil, err
		}
		encSKs[i] = encSK
	}

	return &ThresholdKey{
		threshold: tss.threshold,
		globalVK:  tss.globalVK,
		vks:       tss.vks,
		encSKs:    encSKs,
	}, nil
}

// Encode returns a byte representation of the given (unowned) ThresholdKey in the following form
// (1) threshold, 2 bytes as uint16
// (2) length of marshalled globalVK, 4 bytes as uint32
// (3) marshalled globalVK
// (4) len(vks), 4 bytes as uint32
// (5) Marshalled vks in the form
//     a) length of marshalled vk
//     b) marshalled vk
// (6) Encrypted sks in the form
//     a) length of the cipher text
//     b) cipher text of the key
func (tk *ThresholdKey) Encode() []byte {
	data := make([]byte, 2+4)
	binary.LittleEndian.PutUint16(data[:2], tk.threshold)

	globalVKMarshalled := tk.globalVK.Marshal()
	binary.LittleEndian.PutUint32(data[2:6], uint32(len(globalVKMarshalled)))
	data = append(data, globalVKMarshalled...)

	dataLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(dataLen[:], uint32(len(tk.vks)))
	data = append(data, dataLen...)

	for _, vk := range tk.vks {
		vkMarshalled := vk.Marshal()
		binary.LittleEndian.PutUint32(dataLen, uint32(len(vkMarshalled)))
		data = append(data, dataLen...)
		data = append(data, vkMarshalled...)
	}

	for _, encSK := range tk.encSKs {
		binary.LittleEndian.PutUint32(dataLen, uint32(len(encSK)))
		data = append(data, dataLen...)
		data = append(data, encSK...)
	}
	return data
}

// Decode decodes encoded ThresholdSignature obtained from the dealer using given decryptionKey.
// It returns
// (1) decoded ThresholdSignature,
// (2) whether the owner's secretKey is correctly encoded and matches corresponding verification key,
// (3) an error in decoding (excluding errors obtained while decoding owners secret key),
func Decode(data []byte, dealer, owner uint16, decryptionKey encrypt.SymmetricKey) (*ThresholdKey, bool, error) {
	ind := 0
	dataTooShort := errors.New("Decoding tcoin failed. Given bytes slice is too short")
	if len(data) < ind+2 {
		return nil, false, dataTooShort
	}
	threshold := binary.LittleEndian.Uint16(data[:(ind + 2)])
	ind += 2

	if len(data) < ind+4 {
		return nil, false, dataTooShort
	}
	gvkLen := int(binary.LittleEndian.Uint32(data[ind:(ind + 4)]))
	ind += 4
	if len(data) < ind+gvkLen {
		return nil, false, dataTooShort
	}
	globalVK, err := new(bn256.VerificationKey).Unmarshal(data[ind:(ind + gvkLen)])
	if err != nil {
		return nil, false, errors.New("unmarshal of globalVK failed")
	}
	ind += gvkLen

	if len(data) < ind+4 {
		return nil, false, dataTooShort
	}
	nProcesses := uint16(binary.LittleEndian.Uint32(data[ind:(ind + 4)]))
	ind += 4
	vks := make([]*bn256.VerificationKey, nProcesses)
	for i := range vks {
		if len(data) < ind+4 {
			return nil, false, dataTooShort
		}
		vkLen := int(binary.LittleEndian.Uint32(data[ind:(ind + 4)]))
		ind += 4
		if len(data) < ind+vkLen {
			return nil, false, dataTooShort
		}
		vks[i], err = new(bn256.VerificationKey).Unmarshal(data[ind:(ind + vkLen)])
		if err != nil {
			return nil, false, errors.New("unmarshal of vk failed")
		}
		ind += vkLen
	}
	encSKs := make([]encrypt.CipherText, nProcesses)
	for i := range encSKs {
		if len(data) < ind+4 {
			return nil, false, dataTooShort
		}
		skLen := int(binary.LittleEndian.Uint32(data[ind:(ind + 4)]))
		ind += 4
		if len(data) < ind+skLen {
			return nil, false, dataTooShort
		}
		encSKs[i] = data[ind:(ind + skLen)]
		ind += skLen
	}

	sk, err := decryptSecretKey(encSKs[owner], vks[owner], decryptionKey)

	return &ThresholdKey{
		dealer:    dealer,
		owner:     owner,
		threshold: threshold,
		globalVK:  globalVK,
		vks:       vks,
		encSKs:    encSKs,
		sk:        sk,
	}, (err == nil), nil
}

func decryptSecretKey(data []byte, vk *bn256.VerificationKey, decryptionKey encrypt.SymmetricKey) (*bn256.SecretKey, error) {
	decrypted, err := decryptionKey.Decrypt(data)
	if err != nil {
		return nil, err
	}

	sk, err := new(bn256.SecretKey).Unmarshal(decrypted)
	if err != nil {
		return nil, err
	}

	if !bn256.VerifyKeys(vk, sk) {
		return nil, errors.New("secret key doesn't match with the verification key")
	}
	return sk, nil
}

// CheckSecretKey checks whether the secret key of the given pid is correct.
func (tk *ThresholdKey) CheckSecretKey(pid uint16, decryptionKey encrypt.SymmetricKey) bool {
	_, err := decryptSecretKey(tk.encSKs[pid], tk.vks[pid], decryptionKey)
	return err == nil
}
