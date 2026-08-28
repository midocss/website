package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Slugs of the roles created by the initial migration.
const (
	RoleSuperAdmin = "super_admin"
	RoleStaff      = "staff"
	RoleCustomer   = "customer"
)

// Effects of a per-user permission override.
const (
	PermissionEffectAllow = "allow"
	PermissionEffectDeny  = "deny"
)

type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	Slug        string       `gorm:"size:64;uniqueIndex" json:"slug"`
	NameAr      string       `gorm:"size:128" json:"name_ar"`
	NameEn      string       `gorm:"size:128" json:"name_en"`
	Description *string      `json:"description,omitempty"`
	IsSystem    bool         `json:"is_system"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Slug        string    `gorm:"size:128;uniqueIndex" json:"slug"`
	Resource    string    `gorm:"size:64" json:"resource"`
	Action      string    `gorm:"size:64" json:"action"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Permission) TableName() string { return "permissions" }

// UserPermission grants or revokes a single permission for one user, on top of
// whatever their role provides.
type UserPermission struct {
	UserID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"permission_id"`
	Effect       string    `gorm:"size:8" json:"effect"`
	CreatedAt    time.Time `json:"created_at"`

	Permission *Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}

func (UserPermission) TableName() string { return "user_permissions" }

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	RoleID       uuid.UUID      `gorm:"type:uuid" json:"role_id"`
	Email        string         `gorm:"size:255" json:"email"`
	PasswordHash string         `gorm:"column:password_hash" json:"-"`
	FullName     string         `gorm:"size:160" json:"full_name"`
	Phone        *string        `gorm:"size:32" json:"phone,omitempty"`
	Locale       string         `gorm:"size:8" json:"locale"`
	IsActive     bool           `json:"is_active"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Role            *Role            `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	UserPermissions []UserPermission `gorm:"foreignKey:UserID" json:"user_permissions,omitempty"`
}

func (User) TableName() string { return "users" }

// RefreshToken stores only the SHA-256 hash of the opaque token handed to the
// client, so a database leak cannot be replayed against the API.
type RefreshToken struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid" json:"user_id"`
	TokenHash  string     `gorm:"size:64;uniqueIndex" json:"-"`
	UserAgent  *string    `gorm:"size:255" json:"user_agent,omitempty"`
	IPAddress  *string    `gorm:"type:inet" json:"ip_address,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	ReplacedBy *uuid.UUID `gorm:"type:uuid" json:"replaced_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

func (t RefreshToken) IsUsable(now time.Time) bool {
	return t.RevokedAt == nil && now.Before(t.ExpiresAt)
}
