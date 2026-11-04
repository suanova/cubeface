package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/cubeface/cubeface/internal/common"
)

// RequireSameOrigin protects unsafe methods against cross-site requests. It is
// a companion to SameSite=Lax cookies and does not replace a CSRF token if the
// deployment later becomes cross-origin.
func RequireSameOrigin(allowedOrigin string) gin.HandlerFunc {
	allowed := normalizeOrigin(allowedOrigin)
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")
		if origin == "" {
			if referer := c.GetHeader("Referer"); referer != "" {
				origin = originFromURL(referer)
			}
		}
		// Requests without Origin/Referer (curl, same-process clients) are
		// allowed; browsers always send Origin on cross-site unsafe requests.
		if origin != "" && normalizeOrigin(origin) != allowed {
			common.AbortFail(c, common.NewAPIErrorStatus(http.StatusForbidden,
				"CROSS_ORIGIN_REQUEST", "Cross-origin request rejected"))
			return
		}
		c.Next()
	}
}

func normalizeOrigin(raw string) string {
	raw = strings.TrimSpace(raw)
	if parsed, err := url.Parse(raw); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return strings.ToLower(parsed.Scheme + "://" + parsed.Host)
	}
	return strings.ToLower(strings.TrimRight(raw, "/"))
}

func originFromURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}
