// Package response renders the API's single JSON envelope for both successful
// and failed requests.
package response

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/midocss/website/pkg/apperr"
)

// RequestIDKey is the gin context key holding the current request id.
const RequestIDKey = "request_id"

type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

type Envelope struct {
	Success   bool   `json:"success"`
	Data      any    `json:"data,omitempty"`
	Meta      *Meta  `json:"meta,omitempty"`
	Error     *Body  `json:"error,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

type Body struct {
	Code    apperr.Code         `json:"code"`
	Message string              `json:"message"`
	Fields  []apperr.FieldError `json:"fields,omitempty"`
}

func OK(c *gin.Context, status int, data any) {
	c.JSON(status, Envelope{Success: true, Data: data, RequestID: requestID(c)})
}

func Paginated(c *gin.Context, status int, data any, meta Meta) {
	c.JSON(status, Envelope{Success: true, Data: data, Meta: &meta, RequestID: requestID(c)})
}

func NoContent(c *gin.Context) {
	c.Status(204)
}

// Fail renders err using its mapped status code. Internal errors are logged
// with their cause and never leak it to the client.
func Fail(c *gin.Context, err error) {
	appErr := apperr.From(err)
	if appErr.Code == apperr.CodeInternal {
		slog.ErrorContext(c.Request.Context(), "request failed",
			"error", appErr.Error(),
			"path", c.FullPath(),
			"method", c.Request.Method,
			"request_id", requestID(c),
		)
	}

	c.AbortWithStatusJSON(appErr.HTTPStatus(), Envelope{
		Success: false,
		Error: &Body{
			Code:    appErr.Code,
			Message: appErr.Message,
			Fields:  appErr.Fields,
		},
		RequestID: requestID(c),
	})
}

func requestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
