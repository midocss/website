package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/midocss/website/internal/auth"
	"github.com/midocss/website/internal/transport/http/httpx"
	"github.com/midocss/website/internal/transport/http/middleware"
	"github.com/midocss/website/internal/transport/http/response"
	"github.com/midocss/website/pkg/apperr"
)

type AuthHandler struct {
	service auth.Service
}

func NewAuthHandler(service auth.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register creates a customer account and returns a fresh token pair.
func (h *AuthHandler) Register(c *gin.Context) {
	var in auth.RegisterInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	result, err := h.service.Register(c.Request.Context(), in, clientInfo(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusCreated, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var in auth.LoginInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	result, err := h.service.Login(c.Request.Context(), in, clientInfo(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, result)
}

// Refresh rotates a refresh token into a new token pair.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var in auth.RefreshInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	tokens, err := h.service.Refresh(c.Request.Context(), in.RefreshToken, clientInfo(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, tokens)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var in auth.RefreshInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	if err := h.service.Logout(c.Request.Context(), in.RefreshToken); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// LogoutAll revokes every active session of the caller.
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, apperr.Unauthorized("authentication required"))
		return
	}

	if err := h.service.LogoutAll(c.Request.Context(), userID); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, apperr.Unauthorized("authentication required"))
		return
	}

	profile, err := h.service.Profile(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, profile)
}

func clientInfo(c *gin.Context) auth.ClientInfo {
	return auth.ClientInfo{
		UserAgent: c.Request.UserAgent(),
		IPAddress: httpx.ClientIP(c),
	}
}
