package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/midocss/website/pkg/apperr"
)

func TestPasswordHasherRoundTrip(t *testing.T) {
	// bcrypt.MinCost keeps the test fast; production uses BcryptCost.
	hasher := NewPasswordHasher(bcrypt.MinCost)

	hash, err := hasher.Hash("Str0ngPassword")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "Str0ngPassword" {
		t.Fatal("password must not be stored in plain text")
	}
	if !hasher.Compare(hash, "Str0ngPassword") {
		t.Error("expected the correct password to match")
	}
	if hasher.Compare(hash, "wrong-password") {
		t.Error("expected a wrong password to be rejected")
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	cases := map[string]bool{
		"short1":         false,
		"onlyletters":    false,
		"12345678":       false,
		"passw0rd":       true,
		"Str0ngPassword": true,
	}

	for password, valid := range cases {
		err := ValidatePasswordStrength(password)
		if valid && err != nil {
			t.Errorf("%q: expected valid, got %v", password, err)
		}
		if !valid {
			if err == nil {
				t.Errorf("%q: expected a validation error", password)
				continue
			}
			if apperr.From(err).Code != apperr.CodeValidation {
				t.Errorf("%q: expected validation code, got %s", password, apperr.From(err).Code)
			}
		}
	}
}
