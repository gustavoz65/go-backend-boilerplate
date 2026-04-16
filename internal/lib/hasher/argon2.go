package hasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Memory  = 64 * 1024 // 64 MB em KiB
	argon2Time    = 3
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

// HashArgon2 gera um hash argon2id no formato PHC string.
// Formato: $argon2id$v=19$m=65536,t=3,p=4$<base64salt>$<base64hash>
func HashArgon2(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)

	hash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Threads,
		b64Salt, b64Key,
	)
	return hash, nil
}

// VerifyArgon2 verifica uma senha contra um hash argon2id.
func VerifyArgon2(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	// Formato: ["", "argon2id", "v=19", "m=65536,t=3,p=4", "<salt>", "<hash>"]
	if len(parts) != 6 {
		return false
	}

	if parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
		return false
	}

	var memory, iterations uint32
	var threads uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &threads)
	if err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	storedKey, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	keyLen := uint32(len(storedKey))
	computedKey := argon2.IDKey([]byte(password), salt, iterations, memory, threads, keyLen)

	return subtle.ConstantTimeCompare(computedKey, storedKey) == 1
}

// IsArgon2Hash retorna true se o hash foi gerado com argon2id.
func IsArgon2Hash(hash string) bool {
	return strings.HasPrefix(hash, "$argon2id$")
}
