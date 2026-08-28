package users

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/midocss/website/internal/domain"
	"github.com/midocss/website/pkg/apperr"
)

type Repository interface {
	List(ctx context.Context, query ListQuery) ([]domain.User, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, id uuid.UUID, fields map[string]any) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	EmailExists(ctx context.Context, email string) (bool, error)

	ListRoles(ctx context.Context) ([]domain.Role, error)
	FindRoleBySlug(ctx context.Context, slug string) (*domain.Role, error)
	ListPermissions(ctx context.Context) ([]domain.Permission, error)
	FindPermissionsBySlugs(ctx context.Context, slugs []string) ([]domain.Permission, error)
	ReplaceUserPermissions(ctx context.Context, userID uuid.UUID, overrides []domain.UserPermission) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) List(ctx context.Context, query ListQuery) ([]domain.User, int64, error) {
	page, perPage := query.Normalized()

	tx := r.db.WithContext(ctx).Model(&domain.User{}).Joins("JOIN roles ON roles.id = users.role_id")
	if search := strings.TrimSpace(query.Search); search != "" {
		pattern := "%" + strings.ToLower(search) + "%"
		tx = tx.Where("lower(users.full_name) LIKE ? OR lower(users.email) LIKE ?", pattern, pattern)
	}
	if query.RoleSlug != "" {
		tx = tx.Where("roles.slug = ?", query.RoleSlug)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, apperr.Internal(err)
	}

	var found []domain.User
	err := tx.Preload("Role").
		Order("users.created_at DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&found).Error
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}
	return found, total, nil
}

func (r *gormRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("UserPermissions.Permission").
		First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("user not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &user, nil
}

func (r *gormRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *gormRepository) Update(ctx context.Context, id uuid.UUID, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	result := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return apperr.Internal(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound("user not found")
	}
	return nil
}

func (r *gormRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id)
	if result.Error != nil {
		return apperr.Internal(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound("user not found")
	}
	return nil
}

func (r *gormRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("lower(email) = ?", strings.ToLower(strings.TrimSpace(email))).
		Count(&count).Error
	if err != nil {
		return false, apperr.Internal(err)
	}
	return count > 0, nil
}

func (r *gormRepository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	var roles []domain.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Order("slug").Find(&roles).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return roles, nil
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

func (r *gormRepository) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	var permissions []domain.Permission
	if err := r.db.WithContext(ctx).Order("resource, action").Find(&permissions).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return permissions, nil
}

func (r *gormRepository) FindPermissionsBySlugs(ctx context.Context, slugs []string) ([]domain.Permission, error) {
	if len(slugs) == 0 {
		return nil, nil
	}
	var permissions []domain.Permission
	if err := r.db.WithContext(ctx).Where("slug IN ?", slugs).Find(&permissions).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return permissions, nil
}

func (r *gormRepository) ReplaceUserPermissions(ctx context.Context, userID uuid.UUID, overrides []domain.UserPermission) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&domain.UserPermission{}).Error; err != nil {
			return err
		}
		if len(overrides) == 0 {
			return nil
		}
		return tx.Create(&overrides).Error
	})
	if err != nil {
		return apperr.Internal(err)
	}
	return nil
}
