// Package encrypt is a wrapper around crypto/aes and crypto/rsa.
package encrypt

import "bytes"

// CipherText represents encrypted data.
type CipherText []byte

// CTEq checks ciphertexts' equality.
func CTEq(c1, c2 CipherText) bool {
	return bytes.Equal(c1, c2)
}

// EncryptionKey is used for encrypting messages.
type EncryptionKey interface {
	// Encrypt encrypts message.
	Encrypt([]byte) (CipherText, error)
	// Encode encodes the encryption key.
	Encode() string
}

// DecryptionKey is used for decrypting ciphertexts encrypted with corresponding encryption key.
type DecryptionKey interface {
	// Decrypt decrypts ciphertext that was encrypted with corresponding encryption key.
	Decrypt(CipherText) ([]byte, error)
	// Encode encodes the decryption key.
	Encode() string
}

// SymmetricKey is used for both encrypting and decrypting messages.
type SymmetricKey interface {
	// Encrypt encrypts message.
	Encrypt([]byte) (CipherText, error)
	// Decrypt decrypts ciphertext that was encrypted with the key.
	Decrypt(CipherText) ([]byte, error)
}
