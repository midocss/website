package handler

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/midocss/website/internal/transport/http/httpx"
	"github.com/midocss/website/internal/transport/http/middleware"
	"github.com/midocss/website/internal/transport/http/response"
	"github.com/midocss/website/internal/users"
	"github.com/midocss/website/pkg/apperr"
)

type UserHandler struct {
	service users.Service
}

func NewUserHandler(service users.Service) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) List(c *gin.Context) {
	var query users.ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, apperr.BadRequest("invalid query parameters").WithCause(err))
		return
	}

	items, total, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}

	page, perPage := query.Normalized()
	response.Paginated(c, http.StatusOK, items, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
	})
}

func (h *UserHandler) Get(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}

	user, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, user)
}

// Create provisions a staff or customer account from the dashboard.
func (h *UserHandler) Create(c *gin.Context) {
	var in users.CreateUserInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	user, err := h.service.Create(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusCreated, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}

	var in users.UpdateUserInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	actorID, _ := middleware.UserID(c)
	user, err := h.service.Update(c.Request.Context(), actorID, id, in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}

	actorID, _ := middleware.UserID(c)
	if err := h.service.Delete(c.Request.Context(), actorID, id); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// SetPermissions replaces the per-user permission overrides.
func (h *UserHandler) SetPermissions(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}

	var in users.PermissionOverrideInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	user, err := h.service.SetPermissionOverrides(c.Request.Context(), id, in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, user)
}

func (h *UserHandler) ListRoles(c *gin.Context) {
	roles, err := h.service.Roles(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, roles)
}

func (h *UserHandler) ListPermissions(c *gin.Context) {
	permissions, err := h.service.Permissions(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, permissions)
}

func pathUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		return uuid.Nil, apperr.BadRequest("invalid " + param).WithCause(err)
	}
	return id, nil
}
