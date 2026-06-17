//go:build ignore

package router

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

func (c *Context) JSON(status int, data interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	json.NewEncoder(c.Writer).Encode(data)
}

func (c *Context) Param(key string) string {
	return c.Request.PathValue(key)
}

func (c *Context) Send(content string) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.Write([]byte(content))
}

type Middleware func(*Context, func())

type RouterGroup struct {
	prefix      string
	app         *App
	middlewares []Middleware
}

type App struct {
	RouterGroup
	mux        *http.ServeMux
	registered map[string]bool
}

func New() *App {
	app := &App{
		mux:        http.NewServeMux(),
		registered: make(map[string]bool),
	}
	app.RouterGroup = RouterGroup{
		prefix:      "",
		app:         app,
		middlewares: nil,
	}
	return app
}

func (g *RouterGroup) Group(prefix string, middlewares ...Middleware) *RouterGroup {
	newMiddlewares := append([]Middleware{}, g.middlewares...)
	newMiddlewares = append(newMiddlewares, middlewares...)

	return &RouterGroup{
		prefix:      g.prefix + prefix,
		app:         g.app,
		middlewares: newMiddlewares,
	}
}

func (g *RouterGroup) Handle(method, path string, handler func(*Context)) {
	fullPath := g.prefix + path

	if fullPath == "/" {
		fullPath = "/{$}"
	} else if fullPath != "/" && strings.HasSuffix(fullPath, "/") && !strings.HasSuffix(fullPath, "/{$}") {
		// Clean trailing slash for exact matches unless it's a directory proxy
		// Actually, ServeMux handles trailing slashes as sub-tree matches.
	}

	route := fullPath
	if method != "ANY" && method != "" {
		route = method + " " + fullPath
	}

	if g.app.registered[route] {
		return // Skip duplicate route so manual routes can override auto routes
	}
	g.app.registered[route] = true

	g.app.mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{Writer: w, Request: r}

		index := 0
		var next func()
		next = func() {
			if index < len(g.middlewares) {
				index++
				g.middlewares[index-1](ctx, next)
			} else {
				handler(ctx)
			}
		}
		next()
	})
}

// Proxy routes everything under a path to a microservice
func (g *RouterGroup) Proxy(path string, targetURL string) {
	target, err := url.Parse(targetURL)
	if err != nil {
		panic("invalid target URL for proxy: " + targetURL)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

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

func (g *RouterGroup) GET(path string, handler func(*Context)) { g.Handle(http.MethodGet, path, handler) }
func (g *RouterGroup) POST(path string, handler func(*Context)) { g.Handle(http.MethodPost, path, handler) }
func (g *RouterGroup) PUT(path string, handler func(*Context)) { g.Handle(http.MethodPut, path, handler) }
func (g *RouterGroup) DELETE(path string, handler func(*Context)) { g.Handle(http.MethodDelete, path, handler) }

func (a *App) Listen(port string) error { return http.ListenAndServe(":"+port, a.mux) }

func (a *App) Static(prefix, dir string) {
	a.mux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(dir))))
}
