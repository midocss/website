package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/midocss/website/internal/config"
	"github.com/midocss/website/pkg/apperr"
)

func testTokenManager(accessTTL time.Duration) *TokenManager {
	return NewTokenManager(config.JWT{
		Secret:          "0123456789abcdef0123456789abcdef",
		Issuer:          "test-issuer",
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: time.Hour,
	})
}

func TestParseAccessTokenRoundTrip(t *testing.T) {
	manager := testTokenManager(time.Minute)
	userID := uuid.New()

	token, expiresAt, err := manager.GenerateAccessToken(userID, "staff")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !expiresAt.After(time.Now().UTC()) {
		t.Fatalf("expected future expiry, got %s", expiresAt)
	}

	claims, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("user id = %s, want %s", claims.UserID, userID)
	}
	if claims.Role != "staff" {
		t.Errorf("role = %q, want staff", claims.Role)
	}
}

func TestParseAccessTokenRejectsExpired(t *testing.T) {
	manager := testTokenManager(-time.Minute)

	token, _, err := manager.GenerateAccessToken(uuid.New(), "customer")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = manager.ParseAccessToken(token)
	if apperr.From(err).Code != apperr.CodeUnauthorized {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestParseAccessTokenRejectsForeignSignature(t *testing.T) {
	issuer := testTokenManager(time.Minute)
	token, _, err := issuer.GenerateAccessToken(uuid.New(), "customer")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	other := NewTokenManager(config.JWT{
		Secret:         "ffffffffffffffffffffffffffffffff",
		Issuer:         "test-issuer",
		AccessTokenTTL: time.Minute,
	})
	if _, err := other.ParseAccessToken(token); err == nil {
		t.Fatal("expected a token signed with another secret to be rejected")
	}
}

func TestGenerateRefreshTokenIsUniqueAndHashed(t *testing.T) {
	manager := testTokenManager(time.Minute)

	first, firstHash, expiresAt, err := manager.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	second, secondHash, _, err := manager.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if first == second {
		t.Error("refresh tokens must not repeat")
	}
	if firstHash == secondHash {
		t.Error("refresh token hashes must not repeat")
	}
	if firstHash == first {
		t.Error("stored value must be a hash, not the token itself")
	}
	if HashRefreshToken(first) != firstHash {
		t.Error("HashRefreshToken must be deterministic")
	}
	if !expiresAt.After(time.Now().UTC()) {
		t.Error("refresh token must expire in the future")
	}
}
