// Package rbac resolves the effective permission set of a user: the
// permissions of their role, plus per-user "allow" overrides, minus per-user
// "deny" overrides. Super admins bypass the checks entirely.
package rbac

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/midocss/website/internal/domain"
	"github.com/midocss/website/pkg/apperr"
)

type Authorizer interface {
	// Permissions returns the effective permission slugs of a user.
	Permissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	// Can reports whether the user holds every one of the given permissions.
	Can(ctx context.Context, userID uuid.UUID, permissions ...string) (bool, error)
}

type authorizer struct {
	db *gorm.DB
}

func New(db *gorm.DB) Authorizer {
	return &authorizer{db: db}
}

type permissionRow struct {
	Slug   string
	Effect *string
}

func (a *authorizer) Permissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var user domain.User
	if err := a.db.WithContext(ctx).Preload("Role").First(&user, "id = ?", userID).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	if user.Role != nil && user.Role.Slug == domain.RoleSuperAdmin {
		return a.allPermissions(ctx)
	}

	var rows []permissionRow
	err := a.db.WithContext(ctx).Raw(`
		SELECT p.slug AS slug, up.effect AS effect
		FROM permissions p
		LEFT JOIN role_permissions rp ON rp.permission_id = p.id AND rp.role_id = ?
		LEFT JOIN user_permissions up ON up.permission_id = p.id AND up.user_id = ?
		WHERE rp.role_id IS NOT NULL OR up.user_id IS NOT NULL
	`, user.RoleID, userID).Scan(&rows).Error
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return effective(rows), nil
}

// effective drops every permission carrying a per-user deny override.
func effective(rows []permissionRow) []string {
	slugs := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Effect != nil && *row.Effect == domain.PermissionEffectDeny {
			continue
		}
		slugs = append(slugs, row.Slug)
	}
	return slugs
}

func (a *authorizer) Can(ctx context.Context, userID uuid.UUID, permissions ...string) (bool, error) {
	if len(permissions) == 0 {
		return true, nil
	}

	granted, err := a.Permissions(ctx, userID)
	if err != nil {
		return false, err
	}

	return hasAll(granted, permissions), nil
}

// hasAll reports whether granted contains every required permission.
func hasAll(granted, required []string) bool {
	set := make(map[string]struct{}, len(granted))
	for _, slug := range granted {
		set[slug] = struct{}{}
	}
	for _, slug := range required {
		if _, ok := set[slug]; !ok {
			return false
		}
	}
	return true
}

func (a *authorizer) allPermissions(ctx context.Context) ([]string, error) {
	var slugs []string
	if err := a.db.WithContext(ctx).Model(&domain.Permission{}).Pluck("slug", &slugs).Error; err != nil {
		return nil, apperr.Internal(err)
	}
	return slugs, nil
}
