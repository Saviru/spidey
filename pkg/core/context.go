package core

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/gorilla/websocket"
	json "github.com/goccy/go-json"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()
var jsonpCallbackRegexp = regexp.MustCompile(`[^a-zA-Z0-9_\.]`)

type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	queryCache url.Values
	Keys       map[string]any
	App        *App // Reference to the app so we can access the WSHub
}

// Stores a new key/value pair exclusively for this context
func (c *Context) Set(key string, value any) {
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

// Returns the value for the given key (value, true)
// If value does not exist it returns (nil, false)
func (c *Context) Get(key string) (value any, exists bool) {
	if c.Keys != nil {
		value, exists = c.Keys[key]
	}
	return
}

// Returns the value for the given key if it exists, otherwise it panics
func (c *Context) MustGet(key string) any {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist in context")
}

// JSON
// JSON sends a standard JSON response
func (c *Context) JSON(status int, data interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	json.NewEncoder(c.Writer).Encode(data)
}

// automatically parses the incoming JSON request body into a struct
func (c *Context) BindJSON(obj interface{}) error {
	defer c.Request.Body.Close()
	if err := json.NewDecoder(c.Request.Body).Decode(obj); err != nil {
		return err
	}
	return validate.Struct(obj)
}

// retrieves a path parameter (e.g. /users/{id})
func (c *Context) Param(key string) string {
	return c.Request.PathValue(key)
}

// returns a raw HTML or text string
func (c *Context) Send(content string) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.Write([]byte(content))
}

// formats the JSON with tabs and newlines for easier human reading
func (c *Context) JSONPretty(status int, data interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	pretty, _ := json.MarshalIndent(data, "", "    ")
	c.Writer.Write(pretty)
}

// prevents JSON Hijacking by prepending "while(1);" if the payload is an array
func (c *Context) SecureJSON(status int, data interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	bytes, _ := json.Marshal(data)
	if len(bytes) > 0 && bytes[0] == '[' {
		c.Writer.Write([]byte("while(1);\n"))
	}
	c.Writer.Write(bytes)
}

// prevents Go from automatically escaping HTML characters (like < and >) in JSON payloads
func (c *Context) PureJSON(status int, data interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	encoder := json.NewEncoder(c.Writer)
	encoder.SetEscapeHTML(false)
	encoder.Encode(data)
}

// wraps the JSON response inside a Javascript callback function
func (c *Context) JSONP(status int, data interface{}, callback string) {
	c.Writer.Header().Set("Content-Type", "application/javascript")
	c.Writer.WriteHeader(status)
	safeCallback := jsonpCallbackRegexp.ReplaceAllString(callback, "")
	if safeCallback == "" {
		safeCallback = "callback"
	}
	fmt.Fprintf(c.Writer, "%s(", safeCallback)
	json.NewEncoder(c.Writer).Encode(data)
	fmt.Fprint(c.Writer, ")")
}

// upgrades the HTTP connection to a WebSocket connection
func (c *Context) Upgrade() (*SpideyConn, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	if len(c.App.Config.WSAllowedOrigins) > 0 {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true // allow non-browser clients
			}
			for _, allowed := range c.App.Config.WSAllowedOrigins {
				if allowed == "*" || allowed == origin {
					return true
				}
			}
			return false
		}
	} // else CheckOrigin is nil, meaning gorilla/websocket enforces same-origin safely

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, err
	}

	spideyConn := &SpideyConn{
		ID:   c.Request.RemoteAddr, // Default ID, can be overridden by users
		Conn: conn,
		Hub:  c.App.WS,
		Send: make(chan []byte, 256),
	}

	go spideyConn.writePump()
	go spideyConn.readPump()

	return spideyConn, nil
}

// Query
// Parses the URL once and caches it for all future calls
func (c *Context) getQueryCache() url.Values {
	if c.queryCache == nil {
		c.queryCache = c.Request.URL.Query()
	}
	return c.queryCache
}

// Returns the string value for a URL query parameter
func (c *Context) Query(key string) string {
	return c.getQueryCache().Get(key)
}

// Returns the query value or a fallback default
func (c *Context) DefaultQuery(key, defaultValue string) string {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Safely auto-converts a query parameter into an integer
func (c *Context) QueryInt(key string, defaultValue int) int {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// safely evaluates boolean query parameters ("true", "1", "yes")
func (c *Context) QueryBool(key string) bool {
	val := c.Query(key)
	return val == "true" || val == "1" || val == "yes"
}

// S-Tags
// Sends a raw HTML fragment back to the Spidey Client Engine
func (c *Context) HTML(status int, htmlFragment string) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)
	c.Writer.Write([]byte(htmlFragment))
}

// formats a string according to a format specifier and sends it as HTML.
func (c *Context) HTMLf(status int, format string, args ...any) {
	for i, arg := range args {
		if s, ok := arg.(string); ok {
			args[i] = html.EscapeString(s)
		}
	}
	c.HTML(status, fmt.Sprintf(format, args...))
}

// formats a string and sends it as raw HTML or text.
func (c *Context) Sendf(format string, args ...any) {
	for i, arg := range args {
		if s, ok := arg.(string); ok {
			args[i] = html.EscapeString(s)
		}
	}
	c.Send(fmt.Sprintf(format, args...))
}

// sends an HTTP redirect to the specified URL (302 Found)
func (c *Context) Redirect(url string) {
	http.Redirect(c.Writer, c.Request, url, http.StatusFound)
}

// returns the value of the named cookie provided in the request
func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// adds a Set-Cookie header to the ResponseWriter
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}
