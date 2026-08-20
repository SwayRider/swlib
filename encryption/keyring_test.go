package encryption

import (
	"bytes"
	"testing"
)

func TestKeyRing_EncryptDecryptCurrent(t *testing.T) {
	current := testKey(t, 1)
	ring := NewKeyRing(current, nil)

	plaintext := []byte("private key material")
	blob, keyID, err := ring.EncryptCurrent(plaintext)
	if err != nil {
		t.Fatalf("EncryptCurrent failed: %v", err)
	}
	if keyID != Fingerprint(current) {
		t.Errorf("keyID = %q, want %q", keyID, Fingerprint(current))
	}

	got, err := ring.Decrypt(blob, keyID)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("roundtrip mismatch: got %q, want %q", got, plaintext)
	}
}

func TestKeyRing_DecryptByPreviousKey(t *testing.T) {
	oldKey := testKey(t, 1)
	newKey := testKey(t, 2)

	// Simulate a row encrypted before rotation.
	oldRing := NewKeyRing(oldKey, nil)
	plaintext := []byte("private key material")
	blob, oldKeyID, err := oldRing.EncryptCurrent(plaintext)
	if err != nil {
		t.Fatalf("EncryptCurrent failed: %v", err)
	}

	// After rotation: current is the new key, old key demoted to previous.
	rotatedRing := NewKeyRing(newKey, [][]byte{oldKey})

	got, err := rotatedRing.Decrypt(blob, oldKeyID)
	if err != nil {
		t.Fatalf("Decrypt via previous key failed: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("roundtrip mismatch: got %q, want %q", got, plaintext)
	}
}

func TestKeyRing_DecryptUnknownKeyIDFails(t *testing.T) {
	current := testKey(t, 1)
	dropped := testKey(t, 2)

	droppedRing := NewKeyRing(dropped, nil)
	blob, droppedKeyID, err := droppedRing.EncryptCurrent([]byte("data"))
	if err != nil {
		t.Fatalf("EncryptCurrent failed: %v", err)
	}

	// dropped is no longer configured anywhere (neither current nor previous).
	ring := NewKeyRing(current, nil)
	if _, err := ring.Decrypt(blob, droppedKeyID); err == nil {
		t.Error("expected error decrypting with a key id not present in the ring")
	}
}

func TestKeyRing_EmptyPreviousList(t *testing.T) {
	current := testKey(t, 1)
	ring := NewKeyRing(current, [][]byte{})

	blob, keyID, err := ring.EncryptCurrent([]byte("data"))
	if err != nil {
		t.Fatalf("EncryptCurrent failed: %v", err)
	}
	if _, err := ring.Decrypt(blob, keyID); err != nil {
		t.Fatalf("Decrypt failed with an empty previous list: %v", err)
	}
}
