package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/midocss/website/internal/domain"
	"github.com/midocss/website/internal/rbac"
	"github.com/midocss/website/pkg/apperr"
)

type Service interface {
	Register(ctx context.Context, in RegisterInput, client ClientInfo) (*AuthResult, error)
	Login(ctx context.Context, in LoginInput, client ClientInfo) (*AuthResult, error)
	Refresh(ctx context.Context, refreshToken string, client ClientInfo) (*TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	Profile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
}

type service struct {
	repo       Repository
	tokens     *TokenManager
	hasher     *PasswordHasher
	authorizer rbac.Authorizer
	now        func() time.Time
}

func NewService(repo Repository, tokens *TokenManager, hasher *PasswordHasher, authorizer rbac.Authorizer) Service {
	return &service{
		repo:       repo,
		tokens:     tokens,
		hasher:     hasher,
		authorizer: authorizer,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

func (s *service) Register(ctx context.Context, in RegisterInput, client ClientInfo) (*AuthResult, error) {
	if err := ValidatePasswordStrength(in.Password); err != nil {
		return nil, err
	}

	exists, err := s.repo.EmailExists(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperr.Conflict("an account with this email already exists")
	}

	// Public registration always creates a customer; staff accounts are
	// provisioned from the dashboard.
	role, err := s.repo.FindRoleBySlug(ctx, domain.RoleCustomer)
	if err != nil {
		return nil, err
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}

	locale := in.Locale
	if locale == "" {
		locale = "ar"
	}

	user := &domain.User{
		ID:           uuid.New(),
		RoleID:       role.ID,
		Email:        normalizeEmail(in.Email),
		PasswordHash: hash,
		FullName:     in.FullName,
		Locale:       locale,
		IsActive:     true,
		Role:         role,
	}
	if in.Phone != "" {
		user.Phone = &in.Phone
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return s.issue(ctx, user, client)
}

func (s *service) Login(ctx context.Context, in LoginInput, client ClientInfo) (*AuthResult, error) {
	user, err := s.repo.FindUserByEmail(ctx, in.Email)
	if err != nil {
		var appErr *apperr.Error
		// Do not reveal whether the email exists.
		if errors.As(err, &appErr) && appErr.Code == apperr.CodeNotFound {
			return nil, apperr.Unauthorized("invalid email or password")
		}
		return nil, err
	}

	if !s.hasher.Compare(user.PasswordHash, in.Password) {
		return nil, apperr.Unauthorized("invalid email or password")
	}
	if !user.IsActive {
		return nil, apperr.Forbidden("this account is disabled")
	}

	now := s.now()
	if err := s.repo.TouchLastLogin(ctx, user.ID, now); err != nil {
		return nil, err
	}
	return s.issue(ctx, user, client)
}

// Refresh rotates the refresh token: the presented token is revoked and linked
// to its replacement, and reuse of an already revoked token kills the whole
// session family.
func (s *service) Refresh(ctx context.Context, refreshToken string, client ClientInfo) (*TokenPair, error) {
	stored, err := s.repo.FindRefreshTokenByHash(ctx, HashRefreshToken(refreshToken))
	if err != nil {
		return nil, err
	}

	now := s.now()
	if !stored.IsUsable(now) {
		if stored.RevokedAt != nil {
			if err := s.repo.RevokeUserRefreshTokens(ctx, stored.UserID, now); err != nil {
				return nil, err
			}
		}
		return nil, apperr.Unauthorized("refresh token is expired or revoked")
	}

	user, err := s.repo.FindUserByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, apperr.Forbidden("this account is disabled")
	}

	pair, newToken, err := s.issueTokens(ctx, user, client)
	if err != nil {
		return nil, err
	}
	if err := s.repo.RevokeRefreshToken(ctx, stored.ID, &newToken.ID, now); err != nil {
		return nil, err
	}
	return pair, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	stored, err := s.repo.FindRefreshTokenByHash(ctx, HashRefreshToken(refreshToken))
	if err != nil {
		return err
	}
	return s.repo.RevokeRefreshToken(ctx, stored.ID, nil, s.now())
}

func (s *service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.repo.RevokeUserRefreshTokens(ctx, userID, s.now())
}

func (s *service) Profile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile, err := s.profile(ctx, user)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *service) issue(ctx context.Context, user *domain.User, client ClientInfo) (*AuthResult, error) {
	pair, _, err := s.issueTokens(ctx, user, client)
	if err != nil {
		return nil, err
	}
	profile, err := s.profile(ctx, user)
	if err != nil {
		return nil, err
	}
	return &AuthResult{User: profile, Tokens: *pair}, nil
}

func (s *service) issueTokens(ctx context.Context, user *domain.User, client ClientInfo) (*TokenPair, *domain.RefreshToken, error) {
	roleSlug := ""
	if user.Role != nil {
		roleSlug = user.Role.Slug
	}

	access, accessExpiry, err := s.tokens.GenerateAccessToken(user.ID, roleSlug)
	if err != nil {
		return nil, nil, err
	}

	plain, hash, refreshExpiry, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, nil, err
	}

	record := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: refreshExpiry,
	}
	if client.UserAgent != "" {
		record.UserAgent = &client.UserAgent
	}
	if client.IPAddress != "" {
		record.IPAddress = &client.IPAddress
	}
	if err := s.repo.CreateRefreshToken(ctx, record); err != nil {
		return nil, nil, err
	}

	return &TokenPair{
		AccessToken:      access,
		RefreshToken:     plain,
		TokenType:        "Bearer",
		ExpiresAt:        accessExpiry,
		RefreshExpiresAt: refreshExpiry,
	}, record, nil
}

func (s *service) profile(ctx context.Context, user *domain.User) (UserProfile, error) {
	permissions, err := s.authorizer.Permissions(ctx, user.ID)
	if err != nil {
		return UserProfile{}, err
	}

	roleSlug := ""
	if user.Role != nil {
		roleSlug = user.Role.Slug
	}

	return UserProfile{
		ID:          user.ID,
		FullName:    user.FullName,
		Email:       user.Email,
		Phone:       user.Phone,
		Locale:      user.Locale,
		Role:        roleSlug,
		IsActive:    user.IsActive,
		Permissions: permissions,
	}, nil
}
