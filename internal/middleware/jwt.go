package middleware

import (
	"net/http"
	"strings"

	"github.com/saviru/spidey/internal/auth"
	"github.com/saviru/spidey/pkg/router"
)

const UserIDKey = "user_id"

// RequireJWT is a native Spidey middleware
func RequireJWT(secretKey []byte) router.Middleware {
	return func(c *router.Context, next func()) {
		// 1. Extract the Authorization header
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(c.Writer, "Missing Authorization header", http.StatusUnauthorized)
			return // stop execution
		}

		// 2. Ensure it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(c.Writer, "Invalid Authorization header format", http.StatusUnauthorized)
			return // stop execution
		}
		tokenString := parts[1]

		// 3. Validate the token
		claims, err := auth.ValidateToken(tokenString, secretKey)
		if err != nil {
			http.Error(c.Writer, "Invalid or expired token", http.StatusUnauthorized)
			return // stop execution
		}

		// 4. Inject the UserID into the request context natively
		// (Assuming your developers can fetch from context or you can add a Set/Get to your Context struct!)
		c.Request.Header.Set("X-User-ID", claims.UserID)

		// 5. Call the next handler in the chain
		next()
	}
}
