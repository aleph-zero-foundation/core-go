package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/sha3"
)

// symmetricKey implements SymmetricKey interface using stdlib crypto/aes.
type symmetricKey struct {
	gcm cipher.AEAD
}

// NewSymmetricKey creates a symmetric key for encryption/decryption.
func NewSymmetricKey(key []byte) (SymmetricKey, error) {
	hkey := make([]byte, 32)
	sha3.ShakeSum128(hkey[:], key)

	c, err := aes.NewCipher(hkey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	return &symmetricKey{gcm}, nil
}

func (sk *symmetricKey) Encrypt(msg []byte) (CipherText, error) {
	nonce := make([]byte, sk.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return sk.gcm.Seal(nonce, nonce, msg, nil), nil
}

func (sk *symmetricKey) Decrypt(ct CipherText) ([]byte, error) {
	nonceSize := sk.gcm.NonceSize()
	if len(ct) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ct := ct[:nonceSize], ct[nonceSize:]

	return sk.gcm.Open(nil, nonce, ct, nil)
}
