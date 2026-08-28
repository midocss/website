package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/midocss/website/internal/config"
	"github.com/midocss/website/pkg/apperr"
)

// Claims is the payload of the short-lived access token. Permissions are
// intentionally not embedded: they are resolved per request so a revoked
// permission takes effect immediately instead of at the next token refresh.
type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"uid"`
	Role   string    `json:"role"`
}

type TokenManager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

func NewTokenManager(cfg config.JWT) *TokenManager {
	return &TokenManager{
		secret:     []byte(cfg.Secret),
		issuer:     cfg.Issuer,
		accessTTL:  cfg.AccessTokenTTL,
		refreshTTL: cfg.RefreshTokenTTL,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

func (m *TokenManager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *TokenManager) RefreshTTL() time.Duration { return m.refreshTTL }

// GenerateAccessToken issues a signed JWT for the given user.
func (m *TokenManager) GenerateAccessToken(userID uuid.UUID, roleSlug string) (string, time.Time, error) {
	issuedAt := m.now()
	expiresAt := issuedAt.Add(m.accessTTL)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    m.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID: userID,
		Role:   roleSlug,
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, apperr.Internal(err)
	}
	return signed, expiresAt, nil
}

// ParseAccessToken validates the signature, algorithm and expiry of a token.
func (m *TokenManager) ParseAccessToken(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		return m.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, apperr.Unauthorized("invalid or expired access token").WithCause(err)
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, apperr.Unauthorized("invalid access token")
	}
	return claims, nil
}

// GenerateRefreshToken returns the opaque token given to the client together
// with the hash that is persisted.
func (m *TokenManager) GenerateRefreshToken() (plain string, hash string, expiresAt time.Time, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", time.Time{}, apperr.Internal(err)
	}
	plain = base64.RawURLEncoding.EncodeToString(buf)
	return plain, HashRefreshToken(plain), m.now().Add(m.refreshTTL), nil
}

func HashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
