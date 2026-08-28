package auth

import (
	"time"

	"github.com/google/uuid"
)

type RegisterInput struct {
	FullName string `json:"full_name" binding:"required,min=3,max=160"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	Phone    string `json:"phone" binding:"omitempty,max=32"`
	Locale   string `json:"locale" binding:"omitempty,oneof=ar en"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ClientInfo is captured per session so the customer can review and revoke
// their active devices later.
type ClientInfo struct {
	UserAgent string
	IPAddress string
}

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type UserProfile struct {
	ID          uuid.UUID `json:"id"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	Phone       *string   `json:"phone,omitempty"`
	Locale      string    `json:"locale"`
	Role        string    `json:"role"`
	IsActive    bool      `json:"is_active"`
	Permissions []string  `json:"permissions"`
}

type AuthResult struct {
	User   UserProfile `json:"user"`
	Tokens TokenPair   `json:"tokens"`
}
