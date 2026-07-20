# Authentication in Spidey

Spidey provides robust, built-in support for both JWT (JSON Web Token) authentication and OAuth2 (e.g., Login with Google, GitHub, etc.) out-of-the-box.

---

## OAuth2 (Login with Google, GitHub, etc.)

Spidey includes a generic OAuth2 module (`pkg/auth`) that wraps the official `golang.org/x/oauth2` package, making it incredibly easy to add social logins to your application.

### 1. Creating a Provider
First, initialize your provider using your Client ID and Secret (usually stored in environment variables).

```go
import "github.com/saviru/spidey/pkg/auth"

var githubProvider = auth.NewGitHubProvider(
    "YOUR_CLIENT_ID",
    "YOUR_CLIENT_SECRET",
    "http://localhost:3000/api/auth/github/callback",
    nil, // Uses default scopes if nil
)
```

### 2. The Login Route
Create an API route that generates a secure state, sets it in a cookie (for CSRF protection), and redirects the user to the provider.

```go
//spidey:route GET /auth/github/login
func GithubLogin(c *core.Context) {
    state := auth.GenerateState()
    c.SetCookie(&http.Cookie{Name: "oauthstate", Value: state, HttpOnly: true, Path: "/"})
    c.Redirect(githubProvider.GetLoginURL(state))
}
```

### 3. The Callback Route
Create the callback route to exchange the code for a token and fetch the user's profile.

```go
//spidey:route GET /auth/github/callback
func GithubCallback(c *core.Context) {
    // 1. Verify state cookie prevents CSRF
    cookieState, err := c.Cookie("oauthstate")
    if err != nil || cookieState != c.Query("state") {
        http.Error(c.Writer, "Invalid state", http.StatusBadRequest)
        return
    }

    // 2. Exchange code for an OAuth2 token
    token, err := githubProvider.Exchange(c.Request.Context(), c.Query("code"))
    if err != nil {
        http.Error(c.Writer, "Failed to exchange token", http.StatusInternalServerError)
        return
    }

    // 3. Fetch User Info
    userInfo, err := githubProvider.FetchUserInfo(c.Request.Context(), token)
    if err != nil {
        http.Error(c.Writer, "Failed to fetch user info", http.StatusInternalServerError)
        return
    }

    // 4. Generate a Spidey JWT for this user and log them in!
    userID := userInfo["id"].(float64) // GitHub returns numeric IDs
    spideyToken, _ := auth.GenerateToken(fmt.Sprintf("%v", userID), nil, []byte("my-secret-key"), time.Hour*24)
    c.SetCookie(&http.Cookie{Name: "jwt", Value: spideyToken, HttpOnly: true, Path: "/"})
    
    c.Redirect("/")
}
```

---

## JWT Authentication

Spidey comes with a robust, built-in JWT (JSON Web Token) authentication package out-of-the-box (`github.com/saviru/spidey/pkg/auth` and `pkg/middleware`).

### Generating Tokens
```go
import "github.com/saviru/spidey/pkg/auth"

token, err := auth.GenerateToken(userId, nil, []byte("my-secret-key"), time.Hour * 24)
```

### Refresh Tokens
You can generate both an Access Token and a long-lived Refresh Token at the same time:
```go
accessToken, refreshToken, err := auth.GenerateTokenPair(userId, nil, []byte("my-secret"), time.Minute*15, time.Hour*24*7)
```
Spidey includes a `ValidateRefreshToken` function to securely ensure someone isn't using a refresh token as an access token.

### Dynamic Secret Keys
For high-security environments, you can use dynamic secret keys (Key Rotation) based on a Key ID (`kid`) in the JWT header:
```go
// Generates a token with a custom Key ID
token, err := auth.GenerateTokenWithKID(userId, nil, []byte("dynamic-secret"), time.Hour, "key-2026")
```

### JWT Middleware
You can protect your API routes using the provided JWT middleware, which automatically extracts the token from the `Authorization` header, a `jwt` cookie, or a `token` URL query parameter.

```go
import "github.com/saviru/spidey/pkg/middleware"

//spidey:middleware middleware.RequireJWT([]byte("my-secret-key"))
//spidey:route GET /protected
func ProtectedEndpoint(c *core.Context) {
    // Route is secured!
    userClaims := c.Get("user").(*auth.SpideyClaims)
    c.JSON(200, map[string]interface{}{"user": userClaims})
}
```

For dynamic keys, use `RequireJWTDynamic`:
```go
//spidey:middleware middleware.RequireJWTDynamic(fetchKeyFunc)
```
