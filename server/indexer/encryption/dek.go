package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
)

const (
	// encFileFormat: [4B dek_len][dek_len B encrypted_dek][encrypted_data]
	encFileMinLen = 4 + 1 // at minimum: 4 bytes length + 1 byte data
	dekKeyLen     = 32    // AES-256
)

// EncryptData generates a random DEK, encrypts data with AES-GCM, then encrypts
// the DEK with RSA-OAEP. Returns: [4B dek_len][enc_dek][enc_data].
func EncryptData(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	// Generate random DEK
	dek := make([]byte, dekKeyLen)
	if _, err := rand.Read(dek); err != nil {
		return nil, fmt.Errorf("generate DEK: %w", err)
	}

	// Encrypt data with AES-GCM
	encData, err := aesGCMEncrypt(dek, data)
	if err != nil {
		return nil, fmt.Errorf("encrypt data: %w", err)
	}

	// Encrypt DEK with RSA-OAEP
	encDEK, err := rsa.EncryptOAEP(sha256Hash(), rand.Reader, publicKey, dek, nil)
	if err != nil {
		return nil, fmt.Errorf("encrypt DEK: %w", err)
	}

	// Build output: [4B dek_len][enc_dek][enc_data]
	dekLen := uint32(len(encDEK))
	out := make([]byte, 4+len(encDEK)+len(encData))
	binary.BigEndian.PutUint32(out[:4], dekLen)
	copy(out[4:4+len(encDEK)], encDEK)
	copy(out[4+len(encDEK):], encData)

	return out, nil
}

// DecryptData parses the hybrid-encrypted blob, decrypts the DEK with RSA-OAEP,
// then decrypts the data with AES-GCM.
func DecryptData(encrypted []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if len(encrypted) < encFileMinLen {
		return nil, fmt.Errorf("encrypted data too short")
	}

	dekLen := binary.BigEndian.Uint32(encrypted[:4])
	if uint32(len(encrypted)) < 4+dekLen {
		return nil, fmt.Errorf("truncated encrypted DEK")
	}

	encDEK := encrypted[4 : 4+dekLen]
	encData := encrypted[4+dekLen:]

	// Decrypt DEK with RSA-OAEP
	dek, err := rsa.DecryptOAEP(sha256Hash(), rand.Reader, privateKey, encDEK, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt DEK: %w", err)
	}

	// Decrypt data with AES-GCM
	data, err := aesGCMDecrypt(dek, encData)
	if err != nil {
		return nil, fmt.Errorf("decrypt data: %w", err)
	}

	return data, nil
}

// IsEncryptedData checks if data starts with the encrypted format marker.
// Encrypted files have a 4-byte big-endian DEK length prefix; the DEK length
// for RSA-4096 + OAEP-SHA256 is always 512 bytes.
func IsEncryptedData(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	dekLen := binary.BigEndian.Uint32(data[:4])
	// RSA-4096 OAEP-SHA256 produces exactly 512-byte ciphertext
	return dekLen == 512 && uint32(len(data)) > 4+dekLen
}

func sha256Hash() hash.Hash {
	return sha256.New()
}
