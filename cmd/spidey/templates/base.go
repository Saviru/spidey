//go:build ignore

package pages

import "fmt"

type Renderer func(interface{}) (string, error)

var registry = make(map[string]Renderer)

// Register is called automatically by transpiled files
func Register(name string, fn Renderer) {
	registry[name] = fn
}

// Render is the universal function the user calls in main.go
func Render(name string, data interface{}) (string, error) {
	if fn, ok := registry[name]; ok {
		return fn(data)
	}
	return "", fmt.Errorf("template '%s' not found. Did you run 'spidey build'?", name)
}
