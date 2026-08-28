package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/midocss/website/internal/transport/http/response"
	"github.com/midocss/website/pkg/apperr"
)

type HealthHandler struct {
	db      *gorm.DB
	version string
}

func NewHealthHandler(db *gorm.DB, version string) *HealthHandler {
	return &HealthHandler{db: db, version: version}
}

// Live reports that the process is running.
func (h *HealthHandler) Live(c *gin.Context) {
	response.OK(c, http.StatusOK, gin.H{"status": "ok", "version": h.version})
}

// Ready additionally verifies the database connection.
func (h *HealthHandler) Ready(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		response.Fail(c, apperr.Internal(err))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		response.Fail(c, apperr.Internal(err))
		return
	}
	response.OK(c, http.StatusOK, gin.H{"status": "ready"})
}
