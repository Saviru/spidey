package bundler

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"spidey/pkg/parser"
)

func ProcessPages(projectDir string) error {
	pagesDir := filepath.Join(projectDir, "pages")
	genDir := filepath.Join(projectDir, "lib", "pages")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(genDir, 0755); err != nil {
		return fmt.Errorf("failed to create hidden build dir: %v", err)
	}

	// Delete previously transpiled files (*_spidey.go)
	entries, _ := os.ReadDir(genDir)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_spidey.go") {
			os.Remove(filepath.Join(genDir, entry.Name()))
		}
	}

	baseFile := filepath.Join(genDir, "spidey_base.go")
	os.WriteFile(baseFile, []byte("package pages\n"), 0644)

	// Transpile .spidey files
	err := filepath.WalkDir(pagesDir, func(path string, d fs.DirEntry, err error) error {
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

			goCode, err := parser.TranspileToGo(componentName, string(content))
			if err != nil {
				return err
			}

			genPath := filepath.Join(genDir, componentName+"_spidey.go")
			if err := os.WriteFile(genPath, []byte(goCode), 0644); err != nil {
				return err
			}
		}

		_ = os.Remove(filepath.Join(projectDir, "lib", "pages", "dummy.go"))

		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("transpilation failed: %v", err)
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
