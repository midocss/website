package auth

import (
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/midocss/website/pkg/apperr"
)

// BcryptCost is deliberately above the bcrypt default to slow down offline
// cracking of leaked hashes.
const BcryptCost = 12

type PasswordHasher struct {
	cost int
}

func NewPasswordHasher(cost int) *PasswordHasher {
	if cost <= 0 {
		cost = BcryptCost
	}
	return &PasswordHasher{cost: cost}
}

func (h *PasswordHasher) Hash(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", apperr.Internal(err)
	}
	return string(hashed), nil
}

func (h *PasswordHasher) Compare(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// ValidatePasswordStrength enforces the minimum password policy shared by
// registration, password reset and admin-created accounts.
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return apperr.Validation("password is too weak").WithFields(apperr.FieldError{
			Field:   "password",
			Message: "must be at least 8 characters long",
		})
	}

	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return apperr.Validation("password is too weak").WithFields(apperr.FieldError{
			Field:   "password",
			Message: "must contain at least one letter and one digit",
		})
	}
	return nil
}
