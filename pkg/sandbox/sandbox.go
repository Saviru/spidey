package sandbox

import (
	"fmt"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func EvalFrontmatter(script string, defaultData interface{}) (interface{}, error) {
	i := interp.New(interp.Options{
		Unrestricted: false,
	})

	i.Use(stdlib.Symbols)

	_, err := i.Eval(script)
	if err != nil {
		return nil, fmt.Errorf("Frontmatter syntax error: %v", err)
	}

	// Look for a standard Render() function defined by the user
	v, err := i.Eval("Render()")
	if err == nil && v.IsValid() {
		return v.Interface(), nil
	}

	return defaultData, nil
}
