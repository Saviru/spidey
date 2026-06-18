//go:build ignore

package pages

import "fmt"

type Renderer func(interface{}) (string, error)

var registry = make(map[string]Renderer)

func Register(name string, fn Renderer) {
	registry[name] = fn
}

func Render(name string, data interface{}) (string, error) {
	if fn, ok := registry[name]; ok {
		return fn(data)
	}
	return "", fmt.Errorf("template '%s' not found. Try again with 'spidey build'?", name)
}
