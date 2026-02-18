package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	plaintext := []byte("secret credential data")
	passphrase := "test-passphrase-12345"

	encrypted, err := Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if bytes.Equal(encrypted, plaintext) {
		t.Error("encrypted data should differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted, passphrase)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted data doesn't match: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptWrongPassphrase(t *testing.T) {
	plaintext := []byte("secret data")
	encrypted, err := Encrypt(plaintext, "correct-passphrase")
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(encrypted, "wrong-passphrase")
	if err == nil {
		t.Error("expected error with wrong passphrase")
	}
}

func TestEncryptDifferentCiphertexts(t *testing.T) {
	plaintext := []byte("same data")
	passphrase := "same-passphrase"

	enc1, _ := Encrypt(plaintext, passphrase)
	enc2, _ := Encrypt(plaintext, passphrase)

	// Due to random salt/nonce, encryptions should differ
	if bytes.Equal(enc1, enc2) {
		t.Error("two encryptions of same data should produce different ciphertexts")
	}

	// But both should decrypt to same plaintext
	dec1, _ := Decrypt(enc1, passphrase)
	dec2, _ := Decrypt(enc2, passphrase)

	if !bytes.Equal(dec1, dec2) {
		t.Error("both should decrypt to same plaintext")
	}
}

func TestDecryptInvalidPayload(t *testing.T) {
	_, err := Decrypt([]byte("not json"), "passphrase")
	if err == nil {
		t.Error("expected error for invalid payload")
	}
}

func TestDecryptWrongVersion(t *testing.T) {
	payload := `{"version":99,"salt":"AA==","nonce":"AA==","data":"AA=="}`
	_, err := Decrypt([]byte(payload), "passphrase")
	if err == nil {
		t.Error("expected error for wrong version")
	}
}

func TestHashPassphrase(t *testing.T) {
	h1 := HashPassphrase("test")
	h2 := HashPassphrase("test")
	h3 := HashPassphrase("different")

	if h1 != h2 {
		t.Error("same passphrase should produce same hash")
	}
	if h1 == h3 {
		t.Error("different passphrases should produce different hashes")
	}
	if h1 == "" {
		t.Error("hash should not be empty")
	}
}

func TestEncryptEmptyData(t *testing.T) {
	encrypted, err := Encrypt([]byte{}, "passphrase")
	if err != nil {
		t.Fatalf("Encrypt empty data failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, "passphrase")
	if err != nil {
		t.Fatalf("Decrypt empty data failed: %v", err)
	}

	if len(decrypted) != 0 {
		t.Error("decrypted empty data should be empty")
	}
}

func TestEncryptLargeData(t *testing.T) {
	// 1MB of data
	plaintext := bytes.Repeat([]byte("x"), 1024*1024)
	passphrase := "large-data-test"

	encrypted, err := Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatalf("Encrypt large data failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, passphrase)
	if err != nil {
		t.Fatalf("Decrypt large data failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("large data round-trip failed")
	}
}
