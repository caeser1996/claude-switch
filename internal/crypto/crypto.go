package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	// SaltSize is the size of the random salt for key derivation.
	SaltSize = 32
	// NonceSize is the AES-GCM nonce size.
	NonceSize = 12
	// KeySize is the AES-256 key size.
	KeySize = 32
)

// EncryptedPayload is the wire format for encrypted profile exports.
type EncryptedPayload struct {
	Version int    `json:"version"`
	Salt    []byte `json:"salt"`
	Nonce   []byte `json:"nonce"`
	Data    []byte `json:"data"`
}

// deriveKey uses scrypt to derive an AES-256 key from a passphrase and salt.
func deriveKey(passphrase string, salt []byte) ([]byte, error) {
	// scrypt parameters: N=32768, r=8, p=1 (good balance of security/speed)
	return scrypt.Key([]byte(passphrase), salt, 32768, 8, 1, KeySize)
}

// Encrypt encrypts plaintext with a passphrase using AES-256-GCM.
func Encrypt(plaintext []byte, passphrase string) ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("cannot generate salt: %w", err)
	}

	key, err := deriveKey(passphrase, salt)
	if err != nil {
		return nil, fmt.Errorf("cannot derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("cannot create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cannot create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("cannot generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	payload := EncryptedPayload{
		Version: 1,
		Salt:    salt,
		Nonce:   nonce,
		Data:    ciphertext,
	}

	return json.Marshal(payload)
}

// Decrypt decrypts an encrypted payload with a passphrase.
func Decrypt(encrypted []byte, passphrase string) ([]byte, error) {
	var payload EncryptedPayload
	if err := json.Unmarshal(encrypted, &payload); err != nil {
		return nil, fmt.Errorf("invalid encrypted payload: %w", err)
	}

	if payload.Version != 1 {
		return nil, fmt.Errorf("unsupported encryption version: %d", payload.Version)
	}

	key, err := deriveKey(passphrase, payload.Salt)
	if err != nil {
		return nil, fmt.Errorf("cannot derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("cannot create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cannot create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, payload.Nonce, payload.Data, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong passphrase?): %w", err)
	}

	return plaintext, nil
}

// HashPassphrase returns a SHA-256 hash of a passphrase (for display/verification, not storage).
func HashPassphrase(passphrase string) string {
	h := sha256.Sum256([]byte(passphrase))
	return fmt.Sprintf("%x", h[:8])
}
