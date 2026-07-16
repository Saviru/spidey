package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/saviru/spidey/pkg/auth"
	"github.com/saviru/spidey/pkg/core"
)

const UserIDKey = "user_id"

func RequireJWT(secretKey []byte) core.Middleware {
	return func(c *core.Context, next func()) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(c.Writer, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Ensure it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(c.Writer, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

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
// Useful for Key Rotation, where the secret key is determined by the token's "kid" header.
func RequireJWTDynamic(keyFunc jwt.Keyfunc) core.Middleware {
	return func(c *core.Context, next func()) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(c.Writer, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(c.Writer, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		claims, err := auth.ValidateTokenDynamic(tokenString, keyFunc)
		if err != nil {
			http.Error(c.Writer, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		c.Set("user", claims)

		next()
	}
}
