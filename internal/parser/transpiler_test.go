package parser

import (
	"strings"
	"testing"
)

// Tests for attribute parser
func TestParseComponentAttrs(t *testing.T) {
	tests := []struct {
		name         string
		attrs        string
		wantIsland   bool
		wantPipeline string
	}{
		{
			name:         "no attributes",
			attrs:        "",
			wantIsland:   false,
			wantPipeline: ".",
		},
		{
			name:         "client:load island directive only",
			attrs:        " client:load ",
			wantIsland:   true,
			wantPipeline: ".",
		},
		{
			name:         "string prop with double quotes",
			attrs:        ` title="Click Me"`,
			wantIsland:   false,
			wantPipeline: `(dict "title" "Click Me")`,
		},
		{
			name:         "string prop with single quotes",
			attrs:        ` title='Click Me'`,
			wantIsland:   false,
			wantPipeline: `(dict "title" "Click Me")`,
		},
		{
			name:         "multiple props",
			attrs:        ` title="Click Me" variant="primary"`,
			wantIsland:   false,
			wantPipeline: `(dict "title" "Click Me" "variant" "primary")`,
		},
		{
			name:         "client:load with props",
			attrs:        ` client:load title="Click Me"`,
			wantIsland:   true,
			wantPipeline: `(dict "title" "Click Me")`,
		},
		{
			name:         "dynamic template expression",
			attrs:        ` title="{{ .title }}"`,
			wantIsland:   false,
			wantPipeline: `(dict "title" .title)`,
		},
		{
			name:         "pipeline reference without curlies",
			attrs:        ` title=.title`,
			wantIsland:   false,
			wantPipeline: `(dict "title" .title)`,
		},
		{
			name:         "boolean prop flag",
			attrs:        ` disabled`,
			wantIsland:   false,
			wantPipeline: `(dict "disabled" true)`,
		},
		{
			name:         "numeric prop",
			attrs:        ` count=5`,
			wantIsland:   false,
			wantPipeline: `(dict "count" 5)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsland, gotPipeline := parseComponentAttrs(tt.attrs)
			if gotIsland != tt.wantIsland {
				t.Errorf("parseComponentAttrs(%q) gotIsland = %v, want %v", tt.attrs, gotIsland, tt.wantIsland)
			}
			if gotPipeline != tt.wantPipeline {
				t.Errorf("parseComponentAttrs(%q) gotPipeline = %q, want %q", tt.attrs, gotPipeline, tt.wantPipeline)
			}
		})
	}
}

// Tests for components with props
func TestTranspileToGoComponentProps(t *testing.T) {
	html := `<div><Button title="Click Me" /></div>`
	goCode, err := TranspileToGo("testmod", "index", html, "", "", nil)
	if err != nil {
		t.Fatalf("TranspileToGo returned an error: %v", err)
	}

	// Verify template call contains dict args
	expectedTemplateCall := `{{template "Button" (dict "title" "Click Me")}}`
	if !strings.Contains(goCode, expectedTemplateCall) {
		t.Errorf("Expected generated code to contain %q, but got:\n%s", expectedTemplateCall, goCode)
	}

	// Verify Funcs(spidey.FuncMap) is registered before Parse
	expectedFuncMap := `template.New("index").Funcs(spidey.FuncMap).Parse(`
	if !strings.Contains(goCode, expectedFuncMap) {
		t.Errorf("Expected generated code to register spidey.FuncMap, but got:\n%s", goCode)
	}
}

// verify backward compatibility when components have NO props
func TestTranspileToGoComponentNoProps(t *testing.T) {
	html := `<div><Button /></div>`
	goCode, err := TranspileToGo("testmod", "index", html, "", "", nil)
	if err != nil {
		t.Fatalf("TranspileToGo returned an error: %v", err)
	}

	// Verify components without props still receive dot context '.'
	expectedTemplateCall := `{{template "Button" .}}`
	if !strings.Contains(goCode, expectedTemplateCall) {
		t.Errorf("Expected generated code to contain %q, but got:\n%s", expectedTemplateCall, goCode)
	}
}
