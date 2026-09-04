package encryption

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/argon2"
)

const (
	keyDirName       = "encryption"
	publicKeyFile    = "public.key"
	privateKeyFile   = "private.key.enc"
	argonTime        = 3
	argonMemory      = 64 * 1024
	argonThreads     = 4
	argonKeyLen      = 32
	argonSaltLen     = 16
	aesKeyLen        = 32
	rsaKeyBits      = 4096
)

// KeyPair holds an RSA key pair for hybrid encryption.
type KeyPair struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey // nil when locked
}

// KeyDir returns the encryption key directory path.
func KeyDir(dataDir string) string {
	return filepath.Join(dataDir, keyDirName)
}

// KeyPairPath returns the public key path.
func KeyPairPath(dataDir string) string {
	return filepath.Join(KeyDir(dataDir), publicKeyFile)
}

// PrivateKeyPath returns the encrypted private key path.
func PrivateKeyPath(dataDir string) string {
	return filepath.Join(KeyDir(dataDir), privateKeyFile)
}

// IsLocked returns true if the private key file exists but we haven't loaded it.
func IsLocked(dataDir string) bool {
	return fileExists(PrivateKeyPath(dataDir))
}

// KeysExist returns true if the key pair has been generated.
func KeysExist(dataDir string) bool {
	return fileExists(KeyPairPath(dataDir)) && fileExists(PrivateKeyPath(dataDir))
}

// GenerateKeyPair generates a new RSA-4096 key pair, encrypts the private key
// with a password-derived key (Argon2id), and saves both to disk.
// Returns the public key for immediate use.
func GenerateKeyPair(dataDir, password string) (*KeyPair, error) {
	priv, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key: %w", err)
	}

	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	derivedKey := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	encPriv, err := encryptPrivateKey(priv, derivedKey, salt)
	if err != nil {
		return nil, fmt.Errorf("encrypt private key: %w", err)
	}

	dir := KeyDir(dataDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create key directory: %w", err)
	}

	// Write public key
	pubPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&priv.PublicKey),
	}
	pubPath := KeyPairPath(dataDir)
	if err := os.WriteFile(pubPath, pem.EncodeToMemory(pubPEM), 0o644); err != nil {
		return nil, fmt.Errorf("write public key: %w", err)
	}

	// Write encrypted private key
	privPath := PrivateKeyPath(dataDir)
	if err := os.WriteFile(privPath, encPriv, 0o600); err != nil {
		return nil, fmt.Errorf("write encrypted private key: %w", err)
	}

	return &KeyPair{PublicKey: &priv.PublicKey, PrivateKey: priv}, nil
}

// UnlockKeyPair loads the public key and decrypts the private key with the password.
func UnlockKeyPair(dataDir, password string) (*KeyPair, error) {
	// Load public key
	pubPath := KeyPairPath(dataDir)
	pubPEM, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	pubBlock, _ := pem.Decode(pubPEM)
	if pubBlock == nil {
		return nil, fmt.Errorf("decode public key PEM")
	}
	pub, err := x509.ParsePKCS1PublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	// Load and decrypt private key
	privPath := PrivateKeyPath(dataDir)
	encPriv, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("read encrypted private key: %w", err)
	}

	priv, err := decryptPrivateKey(encPriv, password)
	if err != nil {
		return nil, fmt.Errorf("decrypt private key (wrong password?): %w", err)
	}

	return &KeyPair{PublicKey: pub, PrivateKey: priv}, nil
}

// LoadPublicKey loads only the public key (for encrypting DEKs while locked).
func LoadPublicKey(dataDir string) (*rsa.PublicKey, error) {
	pubPath := KeyPairPath(dataDir)
	pubPEM, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	pubBlock, _ := pem.Decode(pubPEM)
	if pubBlock == nil {
		return nil, fmt.Errorf("decode public key PEM")
	}
	return x509.ParsePKCS1PublicKey(pubBlock.Bytes)
}

// encryptPrivateKey encrypts an RSA private key with a derived key.
// Format: [salt:16][iv:16][encrypted PKCS8]
func encryptPrivateKey(priv *rsa.PrivateKey, key, salt []byte) ([]byte, error) {
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}

	iv := make([]byte, aesBlockLen)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("generate IV: %w", err)
	}

	encrypted, err := aesCBCEncrypt(key, iv, privHDRound(privDER))
	if err != nil {
		return nil, fmt.Errorf("encrypt private key: %w", err)
	}

	result := make([]byte, 0, len(salt)+len(iv)+len(encrypted))
	result = append(result, salt...)
	result = append(result, iv...)
	result = append(result, encrypted...)
	return result, nil
}

// decryptPrivateKey decrypts a private key blob encrypted by encryptPrivateKey.
func decryptPrivateKey(data []byte, password string) (*rsa.PrivateKey, error) {
	if len(data) < argonSaltLen+aesBlockLen {
		return nil, fmt.Errorf("encrypted key too short")
	}

	salt := data[:argonSaltLen]
	iv := data[argonSaltLen : argonSaltLen+aesBlockLen]
	ciphertext := data[argonSaltLen+aesBlockLen:]

	derivedKey := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	plainDER, err := aesCBCDecrypt(derivedKey, iv, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("AES decrypt: %w", err)
	}
	plainDER = privHDRoundUnpad(plainDER)

	priv, err := x509.ParsePKCS8PrivateKey(plainDER)
	if err != nil {
		return nil, fmt.Errorf("parse PKCS8 private key: %w", err)
	}
	rsaKey, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}
	return rsaKey, nil
}

// privHDRound pads PKCS#8 DER to a multiple of the AES block size (PKCS#7).
func privHDRound(data []byte) []byte {
	pad := aesBlockLen - (len(data) % aesBlockLen)
	return append(data, bytes.Repeat([]byte{byte(pad)}, pad)...)
}

func privHDRoundUnpad(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	pad := int(data[len(data)-1])
	if pad > 0 && pad <= aesBlockLen && pad <= len(data) {
		return data[:len(data)-pad]
	}
	return data
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
