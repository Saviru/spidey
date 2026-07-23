package core

import (
	"log"
	"net/http"
	"runtime/debug"
)

// intercepts panics, logs the stack trace
// returns a graceful 500 Internal Server Error without crashing the server
func Recover() Middleware {
	return func(c *Context, next func()) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[Spidey Panic Recovery] %v\n%s\n", err, debug.Stack())
				if c.Writer != nil {
					http.Error(c.Writer, "500 Internal Server Error", http.StatusInternalServerError)
				}
			}
		}()
		next()
	}
}
