package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type CORSConfig struct {
	AllowOrigins     []string // List of allowed origins, or "*" for all
	AllowMethods     []string // List of allowed methods (e.g. GET, POST, OPTIONS)
	AllowHeaders     []string // List of allowed headers (e.g. Content-Type, Authorization)
	ExposeHeaders    []string // List of headers exposed to the client
	AllowCredentials bool     // Whether to allow cookies and credentials
	MaxAge           int      // Max seconds the preflight response can be cached
}

// Use it via app.UseGlobal(middleware.CORS(config)) to ensure it catches OPTIONS requests.
func CORS(config CORSConfig) func(http.Handler) http.Handler {
	// Pre-process configs
	allowAllOrigins := false
	for _, o := range config.AllowOrigins {
		if o == "*" {
			allowAllOrigins = true
			break
		}
	}

	if allowAllOrigins && config.AllowCredentials {
		log.Println("[Warning] CORS configuration has AllowOrigins=['*'] and AllowCredentials=true. This is a CORS spec violation. Spidey will automatically echo the requesting Origin header instead of '*' to prevent browser rejection.")
	}

	methodsStr := strings.Join(config.AllowMethods, ", ")
	if methodsStr == "" {
		methodsStr = "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD"
	}

	if len(config.AllowHeaders) == 0 {
		config.AllowHeaders = []string{"Content-Type", "Authorization", "Accept", "Accept-Language", "Content-Language", "X-Spidey-Request"}
	}

	headersStr := strings.Join(config.AllowHeaders, ", ")
	exposeStr := strings.Join(config.ExposeHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if we should add CORS headers
			if origin != "" {
				// Set Access-Control-Allow-Origin
				if allowAllOrigins {
					if config.AllowCredentials {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Add("Vary", "Origin")
					} else {
						w.Header().Set("Access-Control-Allow-Origin", "*")
					}
				} else {
					for _, allowedOrigin := range config.AllowOrigins {
						if allowedOrigin == origin {
							w.Header().Set("Access-Control-Allow-Origin", origin)
							w.Header().Add("Vary", "Origin")
							break
						}
					}
				}

				// Set Credentials
				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				// Expose Headers
				if exposeStr != "" {
					w.Header().Set("Access-Control-Expose-Headers", exposeStr)
				}
			}

			// Handle Preflight OPTIONS request
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", methodsStr)

				if headersStr != "" {
					w.Header().Set("Access-Control-Allow-Headers", headersStr)
				}

				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
				}

				// Terminate preflight request with 204 No Content
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
