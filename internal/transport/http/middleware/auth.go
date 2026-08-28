package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/midocss/website/internal/auth"
	"github.com/midocss/website/internal/rbac"
	"github.com/midocss/website/internal/transport/http/response"
	"github.com/midocss/website/pkg/apperr"
)

const (
	contextUserID   = "auth_user_id"
	contextRoleSlug = "auth_role_slug"
)

// Authenticate validates the bearer access token and stores the caller
// identity on the request context.
func Authenticate(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		scheme, token, found := strings.Cut(header, " ")
		if !found || !strings.EqualFold(scheme, "bearer") || strings.TrimSpace(token) == "" {
			response.Fail(c, apperr.Unauthorized("missing bearer token"))
			return
		}

		claims, err := tokens.ParseAccessToken(strings.TrimSpace(token))
		if err != nil {
			response.Fail(c, err)
			return
		}

		c.Set(contextUserID, claims.UserID)
		c.Set(contextRoleSlug, claims.Role)
		c.Next()
	}
}

// RequirePermissions rejects the request unless the caller holds every listed
// permission. Must run after Authenticate.
func RequirePermissions(authorizer rbac.Authorizer, permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := UserID(c)
		if !ok {
			response.Fail(c, apperr.Unauthorized("authentication required"))
			return
		}

		allowed, err := authorizer.Can(c.Request.Context(), userID, permissions...)
		if err != nil {
			response.Fail(c, err)
			return
		}
		if !allowed {
			response.Fail(c, apperr.Forbidden("you do not have permission to perform this action"))
			return
		}
		c.Next()
	}
}

// UserID returns the authenticated caller id, if any.
func UserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get(contextUserID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := value.(uuid.UUID)
	return id, ok
}

// RoleSlug returns the authenticated caller role slug, if any.
func RoleSlug(c *gin.Context) (string, bool) {
	value, exists := c.Get(contextRoleSlug)
	if !exists {
		return "", false
	}
	slug, ok := value.(string)
	return slug, ok
}
