package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/saviru/spidey/pkg/auth"
	"github.com/saviru/spidey/pkg/core"
)

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Cookie
	if cookie, err := r.Cookie("jwt"); err == nil {
		return cookie.Value
	}

	return r.URL.Query().Get("token")
}

// Validates a JWT token using a static secret key.
// Extracts the token from the Authorization header, a "jwt" cookie, or a "token" query string.
func RequireJWT(secretKey []byte) core.Middleware {
	return func(c *core.Context, next func()) {
		tokenString := extractToken(c.Request)
		if tokenString == "" {
			http.Error(c.Writer, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(tokenString, secretKey)
		if err != nil {
			http.Error(c.Writer, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Inject the entire claims struct into the request context
		c.Set("user", claims)

		next()
	}
}

// validates a JWT token using a dynamic key lookup function.
func RequireJWTDynamic(keyFunc jwt.Keyfunc) core.Middleware {
	return func(c *core.Context, next func()) {
		tokenString := extractToken(c.Request)
		if tokenString == "" {
			http.Error(c.Writer, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateTokenDynamic(tokenString, keyFunc)
		if err != nil {
			http.Error(c.Writer, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		c.Set("user", claims)

		next()
	}
}
