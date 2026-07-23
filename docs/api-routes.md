# API Routes

You can build your backend directly in the `api/` directory using standard Go files. Spidey provides a powerful "Magic Comments" system to auto-generate and register your endpoints without cluttering a central routing file.

## Defining Routes

Spidey scans all `.go` files inside the `api/` directory (excluding `main.go`) for special comment directives starting with `//spidey:`.

### Dynamic Package Discovery (MVC Support)
You are not restricted to placing all your files directly in the `api/` folder! You can create subdirectories to organize your codebase into a robust **Model-View-Controller (MVC)** architecture.

For example, you can create `api/controllers`, `api/services`, or `api/models`. Spidey dynamically crawls the entire `api/` directory, discovers your Go packages, generates the correct import paths, and securely wires up all your routes and middlewares without any manual configuration!

To define a route, use the `//spidey:route` directive followed by the HTTP Method and the path. The function immediately following the comments will be bound to that route.

```go
package api

import "github.com/saviru/spidey/pkg/core"

//spidey:route GET /users
func GetUsers(c *core.Context) {
    c.JSON(200, map[string]string{"status": "ok"})
}
```

## Using Middlewares

You can attach middlewares specifically to an API route using the `//spidey:middleware` directive. You can stack multiple middlewares by adding multiple lines!

```go
//spidey:middleware AuthCheck
//spidey:middleware RateLimiter
//spidey:route POST /users
func CreateUser(c *core.Context) {
    // Both AuthCheck and RateLimiter will run before CreateUser
}
```

*Note: The middleware functions (`AuthCheck`, `RateLimiter`) must be defined and exported in the `api/` package.*

## Global Middlewares & CORS

Some middlewares, like Cross-Origin Resource Sharing (CORS), need to intercept requests *before* the router matches them to a specific route. For example, browsers send `OPTIONS` preflight requests that would otherwise be rejected.

You can apply standard Go middlewares globally to your entire application using `app.UseGlobal()` in your `main.go`. Spidey includes a built-in CORS middleware for this purpose.

```go
package main

import (
    "github.com/saviru/spidey/pkg/core"
    "github.com/saviru/spidey/pkg/middleware"
)

func main() {
    app := core.New()

    // Enable CORS globally
    app.UseGlobal(middleware.CORS(middleware.CORSConfig{
        AllowOrigins:     []string{"*"}, // Or specific domains: {"https://example.com"}
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
        MaxAge:           86400, // 24 hours caching for preflight requests
    }))

    app.Listen()
}
```

## Redis Caching

Spidey includes a built-in, high-performance Redis caching middleware. You can use it to wrap heavy API routes so their responses are served instantly from memory instead of hitting your database repeatedly!

### 1. Initialization
First, initialize the Redis connection pool in your `main.go` file before starting the server:

```go
package main

import "github.com/saviru/spidey/pkg/middleware"

func main() {
    // Connect to local Redis, empty password, database 0
    middleware.InitRedis("localhost:6379", "", 0)
    
    // ... start your app
}
```

### 2. Caching a Route
Use the `//spidey:middleware middleware.Cache(ttl)` directive to cache a route. 

```go
import "time"

//spidey:middleware middleware.Cache(time.Minute * 5)
//spidey:route GET /api/heavy-data
func GetHeavyData(c *core.Context) {
    // This route will only execute once every 5 minutes!
    // All other requests are intercepted and served instantly from Redis.
    time.Sleep(2 * time.Second) // Simulate heavy DB query
    c.JSON(200, map[string]string{"data": "very heavy data"})
}
```
*Note: The caching middleware automatically caches the status code, headers, and body, and generates a unique cache key based on the URL path and query parameters.*

### 3. Cache Invalidation
When a user updates a resource, you often want to clear the old cache. Spidey provides a handy helper function:

```go
//spidey:route POST /api/heavy-data
func UpdateHeavyData(c *core.Context) {
    // ... save new data to DB ...
    
    // Instantly bust the cache for the GET route!
    middleware.InvalidateCache("/api/heavy-data")
    
    c.JSON(200, map[string]string{"status": "updated and cache cleared!"})
}
```

## Rate Limiting

Spidey includes an enterprise-grade Distributed Rate Limiting Middleware. It uses Redis for cross-server synchronization, but gracefully falls back to a high-performance in-memory Token Bucket (`golang.org/x/time/rate`) if Redis is unavailable.

### Basic Usage
The most common use case is limiting requests per IP address. Spidey injects standard `X-RateLimit-*` HTTP headers automatically!

> **Security Note (Reverse Proxies)**: By default, Spidey securely uses the raw TCP connection IP (`RemoteAddr`) and ignores headers like `X-Forwarded-For` to prevent trivial IP spoofing attacks. If you are running Spidey behind a trusted reverse proxy (like Nginx, AWS, or Cloudflare), you must explicitly set `TrustProxy: true` in your `RateLimitConfig` so the limiter safely extracts the client IP from the headers.

```go
import (
    "time"
    "github.com/saviru/spidey/pkg/middleware"
)

// Limit to 5 requests per 10 seconds per IP
//spidey:middleware middleware.RateLimit(middleware.RateLimitConfig{MaxRequests: 5, Window: 10 * time.Second})
//spidey:route GET /api/fragile-endpoint
func GetFragileData(c *core.Context) {
    c.JSON(200, map[string]string{"data": "You are within limits!"})
}
```

## CSRF Protection

Spidey natively supports the OWASP Custom Request Header Defense mechanism for API-driven CSRF protection.

The client-side S-Tags engine automatically attaches an `X-Spidey-Request: true` header to all `fetch()` requests. To enforce this protection on your backend routes and block malicious Cross-Site Request Forgery attempts, apply the `middleware.CSRF()` middleware to your state-changing API endpoints (POST, PUT, DELETE):

```go
//spidey:middleware middleware.CSRF()
//spidey:route POST /api/transfer-funds
func TransferFunds(c *core.Context) {
    // This route is fully protected against CSRF attacks!
    c.JSON(200, map[string]string{"status": "success"})
}
```

### Advanced Usage (Custom Keys & Bypassing)
You can completely customize how limits are applied. For example, rate limit by User ID or API Key instead of IP, and provide custom rejection messages.

```go
// Define your advanced configuration in standard Go code (for IDE support & clean DX!)
var AdvancedLimiter = middleware.RateLimit(middleware.RateLimitConfig{
    MaxRequests: 100, 
    Window: time.Minute,
    KeyFunc: func(c *core.Context) string {
        return c.Request.Header.Get("X-API-Key")
    },
    SkipFunc: func(c *core.Context) bool {
        return c.Request.RemoteAddr == "127.0.0.1" // Skip limits for localhost
    },
    RejectFunc: func(c *core.Context) {
        c.JSON(429, map[string]string{"error": "Whoa, slow down! Limit exceeded."})
    },
})

// Apply the middleware using the variable name!
//spidey:middleware AdvancedLimiter
//spidey:route GET /api/advanced-limit
func AdvancedRoute(c *core.Context) {
    c.JSON(200, map[string]string{"msg": "Success!"})
}
```

## Dynamic Parameters


You can define dynamic path parameters directly in your API comments using the bracket syntax `[paramName]`. Spidey's compiler automatically transforms this into the `{paramName}` format that the internal router expects.

```go
//spidey:route GET /users/[id]
func GetUser(c *core.Context) {
    userID := c.Param("id")
    c.JSON(200, map[string]string{"user_id": userID})
}
```
