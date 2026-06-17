package bundler

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"spidey/pkg/parser"

	"github.com/evanw/esbuild/pkg/api"
)

func ProcessPages(projectDir string, templates embed.FS, isDev bool) error {
	pagesDir := filepath.Join(projectDir, "pages")
	genDir := filepath.Join(projectDir, "lib", "pages")

	os.RemoveAll(genDir)
	os.MkdirAll(genDir, 0755)

	// Read separate base.go template and inject it safely
	baseCodeBytes, err := templates.ReadFile("templates/base.go")
	if err != nil {
		return fmt.Errorf("could not read base.go template: %v", err)
	}
	baseCode := strings.Replace(string(baseCodeBytes), "//go:build ignore", "", 1)
	baseCode = strings.TrimSpace(baseCode) + "\n"

	os.WriteFile(filepath.Join(genDir, "spidey_base.go"), []byte(baseCode), 0644)

	// User Layout Check
	appLayoutStr := ""
	appLayoutPath := filepath.Join(projectDir, "app.spidey")
	if appLayoutBytes, err := os.ReadFile(appLayoutPath); err == nil {
		appLayoutStr = string(appLayoutBytes)
		
		// Inject Livereload if in dev mode
		if isDev {
			script := `<script>const evtSource = new EventSource("http://localhost:3001/livereload");evtSource.onmessage = function(e) { if(e.data === "reload") { setTimeout(() => window.location.reload(), 100); } };</script>`
			if strings.Contains(appLayoutStr, "</body>") {
				appLayoutStr = strings.Replace(appLayoutStr, "</body>", script+"\n</body>", 1)
			} else {
				appLayoutStr += "\n" + script
			}
		}
	}

	componentsDir := filepath.Join(projectDir, "components")
	var componentsBuilder strings.Builder

	filepath.WalkDir(componentsDir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".spidey") {
			content, _ := os.ReadFile(path)
			name := strings.TrimSuffix(filepath.Base(path), ".spidey")

			// Wrap HTML in a Go template define block
			componentsBuilder.WriteString(fmt.Sprintf("\n{{define \"%s\"}}\n%s\n{{end}}\n", name, string(content)))
		}
		return nil
	})

	componentsStr := componentsBuilder.String()

	// Transpile .spidey files
	err = filepath.WalkDir(pagesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".spidey") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			fileName := filepath.Base(path)
			componentName := strings.TrimSuffix(fileName, ".spidey")

			goCode, err := parser.TranspileToGo(componentName, string(content), appLayoutStr, componentsStr)
			if err != nil {
				return err
			}

			genPath := filepath.Join(genDir, componentName+"_spidey.go")
			os.WriteFile(genPath, []byte(goCode), 0644)
		}
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("transpilation failed: %v", err)
	}

	// 5. esbuild frontend logic...
	componentEntry := filepath.Join(projectDir, "components", "index.js")
	if _, err := os.Stat(componentEntry); err == nil {
		fmt.Println("Spidey: Bundling frontend components...")
		api.Build(api.BuildOptions{
			EntryPoints:       []string{componentEntry},
			Outfile:           filepath.Join(projectDir, "public", "assets", "bundle.js"),
			Bundle:            true,
			MinifyWhitespace:  true,
			MinifyIdentifiers: true,
			MinifySyntax:      true,
			Write:             true,
		})
	}

	return nil
}

func CompileBinary(projectDir string) error {
	cmd := exec.Command("go", "build", "-o", "bin/server.exe", "./api/main.go")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
