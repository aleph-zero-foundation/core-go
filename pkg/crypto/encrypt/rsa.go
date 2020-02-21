package encrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"math/big"
	"strconv"
	"strings"
)

// encryptionKey implements EncryptionKey interface using stdlib crypto/rsa
type encryptionKey struct {
	encKey *rsa.PublicKey
}

// decryptionKey implements DecryptionKey interface using stdlib crypto/rsa
type decryptionKey struct {
	decKey *rsa.PrivateKey
}

// GenerateKeys creates a pair of keys for encryption/decryption
func GenerateKeys() (EncryptionKey, DecryptionKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return &encryptionKey{&privKey.PublicKey}, &decryptionKey{privKey}, nil
}

func (ek *encryptionKey) Encrypt(msg []byte) (CipherText, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, ek.encKey, msg, nil)
}

func (dk *decryptionKey) Decrypt(ct CipherText) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, dk.decKey, ct, nil)
}

func (ek *encryptionKey) Encode() string {
	return ek.encKey.N.Text(big.MaxBase) + "#" + strconv.Itoa(ek.encKey.E)
}

// NewEncryptionKey creates encryptionKey from string representation
func NewEncryptionKey(text string) (EncryptionKey, error) {
	msg := "wrong format of encryption key"
	data := strings.Split(text, "#")
	if len(data) != 2 {
		return nil, errors.New(msg)
	}
	N, ok := new(big.Int).SetString(data[0], big.MaxBase)
	if !ok {
		return nil, errors.New(msg)
	}
	if N.Sign() != 1 {
		return nil, errors.New(msg)
	}
	E, err := strconv.Atoi(data[1])
	if err != nil {
		return nil, err
	}
	return &encryptionKey{&rsa.PublicKey{N, E}}, nil
}

func (dk *decryptionKey) Encode() string {
	result := ""
	result += dk.decKey.N.Text(big.MaxBase) + "#" + strconv.Itoa(dk.decKey.E) + "#"
	result += dk.decKey.D.Text(big.MaxBase) + "#"
	for i, p := range dk.decKey.Primes {
		if i > 0 {
			result += "*"
		}
		result += p.Text(big.MaxBase)
	}
	return result
}

// NewDecryptionKey creates a decryptionKey from its string representation
func NewDecryptionKey(text string) (DecryptionKey, error) {
	errMsg := "wrong format of decryption key"
	data := strings.Split(text, "#")
	if len(data) != 4 {
		return nil, errors.New(errMsg)
	}
	N, ok := new(big.Int).SetString(data[0], big.MaxBase)
	if !ok {
		return nil, errors.New(errMsg)
	}
	E, err := strconv.Atoi(data[1])
	if err != nil {
		return nil, errors.New(errMsg)
	}
	D, ok := new(big.Int).SetString(data[2], big.MaxBase)
	if !ok {
		return nil, errors.New(errMsg)
	}
	primes := []*big.Int{}
	primeStrings := strings.Split(data[3], "*")
	if len(primeStrings) < 2 {
		return nil, errors.New(errMsg)
	}
	for _, pString := range primeStrings {
		prime, ok := new(big.Int).SetString(pString, big.MaxBase)
		if !ok {
			return nil, errors.New(errMsg)
		}
		primes = append(primes, prime)
	}
	pKey := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{N, E},
		D:         D,
		Primes:    primes,
	}
	if err := pKey.Validate(); err != nil {
		return nil, err
	}
	return &decryptionKey{pKey}, nil
}
