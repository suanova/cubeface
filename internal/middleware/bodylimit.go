package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/cubeface/cubeface/internal/common"
)

const maxBodyBytes = 1 << 20 // 1 MiB

// LimitBody caps request bodies and, when a body is present, requires JSON for
// API calls.
func LimitBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBodyBytes {
			common.AbortFail(c, common.NewAPIErrorStatus(http.StatusRequestEntityTooLarge,
				"INVALID_REQUEST", "Request body is too large"))
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

		if c.Request.ContentLength > 0 && strings.EqualFold(c.Request.Method, http.MethodPost) {
			contentType := c.ContentType()
			if contentType != "" && contentType != "application/json" {
				common.AbortFail(c, common.NewAPIErrorStatus(http.StatusUnsupportedMediaType,
					"INVALID_REQUEST", "Content-Type must be application/json"))
				return
			}
		}
		c.Next()
	}
}

// SecurityHeaders sets conservative defaults for both API and static replies.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// Recovery converts panics into the standard error envelope.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		common.AbortFail(c, common.NewAPIErrorStatus(http.StatusInternalServerError,
			"INTERNAL_ERROR", "Internal server error"))
	})
}
