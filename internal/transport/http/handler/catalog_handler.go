package handler

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/midocss/website/internal/catalog"
	"github.com/midocss/website/internal/transport/http/httpx"
	"github.com/midocss/website/internal/transport/http/response"
	"github.com/midocss/website/pkg/apperr"
)

// CatalogHandler serves both the dashboard CRUD routes and the public
// read-only routes; the public ones narrow the query to published/active rows.
type CatalogHandler struct {
	service catalog.Service
}

func NewCatalogHandler(service catalog.Service) *CatalogHandler {
	return &CatalogHandler{service: service}
}

func (h *CatalogHandler) ListProjectTypes(c *gin.Context) {
	h.listProjectTypes(c, false)
}

func (h *CatalogHandler) ListPublicProjectTypes(c *gin.Context) {
	h.listProjectTypes(c, true)
}

func (h *CatalogHandler) listProjectTypes(c *gin.Context, public bool) {
	query, ok := bindListQuery(c, public)
	if !ok {
		return
	}

	items, total, err := h.service.ListProjectTypes(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	respondPaginated(c, query, items, total)
}

func (h *CatalogHandler) GetProjectType(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}

	projectType, err := h.service.GetProjectType(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, projectType)
}

func (h *CatalogHandler) CreateProjectType(c *gin.Context) {
	var in catalog.CreateProjectTypeInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	projectType, err := h.service.CreateProjectType(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusCreated, projectType)
}

func (h *CatalogHandler) UpdateProjectType(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}

	var in catalog.UpdateProjectTypeInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	projectType, err := h.service.UpdateProjectType(c.Request.Context(), id, in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, projectType)
}

func (h *CatalogHandler) DeleteProjectType(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.DeleteProjectType(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

func (h *CatalogHandler) ListPortfolio(c *gin.Context) {
	h.listPortfolio(c, false)
}

func (h *CatalogHandler) ListPublicPortfolio(c *gin.Context) {
	h.listPortfolio(c, true)
}

func (h *CatalogHandler) listPortfolio(c *gin.Context, public bool) {
	query, ok := bindListQuery(c, public)
	if !ok {
		return
	}

	items, total, err := h.service.ListPortfolio(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	respondPaginated(c, query, items, total)
}

func (h *CatalogHandler) GetPortfolioProject(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}

	project, err := h.service.GetPortfolioProject(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, project)
}

func (h *CatalogHandler) GetPublicPortfolioProject(c *gin.Context) {
	project, err := h.service.GetPortfolioProjectBySlug(c.Request.Context(), c.Param("slug"), true)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, project)
}

func (h *CatalogHandler) CreatePortfolioProject(c *gin.Context) {
	var in catalog.CreatePortfolioProjectInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	project, err := h.service.CreatePortfolioProject(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusCreated, project)
}

func (h *CatalogHandler) UpdatePortfolioProject(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}

	var in catalog.UpdatePortfolioProjectInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	project, err := h.service.UpdatePortfolioProject(c.Request.Context(), id, in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, project)
}

func (h *CatalogHandler) DeletePortfolioProject(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.DeletePortfolioProject(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

func (h *CatalogHandler) ListPackages(c *gin.Context) {
	h.listPackages(c, false)
}

func (h *CatalogHandler) ListPublicPackages(c *gin.Context) {
	h.listPackages(c, true)
}

func (h *CatalogHandler) listPackages(c *gin.Context, public bool) {
	query, ok := bindListQuery(c, public)
	if !ok {
		return
	}

	items, total, err := h.service.ListPackages(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	respondPaginated(c, query, items, total)
}

func (h *CatalogHandler) GetPackage(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}

	pkg, err := h.service.GetPackage(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, pkg)
}

func (h *CatalogHandler) GetPublicPackage(c *gin.Context) {
	pkg, err := h.service.GetPackageBySlug(c.Request.Context(), c.Param("slug"), true)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, pkg)
}

func (h *CatalogHandler) CreatePackage(c *gin.Context) {
	var in catalog.CreatePackageInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	pkg, err := h.service.CreatePackage(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusCreated, pkg)
}

func (h *CatalogHandler) UpdatePackage(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}

	var in catalog.UpdatePackageInput
	if err := httpx.BindJSON(c, &in); err != nil {
		response.Fail(c, err)
		return
	}

	pkg, err := h.service.UpdatePackage(c.Request.Context(), id, in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, pkg)
}

func (h *CatalogHandler) DeletePackage(c *gin.Context) {
	id, err := pathUUID(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.DeletePackage(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

func bindListQuery(c *gin.Context, public bool) (catalog.ListQuery, bool) {
	var query catalog.ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, apperr.BadRequest("invalid query parameters").WithCause(err))
		return query, false
	}
	if public {
		query = query.Public()
	}
	return query, true
}

func respondPaginated(c *gin.Context, query catalog.ListQuery, items any, total int64) {
	page, perPage := query.Normalized()
	response.Paginated(c, http.StatusOK, items, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
	})
}
