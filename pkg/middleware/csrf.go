package middleware

import (
	"net/http"

	"github.com/saviru/spidey/pkg/core"
)

func CSRF() core.Middleware {
	return func(c *core.Context, next func()) {
		method := c.Request.Method

		// Safe methods do not require CSRF protection
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions || method == http.MethodTrace {
			next()
			return
		}

		// Verify the presence of the custom header
		if c.Request.Header.Get("X-Spidey-Request") != "true" {
			c.JSON(http.StatusForbidden, map[string]string{
				"error": "CSRF verification failed. Missing X-Spidey-Request header.",
			})
			return
		}

		next()
	}
}
