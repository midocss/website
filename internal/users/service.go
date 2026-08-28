package users

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/midocss/website/internal/auth"
	"github.com/midocss/website/internal/domain"
	"github.com/midocss/website/internal/rbac"
	"github.com/midocss/website/pkg/apperr"
)

type Service interface {
	List(ctx context.Context, query ListQuery) ([]UserView, int64, error)
	Get(ctx context.Context, id uuid.UUID) (*UserView, error)
	Create(ctx context.Context, in CreateUserInput) (*UserView, error)
	Update(ctx context.Context, actorID, id uuid.UUID, in UpdateUserInput) (*UserView, error)
	Delete(ctx context.Context, actorID, id uuid.UUID) error
	Roles(ctx context.Context) ([]domain.Role, error)
	Permissions(ctx context.Context) ([]domain.Permission, error)
	SetPermissionOverrides(ctx context.Context, id uuid.UUID, in PermissionOverrideInput) (*UserView, error)
}

type service struct {
	repo       Repository
	hasher     *auth.PasswordHasher
	authorizer rbac.Authorizer
}

func NewService(repo Repository, hasher *auth.PasswordHasher, authorizer rbac.Authorizer) Service {
	return &service{repo: repo, hasher: hasher, authorizer: authorizer}
}

func (s *service) List(ctx context.Context, query ListQuery) ([]UserView, int64, error) {
	found, total, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	views := make([]UserView, 0, len(found))
	for i := range found {
		views = append(views, toView(&found[i], nil))
	}
	return views, total, nil
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*UserView, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	permissions, err := s.authorizer.Permissions(ctx, id)
	if err != nil {
		return nil, err
	}
	view := toView(user, permissions)
	return &view, nil
}

func (s *service) Create(ctx context.Context, in CreateUserInput) (*UserView, error) {
	if err := auth.ValidatePasswordStrength(in.Password); err != nil {
		return nil, err
	}

	exists, err := s.repo.EmailExists(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperr.Conflict("an account with this email already exists")
	}

	role, err := s.repo.FindRoleBySlug(ctx, in.RoleSlug)
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
		Email:        strings.ToLower(strings.TrimSpace(in.Email)),
		PasswordHash: hash,
		FullName:     in.FullName,
		Locale:       locale,
		IsActive:     true,
		Role:         role,
	}
	if in.Phone != "" {
		user.Phone = &in.Phone
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	view := toView(user, nil)
	return &view, nil
}

func (s *service) Update(ctx context.Context, actorID, id uuid.UUID, in UpdateUserInput) (*UserView, error) {
	fields := map[string]any{}
	if in.FullName != nil {
		fields["full_name"] = *in.FullName
	}
	if in.Phone != nil {
		fields["phone"] = *in.Phone
	}
	if in.Locale != nil {
		fields["locale"] = *in.Locale
	}
	if in.IsActive != nil {
		// Guard against an admin locking themselves out.
		if actorID == id && !*in.IsActive {
			return nil, apperr.Conflict("you cannot deactivate your own account")
		}
		fields["is_active"] = *in.IsActive
	}
	if in.RoleSlug != nil {
		role, err := s.repo.FindRoleBySlug(ctx, *in.RoleSlug)
		if err != nil {
			return nil, err
		}
		fields["role_id"] = role.ID
	}

	if err := s.repo.Update(ctx, id, fields); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *service) Delete(ctx context.Context, actorID, id uuid.UUID) error {
	if actorID == id {
		return apperr.Conflict("you cannot delete your own account")
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *service) Roles(ctx context.Context) ([]domain.Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *service) Permissions(ctx context.Context) ([]domain.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *service) SetPermissionOverrides(ctx context.Context, id uuid.UUID, in PermissionOverrideInput) (*UserView, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return nil, err
	}

	slugs := append(append([]string{}, in.Allow...), in.Deny...)
	permissions, err := s.repo.FindPermissionsBySlugs(ctx, slugs)
	if err != nil {
		return nil, err
	}

	byslug := make(map[string]uuid.UUID, len(permissions))
	for _, permission := range permissions {
		byslug[permission.Slug] = permission.ID
	}

	overrides := make([]domain.UserPermission, 0, len(slugs))
	seen := make(map[string]struct{}, len(slugs))
	for _, entry := range []struct {
		slugs  []string
		effect string
	}{
		{in.Allow, domain.PermissionEffectAllow},
		{in.Deny, domain.PermissionEffectDeny},
	} {
		for _, slug := range entry.slugs {
			permissionID, ok := byslug[slug]
			if !ok {
				return nil, apperr.Validation("unknown permission").WithFields(apperr.FieldError{
					Field:   "permissions",
					Message: "unknown permission: " + slug,
				})
			}
			if _, duplicated := seen[slug]; duplicated {
				return nil, apperr.Validation("conflicting permission override").WithFields(apperr.FieldError{
					Field:   "permissions",
					Message: slug + " cannot be both allowed and denied",
				})
			}
			seen[slug] = struct{}{}
			overrides = append(overrides, domain.UserPermission{
				UserID:       id,
				PermissionID: permissionID,
				Effect:       entry.effect,
			})
		}
	}

	if err := s.repo.ReplaceUserPermissions(ctx, id, overrides); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func toView(user *domain.User, permissions []string) UserView {
	roleSlug := ""
	if user.Role != nil {
		roleSlug = user.Role.Slug
	}
	return UserView{
		ID:          user.ID,
		FullName:    user.FullName,
		Email:       user.Email,
		Phone:       user.Phone,
		Locale:      user.Locale,
		Role:        roleSlug,
		IsActive:    user.IsActive,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		Permissions: permissions,
	}
}
