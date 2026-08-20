package encryption

import "fmt"

// KeyRing holds a current master key (used for all new encryption) plus a
// set of previous master keys (used only to decrypt data encrypted before a
// rotation). Keys are looked up by Fingerprint, so callers must persist the
// fingerprint returned by EncryptCurrent alongside the ciphertext.
type KeyRing struct {
	current   []byte
	currentID string
	keys      map[string][]byte // fingerprint -> key, includes current
}

// NewKeyRing builds a KeyRing from a required current key and zero or more
// previous keys.
func NewKeyRing(current []byte, previous [][]byte) *KeyRing {
	r := &KeyRing{
		current:   current,
		currentID: Fingerprint(current),
		keys:      make(map[string][]byte, len(previous)+1),
	}
	r.keys[r.currentID] = current
	for _, key := range previous {
		r.keys[Fingerprint(key)] = key
	}
	return r
}

// EncryptCurrent encrypts plaintext under the ring's current key and
// returns the ciphertext blob along with the key's fingerprint, to be
// stored alongside it for later lookup by Decrypt.
func (r *KeyRing) EncryptCurrent(plaintext []byte) (blob []byte, keyID string, err error) {
	blob, err = Encrypt(r.current, plaintext)
	if err != nil {
		return nil, "", err
	}
	return blob, r.currentID, nil
}

// Decrypt decrypts blob using the key identified by keyID (checked against
// the current key and every previous key). It returns an error if keyID
// does not match any key configured on the ring.
func (r *KeyRing) Decrypt(blob []byte, keyID string) ([]byte, error) {
	key, ok := r.keys[keyID]
	if !ok {
		return nil, fmt.Errorf("encryption: no key configured for key id %q", keyID)
	}
	return Decrypt(key, blob)
}
