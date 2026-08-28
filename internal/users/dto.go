package users

import (
	"time"

	"github.com/google/uuid"
)

type CreateUserInput struct {
	FullName string `json:"full_name" binding:"required,min=3,max=160"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	Phone    string `json:"phone" binding:"omitempty,max=32"`
	Locale   string `json:"locale" binding:"omitempty,oneof=ar en"`
	RoleSlug string `json:"role_slug" binding:"required"`
}

type UpdateUserInput struct {
	FullName *string `json:"full_name" binding:"omitempty,min=3,max=160"`
	Phone    *string `json:"phone" binding:"omitempty,max=32"`
	Locale   *string `json:"locale" binding:"omitempty,oneof=ar en"`
	RoleSlug *string `json:"role_slug" binding:"omitempty"`
	IsActive *bool   `json:"is_active"`
}

// PermissionOverrideInput replaces the caller-specified user's per-user
// permission overrides in one call.
type PermissionOverrideInput struct {
	Allow []string `json:"allow" binding:"omitempty,dive,required"`
	Deny  []string `json:"deny" binding:"omitempty,dive,required"`
}

type ListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PerPage  int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search   string `form:"search" binding:"omitempty,max=160"`
	RoleSlug string `form:"role" binding:"omitempty,max=64"`
}

func (q ListQuery) Normalized() (page, perPage int) {
	page, perPage = q.Page, q.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	return page, perPage
}

type UserView struct {
	ID          uuid.UUID  `json:"id"`
	FullName    string     `json:"full_name"`
	Email       string     `json:"email"`
	Phone       *string    `json:"phone,omitempty"`
	Locale      string     `json:"locale"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	Permissions []string   `json:"permissions,omitempty"`
}
