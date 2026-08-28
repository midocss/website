// Command seed creates (or updates) the super admin account from the
// SEED_ADMIN_* environment variables. It is idempotent.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/midocss/website/internal/auth"
	"github.com/midocss/website/internal/config"
	"github.com/midocss/website/internal/domain"
	"github.com/midocss/website/internal/platform/database"
	"github.com/midocss/website/internal/platform/logger"
)

func main() {
	// Loaded before the flags so .env can supply their defaults.
	_ = godotenv.Load()

	email := flag.String("email", os.Getenv("SEED_ADMIN_EMAIL"), "super admin email")
	password := flag.String("password", os.Getenv("SEED_ADMIN_PASSWORD"), "super admin password")
	name := flag.String("name", envOr("SEED_ADMIN_NAME", "Super Admin"), "super admin full name")
	flag.Parse()

	if err := run(*email, *password, *name); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run(email, password, name string) error {
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return errors.New("both -email and -password (or SEED_ADMIN_EMAIL/SEED_ADMIN_PASSWORD) are required")
	}
	if err := auth.ValidatePasswordStrength(password); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New(cfg.App.Environment, cfg.App.Debug)
	slog.SetDefault(log)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.Open(ctx, cfg.Database, false)
	if err != nil {
		return err
	}
	defer func() { _ = database.Close(db) }()

	var role domain.Role
	if err := db.WithContext(ctx).First(&role, "slug = ?", domain.RoleSuperAdmin).Error; err != nil {
		return err
	}

	hash, err := auth.NewPasswordHasher(auth.BcryptCost).Hash(password)
	if err != nil {
		return err
	}
	normalized := strings.ToLower(strings.TrimSpace(email))

	var existing domain.User
	err = db.WithContext(ctx).First(&existing, "lower(email) = ?", normalized).Error
	switch {
	case err == nil:
		updates := map[string]any{"password_hash": hash, "role_id": role.ID, "is_active": true}
		if err := db.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
		log.Info("super admin updated", "email", normalized)
	case errors.Is(err, gorm.ErrRecordNotFound):
		user := domain.User{
			ID:           uuid.New(),
			RoleID:       role.ID,
			Email:        normalized,
			PasswordHash: hash,
			FullName:     name,
			Locale:       cfg.App.DefaultLang,
			IsActive:     true,
		}
		if err := db.WithContext(ctx).Create(&user).Error; err != nil {
			return err
		}
		log.Info("super admin created", "email", normalized)
	default:
		return err
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
