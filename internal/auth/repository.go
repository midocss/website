package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/midocss/website/internal/domain"
	"github.com/midocss/website/pkg/apperr"
)

// Repository is the data access contract of the auth service.
type Repository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	FindUserByEmail(ctx context.Context, email string) (*domain.User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	TouchLastLogin(ctx context.Context, id uuid.UUID, at time.Time) error

	FindRoleBySlug(ctx context.Context, slug string) (*domain.Role, error)

	CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	FindRefreshTokenByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID, replacedBy *uuid.UUID, at time.Time) error
	RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID, at time.Time) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) CreateUser(ctx context.Context, user *domain.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.Email = normalizeEmail(user.Email)
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *gormRepository) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("lower(email) = ?", normalizeEmail(email)).
		First(&user).Error
	return userOrError(&user, err)
}

func (r *gormRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Preload("Role").First(&user, "id = ?", id).Error
	return userOrError(&user, err)
}

func (r *gormRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("lower(email) = ?", normalizeEmail(email)).
		Count(&count).Error
	if err != nil {
		return false, apperr.Internal(err)
	}
	return count > 0, nil
}

func (r *gormRepository) TouchLastLogin(ctx context.Context, id uuid.UUID, at time.Time) error {
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", id).
		Update("last_login_at", at).Error
	if err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *gormRepository) FindRoleBySlug(ctx context.Context, slug string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).First(&role, "slug = ?", slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("role not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &role, nil
}

func (r *gormRepository) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *gormRepository) FindRefreshTokenByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := r.db.WithContext(ctx).First(&token, "token_hash = ?", hash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.Unauthorized("invalid refresh token")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &token, nil
}

func (r *gormRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID, replacedBy *uuid.UUID, at time.Time) error {
	updates := map[string]any{"revoked_at": at}
	if replacedBy != nil {
		updates["replaced_by"] = *replacedBy
	}
	err := r.db.WithContext(ctx).
		Model(&domain.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Updates(updates).Error
	if err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *gormRepository) RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID, at time.Time) error {
	err := r.db.WithContext(ctx).
		Model(&domain.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", at).Error
	if err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func userOrError(user *domain.User, err error) (*domain.User, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("user not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
