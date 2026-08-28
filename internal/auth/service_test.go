package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/midocss/website/internal/domain"
	"github.com/midocss/website/pkg/apperr"
)

type fakeRepo struct {
	users         map[uuid.UUID]*domain.User
	usersByEmail  map[string]*domain.User
	roles         map[string]*domain.Role
	refreshTokens map[uuid.UUID]*domain.RefreshToken
	tokensByHash  map[string]*domain.RefreshToken
}

func newFakeRepo() *fakeRepo {
	customer := &domain.Role{ID: uuid.New(), Slug: domain.RoleCustomer}
	return &fakeRepo{
		users:         map[uuid.UUID]*domain.User{},
		usersByEmail:  map[string]*domain.User{},
		roles:         map[string]*domain.Role{domain.RoleCustomer: customer},
		refreshTokens: map[uuid.UUID]*domain.RefreshToken{},
		tokensByHash:  map[string]*domain.RefreshToken{},
	}
}

func (f *fakeRepo) CreateUser(_ context.Context, user *domain.User) error {
	f.users[user.ID] = user
	f.usersByEmail[normalizeEmail(user.Email)] = user
	return nil
}

func (f *fakeRepo) FindUserByEmail(_ context.Context, email string) (*domain.User, error) {
	user, ok := f.usersByEmail[normalizeEmail(email)]
	if !ok {
		return nil, apperr.NotFound("user not found")
	}
	return user, nil
}

func (f *fakeRepo) FindUserByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	user, ok := f.users[id]
	if !ok {
		return nil, apperr.NotFound("user not found")
	}
	return user, nil
}

func (f *fakeRepo) EmailExists(_ context.Context, email string) (bool, error) {
	_, ok := f.usersByEmail[normalizeEmail(email)]
	return ok, nil
}

func (f *fakeRepo) TouchLastLogin(_ context.Context, id uuid.UUID, at time.Time) error {
	if user, ok := f.users[id]; ok {
		user.LastLoginAt = &at
	}
	return nil
}

func (f *fakeRepo) FindRoleBySlug(_ context.Context, slug string) (*domain.Role, error) {
	role, ok := f.roles[slug]
	if !ok {
		return nil, apperr.NotFound("role not found")
	}
	return role, nil
}

func (f *fakeRepo) CreateRefreshToken(_ context.Context, token *domain.RefreshToken) error {
	f.refreshTokens[token.ID] = token
	f.tokensByHash[token.TokenHash] = token
	return nil
}

func (f *fakeRepo) FindRefreshTokenByHash(_ context.Context, hash string) (*domain.RefreshToken, error) {
	token, ok := f.tokensByHash[hash]
	if !ok {
		return nil, apperr.Unauthorized("invalid refresh token")
	}
	return token, nil
}

func (f *fakeRepo) RevokeRefreshToken(_ context.Context, id uuid.UUID, replacedBy *uuid.UUID, at time.Time) error {
	if token, ok := f.refreshTokens[id]; ok && token.RevokedAt == nil {
		token.RevokedAt = &at
		token.ReplacedBy = replacedBy
	}
	return nil
}

func (f *fakeRepo) RevokeUserRefreshTokens(_ context.Context, userID uuid.UUID, at time.Time) error {
	for _, token := range f.refreshTokens {
		if token.UserID == userID && token.RevokedAt == nil {
			token.RevokedAt = &at
		}
	}
	return nil
}

type fakeAuthorizer struct {
	permissions []string
}

func (f fakeAuthorizer) Permissions(context.Context, uuid.UUID) ([]string, error) {
	return f.permissions, nil
}

func (f fakeAuthorizer) Can(context.Context, uuid.UUID, ...string) (bool, error) {
	return true, nil
}

func newTestService(t *testing.T) (Service, *fakeRepo) {
	t.Helper()
	repo := newFakeRepo()
	svc := NewService(
		repo,
		testTokenManager(time.Minute),
		NewPasswordHasher(bcrypt.MinCost),
		fakeAuthorizer{permissions: []string{"orders.view"}},
	)
	return svc, repo
}

func registerCustomer(t *testing.T, svc Service) *AuthResult {
	t.Helper()
	result, err := svc.Register(context.Background(), RegisterInput{
		FullName: "Test Customer",
		Email:    "Customer@Example.com",
		Password: "passw0rd123",
	}, ClientInfo{})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	return result
}

func TestRegisterCreatesCustomerAndRejectsDuplicateEmail(t *testing.T) {
	svc, repo := newTestService(t)

	result := registerCustomer(t, svc)
	if result.User.Role != domain.RoleCustomer {
		t.Errorf("role = %q, want customer", result.User.Role)
	}
	if result.User.Email != "customer@example.com" {
		t.Errorf("email = %q, want the normalized lowercase form", result.User.Email)
	}
	if result.Tokens.AccessToken == "" || result.Tokens.RefreshToken == "" {
		t.Fatal("expected a full token pair")
	}
	stored := repo.usersByEmail["customer@example.com"]
	if stored.PasswordHash == "passw0rd123" {
		t.Fatal("password must be hashed before storage")
	}

	_, err := svc.Register(context.Background(), RegisterInput{
		FullName: "Someone Else",
		Email:    "customer@example.com",
		Password: "passw0rd123",
	}, ClientInfo{})
	if apperr.From(err).Code != apperr.CodeConflict {
		t.Fatalf("expected a conflict on duplicate email, got %v", err)
	}
}

func TestRegisterRejectsWeakPassword(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.Register(context.Background(), RegisterInput{
		FullName: "Weak Password",
		Email:    "weak@example.com",
		Password: "abcdefgh",
	}, ClientInfo{})
	if apperr.From(err).Code != apperr.CodeValidation {
		t.Fatalf("expected a validation error, got %v", err)
	}
}

func TestLoginUsesGenericErrorForUnknownEmailAndWrongPassword(t *testing.T) {
	svc, _ := newTestService(t)
	registerCustomer(t, svc)

	_, unknownErr := svc.Login(context.Background(), LoginInput{
		Email:    "nobody@example.com",
		Password: "passw0rd123",
	}, ClientInfo{})
	_, wrongErr := svc.Login(context.Background(), LoginInput{
		Email:    "customer@example.com",
		Password: "wrong-passw0rd",
	}, ClientInfo{})

	if apperr.From(unknownErr).Message != apperr.From(wrongErr).Message {
		t.Fatal("login must not reveal whether the email exists")
	}
	if apperr.From(wrongErr).Code != apperr.CodeUnauthorized {
		t.Fatalf("expected unauthorized, got %v", wrongErr)
	}
}

func TestLoginRejectsDisabledAccount(t *testing.T) {
	svc, repo := newTestService(t)
	result := registerCustomer(t, svc)
	repo.users[result.User.ID].IsActive = false

	_, err := svc.Login(context.Background(), LoginInput{
		Email:    "customer@example.com",
		Password: "passw0rd123",
	}, ClientInfo{})
	if apperr.From(err).Code != apperr.CodeForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestRefreshRotatesTokenAndRevokesTheOldOne(t *testing.T) {
	svc, repo := newTestService(t)
	result := registerCustomer(t, svc)
	original := result.Tokens.RefreshToken

	rotated, err := svc.Refresh(context.Background(), original, ClientInfo{})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if rotated.RefreshToken == original {
		t.Fatal("refresh must issue a new refresh token")
	}

	old := repo.tokensByHash[HashRefreshToken(original)]
	if old.RevokedAt == nil {
		t.Fatal("the presented refresh token must be revoked")
	}
	if old.ReplacedBy == nil {
		t.Fatal("the revoked token must point at its replacement")
	}
}

// Replaying a revoked refresh token means it leaked, so every session of that
// user is dropped.
func TestRefreshReuseRevokesTheWholeSession(t *testing.T) {
	svc, _ := newTestService(t)
	result := registerCustomer(t, svc)
	original := result.Tokens.RefreshToken

	rotated, err := svc.Refresh(context.Background(), original, ClientInfo{})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	if _, err := svc.Refresh(context.Background(), original, ClientInfo{}); apperr.From(err).Code != apperr.CodeUnauthorized {
		t.Fatalf("expected the replayed token to be rejected, got %v", err)
	}
	if _, err := svc.Refresh(context.Background(), rotated.RefreshToken, ClientInfo{}); err == nil {
		t.Fatal("expected the rotated token to be revoked after a replay")
	}
}

func TestLogoutRevokesOnlyTheGivenToken(t *testing.T) {
	svc, repo := newTestService(t)
	result := registerCustomer(t, svc)

	second, err := svc.Login(context.Background(), LoginInput{
		Email:    "customer@example.com",
		Password: "passw0rd123",
	}, ClientInfo{})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if err := svc.Logout(context.Background(), result.Tokens.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if repo.tokensByHash[HashRefreshToken(result.Tokens.RefreshToken)].RevokedAt == nil {
		t.Error("the logged out token must be revoked")
	}
	if repo.tokensByHash[HashRefreshToken(second.Tokens.RefreshToken)].RevokedAt != nil {
		t.Error("other sessions must stay active")
	}

	if err := svc.LogoutAll(context.Background(), result.User.ID); err != nil {
		t.Fatalf("logout all: %v", err)
	}
	if repo.tokensByHash[HashRefreshToken(second.Tokens.RefreshToken)].RevokedAt == nil {
		t.Error("logout-all must revoke every session")
	}
}

func TestProfileIncludesEffectivePermissions(t *testing.T) {
	svc, _ := newTestService(t)
	result := registerCustomer(t, svc)

	profile, err := svc.Profile(context.Background(), result.User.ID)
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if len(profile.Permissions) != 1 || profile.Permissions[0] != "orders.view" {
		t.Errorf("permissions = %v, want [orders.view]", profile.Permissions)
	}
}
