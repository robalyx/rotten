package tui

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	"golang.org/x/crypto/argon2"
)

// HashType represents the different hashing algorithms available.
type HashType string

const (
	// HashTypeArgon2id uses the Argon2id algorithm for hashing.
	HashTypeArgon2id HashType = "argon2id"
	// HashTypeSHA256 uses the SHA256 algorithm for hashing.
	HashTypeSHA256 HashType = "sha256"
)

// HashResult represents a hashed ID with its index.
type HashResult struct {
	Index int
	Hash  string
}

// hashID converts a single ID to a hash using the specified algorithm with the provided salt.
func hashID(id uint64, salt string, hashType HashType, iterations uint32, memory uint32) string {
	// Convert ID to bytes in little-endian format
	idBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(idBytes, id)

	var hash []byte
	switch hashType {
	case HashTypeArgon2id:
		// Use Argon2id with specified parameters
		hash = argon2.IDKey(idBytes, []byte(salt), iterations, memory*1024, 1, 32)
	case HashTypeSHA256:
		// Iterative SHA256 hashing with salt
		hash = []byte(salt)
		h := sha256.New()
		for range iterations {
			h.Reset()
			h.Write(idBytes)
			h.Write(hash)
			hash = h.Sum(nil)
		}
	}

	return hex.EncodeToString(hash)
}
