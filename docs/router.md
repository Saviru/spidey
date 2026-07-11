# Router API

Spidey's `hub/router` package provides a robust routing and context engine. It natively wraps Go's `http.ServeMux` but adds an ergonomic layer similar to Express or Fiber.

## The Context Object

The `router.Context` provides various helpers for handling requests and generating responses quickly.

### Sending Responses

- `c.JSON(status, data)`: Sends a standard JSON response.
- `c.JSONPretty(status, data)`: Sends pretty-printed JSON (formatted with tabs and newlines).
- `c.SecureJSON(status, data)`: Prevents JSON hijacking by prepending `"while(1);\n"` if the payload is a JSON array.
- `c.PureJSON(status, data)`: Sends JSON without Go automatically escaping HTML characters (like `<` and `>`).
- `c.JSONP(status, data, callback)`: Wraps the JSON response inside a Javascript callback function.
- `c.Send("text")`: Sends raw text or HTML string.
- `c.HTML(status, "html")`: Sends an HTML fragment with the proper `text/html` content type. Extremely useful when returning UI updates for `s-tags`.

### Parsing Requests

- `c.BindJSON(&obj)`: Parses incoming JSON request bodies directly into a Go struct. It automatically validates the struct using the `validator/v10` package.
- `c.Param("key")`: Gets a URL path parameter (e.g., from `/users/{id}`).
- `c.Query("key")`: Gets a URL query parameter. Spidey caches the parsed URL on the first call for performance.
- `c.DefaultQuery("key", default)`: Gets a query parameter, falling back to a default value if missing.
- `c.QueryInt("key", default)`: Safely auto-converts a query parameter into an integer.
- `c.QueryBool("key")`: Safely evaluates boolean query parameters (recognizes `"true"`, `"1"`, or `"yes"`).

## Routing and Groups

You can define routes and group them with shared prefixes and middlewares. Spidey supports all standard HTTP methods: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`, `OPTIONS`, `CONNECT`, `TRACE`, and `ANY`.

```go
app.GET("/health", func(c *router.Context) {
    c.Send("OK")
})

api := app.Group("/api")
api.POST("/login", LoginHandler)
```

## Middlewares

Middlewares can be applied globally, to route groups, or to specific routes. Spidey's custom router natively supports three types of middleware formats:

1. **Spidey Middleware**: `func(*router.Context, func())`
2. **Standard Go Middleware**: `func(http.Handler) http.Handler`
3. **Standard Handlers**: `func(*router.Context)`

Spidey automatically wraps standard Go middlewares using the `router.WrapStd()` function, allowing you to use existing ecosystem middlewares out-of-the-box!

```go
// Using multiple middlewares on a group
app.Group("/api", LoggerMiddleware, AuthCheck)
```

## Reverse Proxy

You can easily proxy requests under a specific path to another microservice using the built-in `Proxy` method.

```go
// Forwards all traffic from /api/users/* to http://localhost:8080/*
app.Group("/api").Proxy("/users", "http://localhost:8080")
```

## Serving Static Assets

You can manually serve static folders using `Static`:

```go
app.Static("/assets", "public/assets")
```
