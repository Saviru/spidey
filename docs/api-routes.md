# API Routes

You can build your backend directly in the `api/` directory using standard Go files. Spidey provides a powerful "Magic Comments" system to auto-generate and register your endpoints without cluttering a central routing file.

## Defining Routes

Spidey scans all `.go` files inside the `api/` directory (excluding `main.go`) for special comment directives starting with `//spidey:`.

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

## Dynamic Parameters


You can define dynamic path parameters directly in your API comments using the bracket syntax `[paramName]`. Spidey's compiler automatically transforms this into the `{paramName}` format that the internal router expects.

```go
//spidey:route GET /users/[id]
func GetUser(c *core.Context) {
    userID := c.Param("id")
    c.JSON(200, map[string]string{"user_id": userID})
}
```
