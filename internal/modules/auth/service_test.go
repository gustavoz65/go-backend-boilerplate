package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func newTestAuthService() *Service {
	return &Service{}
}

func TestHashPassword_UsesArgon2(t *testing.T) {
	svc := newTestAuthService()
	hash, err := svc.hashPassword("testpass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hash) < 10 || hash[:10] != "$argon2id$" {
		t.Errorf("expected argon2id hash, got: %.20s", hash)
	}
}

func TestCheckPassword_Argon2_Correct(t *testing.T) {
	svc := newTestAuthService()
	hash, _ := svc.hashPassword("testpass")
	valid, needsMigration := svc.checkPassword("testpass", hash)
	if !valid {
		t.Error("expected valid=true for correct argon2 password")
	}
	if needsMigration {
		t.Error("expected needsMigration=false for argon2 hash")
	}
}

func TestCheckPassword_Argon2_Wrong(t *testing.T) {
	svc := newTestAuthService()
	hash, _ := svc.hashPassword("testpass")
	valid, _ := svc.checkPassword("wrongpass", hash)
	if valid {
		t.Error("expected valid=false for wrong password")
	}
}

func TestCheckPassword_Bcrypt_NeedsMigration(t *testing.T) {
	svc := newTestAuthService()
	bcryptHash, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.MinCost)
	valid, needsMigration := svc.checkPassword("testpass", string(bcryptHash))
	if !valid {
		t.Error("expected valid=true for correct bcrypt password")
	}
	if !needsMigration {
		t.Error("expected needsMigration=true for legacy bcrypt hash")
	}
}

func TestCheckPassword_Bcrypt_Wrong(t *testing.T) {
	svc := newTestAuthService()
	bcryptHash, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.MinCost)
	valid, needsMigration := svc.checkPassword("wrongpass", string(bcryptHash))
	if valid {
		t.Error("expected valid=false for wrong bcrypt password")
	}
	if needsMigration {
		t.Error("expected needsMigration=false when password is wrong")
	}
}
