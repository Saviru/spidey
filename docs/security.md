# Security in Spidey

Spidey is built with a **Secure-by-Default** philosophy. It provides native, framework-level protections against common web application vulnerabilities, including Cross-Site Scripting (XSS), Cross-Site Request Forgery (CSRF), Cross-Site WebSocket Hijacking (CSWSH), Arbitrary Code Execution (ACE), and Denial of Service (DoS) attacks.

---

## 1. Cross-Site Scripting (XSS) Protections

Spidey employs multiple layers of defense to prevent XSS at the framework, client, and API levels.

### Framework-Level HTML Escaping (`HTMLf` and `Sendf`)
When rendering dynamic variables inside HTML templates, developers often forget to escape user input. Spidey provides format-based response helpers that automatically escape string arguments:
- `c.HTMLf(status, format, args...)`
- `c.Sendf(format, args...)`

These methods format the string using `fmt.Sprintf` but automatically run Go's `html.EscapeString()` on any `string` variables passed in `args`.

```go
// Secure: user input is automatically escaped!
c.HTMLf(200, "<div>Welcome %s!</div>", c.Query("name"))
```

### Client-Side DOM Sanitization
Spidey's client-side S-Tags engine (`s-get`, `s-post`, and WebSockets) dynamically updates the DOM. To prevent DOM-based XSS from malicious server payloads, Spidey automatically sanitizes all incoming HTML strings before injecting them into the DOM via `innerHTML`, `outerHTML`, or `insertAdjacentHTML`.

The sanitization engine recursively strips out:
- `<script>` and `<iframe>` elements.
- Inline event handlers (e.g., `onload`, `onerror`, `onclick`).
- Unsafe URI schemes like `javascript:` inside attributes (e.g., `<a href="javascript:...">`).

### JSONP Callback Sanitization
If you use the built-in `c.JSONP()` response method, Spidey protects against Reflected XSS via callback parameter injection. It uses a strict regular expression to strip out any character from the callback parameter that is not a valid JavaScript identifier (`[a-zA-Z0-9_\.]`). If the resulting callback becomes empty, it falls back to a safe default literal `"callback"`.

---

## 2. Cross-Site Request Forgery (CSRF) Protection

Spidey natively implements the **OWASP Custom Request Header Defense** mechanism:
- **Client-Side**: The S-Tags engine automatically attaches a custom header `X-Spidey-Request: true` to all outgoing AJAX fetch requests. This forces modern browsers to trigger a CORS preflight (`OPTIONS` request) for cross-origin requests, blocking them before execution unless explicitly permitted.
- **Server-Side**: Apply the built-in `middleware.CSRF()` middleware to your state-changing routes. It will verify the presence of the `X-Spidey-Request` header on non-safe methods (POST, PUT, DELETE, PATCH) and reject unauthenticated cross-site requests with a `403 Forbidden` status.

```go
//spidey:middleware middleware.CSRF()
//spidey:route POST /api/sensitive-action
func SensitiveAction(c *core.Context) {
    c.JSON(200, map[string]string{"status": "protected"})
}
```

---

## 3. Cross-Site WebSocket Hijacking (CSWSH)

Spidey's WebSocket server is secure by default:
- **Same-Origin Check**: The connection upgrader enforces a strict Same-Origin check. If the browser's `Origin` header does not match your server's host, the connection is instantly rejected.
- **Configurable Origins**: If you require cross-origin WebSockets (e.g., for a public API or a mobile app), you can explicitly specify permitted origins in your `spidey.config.json` file:

```json
{
    "wsAllowedOrigins": [
        "https://app.example.com",
        "https://admin.example.com"
    ]
}
```
*Note: You can use `*` to allow all origins, but do so with caution.*

---

## 4. CORS (Cross-Origin Resource Sharing) Security

Spidey's CORS middleware enforces correct specs and prevents caching issues:
- **Spec Enforcement**: The CORS standard forbids setting `Access-Control-Allow-Origin: *` when `Access-Control-Allow-Credentials` is `true`. If you configure this combination, Spidey will log a warning at startup and automatically fallback to echoing the requesting client's `Origin` header instead.
- **Cache Poisoning Prevention**: When reflecting specific origins dynamically, Spidey automatically sets the `Vary: Origin` header. This tells CDNs, reverse proxies, and browsers to cache separate copies of the response for different origins, preventing cache poisoning and accidental DoS.
- **Safe Allowed Headers**: To prevent request header spoofing/bypassing (e.g., custom admin tokens, IP forwarding header injections), Spidey no longer echoes back the client-requested `Access-Control-Request-Headers` when `AllowHeaders` is left empty. Instead, it enforces a secure default list of common safe headers (`Content-Type`, `Authorization`, `Accept`, `Accept-Language`, `Content-Language`, and `X-Spidey-Request`).

---

## 5. Runtime Sandboxing (ACE / SSTI)

Spidey supports frontmatter Go execution in `.spidey` layout files. To prevent Server-Side Template Injection (SSTI) and Arbitrary Code Execution (ACE) where a layout file could execute malicious system commands:
- Spidey runs Go frontmatter in a secure, restricted `yaegi` interpreter sandbox.
- System-level operations (like starting shell commands, reading environment variables, or accessing filesystem utilities) are locked down and disabled.

---

## 6. Resilience (Panic Recovery)

In Go, an unhandled panic in a goroutine or route handler can crash the entire application process. Spidey registers a global `Recover()` middleware by default. If a route handler panics (e.g., due to a nil pointer dereference or slice out of bounds), Spidey:
- Gracefully catches the panic.
- Logs the stack trace.
- Sends a clean `500 Internal Server Error` response back to the client, ensuring the rest of the server stays up and healthy.

---

## 7. Rate Limiting Spoof Protection

The built-in `RateLimit` middleware is secure against IP spoofing:
- **Same IP Check**: By default, the rate limiter uses the raw TCP connection's remote IP address (`RemoteAddr`). It ignores the `X-Forwarded-For` header to prevent clients from spoofing their IP.
- **Proxy Mode**: If you run Spidey behind a trusted reverse proxy (like Nginx or Cloudflare), you must explicitly set `TrustProxy: true` in your `RateLimitConfig` so the limiter knows it can safely read the proxy header.
