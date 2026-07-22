package core

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	json "github.com/goccy/go-json"
)

type Middleware func(*Context, func())

// WrapStd natively adapts standard Go middlewares (func(http.Handler) http.Handler) into Spidey's format
func WrapStd(std func(http.Handler) http.Handler) Middleware {
	return func(c *Context, next func()) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Writer = w
			c.Request = r
			next()
		})

		wrappedHandler := std(nextHandler)
		wrappedHandler.ServeHTTP(c.Writer, c.Request)
	}
}

type Route struct {
	Name string
	Path string
}

func (r *Route) SetName(name string) {
	r.Name = name
}

type RouterGroup struct {
	prefix      string
	app         *App
	middlewares []Middleware
}

type App struct {
	RouterGroup
	mux               *http.ServeMux
	registered        map[string]bool
	Config            Config
	globalMiddlewares []func(http.Handler) http.Handler
	WS                *WSHub
}

func New() *App {
	app := &App{
		mux:        http.NewServeMux(),
		registered: make(map[string]bool),
		WS:         NewWSHub(),
	}

	// Default config
	app.Config.Port = 3000
	app.Config.Directories.PublicDir = "public"
	app.Config.Directories.OutputDir = "bin/server"

	if file, err := os.Open("spidey.config.json"); err == nil {
		defer file.Close()
		json.NewDecoder(file).Decode(&app.Config)
	}

	app.RouterGroup = RouterGroup{
		prefix:      "",
		app:         app,
		middlewares: nil,
	}
	return app
}

func (g *RouterGroup) Group(prefix string, middlewares ...any) *RouterGroup {
	var parsedMiddlewares []Middleware
	for _, m := range middlewares {
		switch v := m.(type) {
		case Middleware:
			parsedMiddlewares = append(parsedMiddlewares, v)
		case func(*Context, func()):
			parsedMiddlewares = append(parsedMiddlewares, v)
		case func(http.Handler) http.Handler:
			parsedMiddlewares = append(parsedMiddlewares, WrapStd(v))
		default:
			panic("Error: Unsupported middleware type. Must be Spidey Middleware or standard func(http.Handler) http.Handler")
		}
	}

	newMiddlewares := append([]Middleware{}, g.middlewares...)
	newMiddlewares = append(newMiddlewares, parsedMiddlewares...)

	return &RouterGroup{
		prefix:      g.prefix + prefix,
		app:         g.app,
		middlewares: newMiddlewares,
	}
}

func (g *RouterGroup) Handle(method, path string, handlers ...any) *Route {
	fullPath := g.prefix + path

	if fullPath == "/" {
		fullPath = "/{$}"
	}

	route := fullPath
	if method != "ANY" && method != "" {
		route = method + " " + fullPath
	}

	if g.app.registered[route] {
		return &Route{Path: fullPath} // Skip duplicate route
	}
	g.app.registered[route] = true

	// Parse all route-specific handlers dynamically
	var routeMiddlewares []Middleware
	for _, m := range handlers {
		switch v := m.(type) {
		case Middleware:
			routeMiddlewares = append(routeMiddlewares, v)
		case func(*Context, func()):
			routeMiddlewares = append(routeMiddlewares, v)
		case func(http.Handler) http.Handler:
			routeMiddlewares = append(routeMiddlewares, WrapStd(v))
		case func(*Context):
			// Automatically adapt standard final handlers
			routeMiddlewares = append(routeMiddlewares, func(c *Context, next func()) {
				v(c)
				next()
			})
		default:
			panic("Spidey Error: Unsupported handler type passed to route.")
		}
	}

	// combine Group middlewares with route-specific middlewares
	totalMiddlewares := append([]Middleware{}, g.middlewares...)
	totalMiddlewares = append(totalMiddlewares, routeMiddlewares...)

	g.app.mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{Writer: w, Request: r, App: g.app}

		index := 0
		var next func()
		next = func() {
			if index < len(totalMiddlewares) {
				index++
				totalMiddlewares[index-1](ctx, next)
			}
		}
		next()
	})

	return &Route{Path: fullPath}
}

// Proxy routes everything under a path to a microservice
func (g *RouterGroup) Proxy(path string, targetURL string) {
	target, err := url.Parse(targetURL)
	if err != nil {
		panic("invalid target URL for proxy: " + targetURL)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Override the Director to rewrite the Host header,
	// preventing 403 Forbidden errors from Cloudflare/external APIs.
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	fullPath := g.prefix + path
	if !strings.HasSuffix(fullPath, "/") {
		fullPath += "/"
	}
	trimPrefix := fullPath[:len(fullPath)-1]

	g.Handle("ANY", fullPath, func(c *Context) {
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, trimPrefix)
		proxy.ServeHTTP(c.Writer, c.Request)
	})
}

func (g *RouterGroup) GET(path string, handlers ...any) *Route {
	return g.Handle(http.MethodGet, path, handlers...)
}
func (g *RouterGroup) POST(path string, handlers ...any) *Route {
	return g.Handle(http.MethodPost, path, handlers...)
}
func (g *RouterGroup) PUT(path string, handlers ...any) *Route {
	return g.Handle(http.MethodPut, path, handlers...)
}
func (g *RouterGroup) DELETE(path string, handlers ...any) *Route {
	return g.Handle(http.MethodDelete, path, handlers...)
}
func (g *RouterGroup) PATCH(path string, handlers ...any) *Route {
	return g.Handle(http.MethodPatch, path, handlers...)
}
func (g *RouterGroup) HEAD(path string, handlers ...any) *Route {
	return g.Handle(http.MethodHead, path, handlers...)
}
func (g *RouterGroup) OPTIONS(path string, handlers ...any) *Route {
	return g.Handle(http.MethodOptions, path, handlers...)
}
func (g *RouterGroup) CONNECT(path string, handlers ...any) *Route {
	return g.Handle(http.MethodConnect, path, handlers...)
}
func (g *RouterGroup) TRACE(path string, handlers ...any) *Route {
	return g.Handle(http.MethodTrace, path, handlers...)
}
func (g *RouterGroup) ANY(path string, handlers ...any) *Route {
	return g.Handle("ANY", path, handlers...)
}

func (a *App) Static(prefix, dir string) {
	a.mux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(dir))))
}

// UseGlobal applies standard Go middlewares to the root multiplexer.
func (a *App) UseGlobal(middleware func(http.Handler) http.Handler) {
	a.globalMiddlewares = append(a.globalMiddlewares, middleware)
}
