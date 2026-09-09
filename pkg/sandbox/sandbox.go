package sandbox

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"
)

var AllowedImports = map[string]bool{
	"fmt":           true,
	"strings":       true,
	"strconv":       true,
	"time":          true,
	"math":          true,
	"math/rand":     true,
	"encoding/json": true,
	"html":          true,
	"html/template": true,
}

var BlockedImports = map[string]string{
	"os":      "access to the operating system is blocked",
	"os/exec": "executing shell commands is blocked",
	"syscall": "low-level system calls are blocked",
	"unsafe":  "memory manipulation is blocked",
	"plugin":  "dynamic library loading is blocked",
	"runtime": "runtime introspection is blocked",
}

func ValidateFrontmatter(code string) error {
	fset := token.NewFileSet()
	// Wrap user code in a virtual package for AST parsing
	src := fmt.Sprintf("package sandbox\n\n%s", code)
	node, err := parser.ParseFile(fset, "", src, parser.ImportsOnly)
	if err != nil {
		return fmt.Errorf("frontmatter syntax error: %w", err)
	}
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if reason, blocked := BlockedImports[path]; blocked {
			return fmt.Errorf("security violation: import %q is forbidden (%s)", path, reason)
		}
		if !AllowedImports[path] {
			return fmt.Errorf("security violation: import %q is not in the sandbox allowlist", path)
		}
	}
	return nil
}

func MergeParams(renderData, urlParams interface{}) interface{} {
	paramMap, hasParams := urlParams.(map[string]interface{})
	if !hasParams || paramMap == nil {
		return renderData
	}
	resMap, isMap := renderData.(map[string]interface{})
	if !isMap || resMap == nil {
		return urlParams
	}

	for k, v := range paramMap {
		if _, exists := resMap[k]; !exists {
			resMap[k] = v
		}
	}
	return resMap
}
