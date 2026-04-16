package hasher_test

import (
	"strings"
	"testing"

	"github.com/example/go-backend-boilerplate/internal/lib/hasher"
)

func TestHashArgon2_Prefix(t *testing.T) {
	hash, err := hasher.HashArgon2("mypassword")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("expected hash to start with $argon2id$, got %s", hash)
	}
}

func TestHashArgon2_Unique(t *testing.T) {
	hash1, err := hasher.HashArgon2("mypassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hash2, err := hasher.HashArgon2("mypassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("two hashes of the same password should be different (different salts)")
	}
}

func TestVerifyArgon2_Correct(t *testing.T) {
	hash, err := hasher.HashArgon2("mypassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasher.VerifyArgon2("mypassword", hash) {
		t.Error("expected correct password to verify successfully")
	}
}

func TestVerifyArgon2_Wrong(t *testing.T) {
	hash, err := hasher.HashArgon2("mypassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasher.VerifyArgon2("wrongpassword", hash) {
		t.Error("expected wrong password to fail verification")
	}
}

func TestIsArgon2Hash_True(t *testing.T) {
	if !hasher.IsArgon2Hash("$argon2id$v=19$m=65536,t=3,p=4$abc$def") {
		t.Error("expected argon2id string to be detected")
	}
}

func TestIsArgon2Hash_False(t *testing.T) {
	if hasher.IsArgon2Hash("$2a$10$somebcrypthash") {
		t.Error("expected bcrypt string to not be detected as argon2")
	}
}
