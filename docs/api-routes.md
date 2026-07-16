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

## Built-In JWT Authentication

Spidey comes with a robust, built-in JWT (JSON Web Token) authentication package out-of-the-box (`github.com/saviru/spidey/pkg/auth` and `pkg/middleware`).

### Generating Tokens
```go
import "github.com/saviru/spidey/pkg/auth"

token, err := auth.GenerateToken(userId, []byte("my-secret-key"), time.Hour * 24)
```

### Dynamic Secret Keys
For high-security environments, you can use dynamic secret keys (Key Rotation) based on a Key ID (`kid`) in the JWT header:
```go
// Generates a token with a custom Key ID
token, err := auth.GenerateTokenDynamic(userId, "key-2026", []byte("dynamic-secret"), time.Hour)
```

### JWT Middleware
You can protect your API routes using the provided JWT middleware, which automatically extracts the token from the `Authorization` header, a `jwt` cookie, or a `token` URL query parameter.

```go
import "github.com/saviru/spidey/pkg/middleware"

//spidey:middleware middleware.RequireJWT([]byte("my-secret-key"))
//spidey:route GET /protected
func ProtectedEndpoint(c *core.Context) {
    // Route is secured!
}
```

For dynamic keys, use `RequireJWTDynamic`:
```go
middleware.RequireJWTDynamic(func(token *jwt.Token) (interface{}, error) {
    // Return the correct secret key based on the token's kid
    kid := token.Header["kid"].(string)
    return fetchSecretFromDB(kid), nil
})
```

*Note: The middleware functions (`AuthCheck`, `RateLimiter`) must be defined and exported in the `api/` package.*

## Dynamic Parameters

You can define dynamic path parameters directly in your API comments using the bracket syntax `[paramName]`. Spidey's compiler automatically transforms this into the `{paramName}` format that the internal router expects.

```go
//spidey:route GET /users/[id]
func GetUser(c *core.Context) {
    userID := c.Param("id")
    c.JSON(200, map[string]string{"user_id": userID})
}
```
