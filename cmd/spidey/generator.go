package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/saviru/spidey/internal/cli"
)

func createFileIfNotExists(path string, content []byte) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.WriteFile(path, content, 0644)
	} else {
		cli.Warn("init", fmt.Sprintf("Skipped %s (already exists)", path))
	}
}

func initProject(projectName string) {
	if projectName != "" {
		cli.Info("init", fmt.Sprintf("Initializing Go module: %s", projectName))
		cmd := exec.Command("go", "mod", "init", projectName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			cli.Error("init", fmt.Sprintf("Failed to initialize Go module: %v", err))
		}
	}

	cli.Info("init", "Fetching dependencies...")
	getCmd := exec.Command("go", "get", "github.com/goccy/go-json", "github.com/go-playground/validator/v10", "github.com/saviru/spidey@v0.26.904", "github.com/traefik/yaegi/interp", "github.com/traefik/yaegi/stdlib", "github.com/redis/go-redis/v9", "github.com/gorilla/websocket")
	getCmd.Stdout = os.Stdout
	getCmd.Stderr = os.Stderr
	if err := getCmd.Run(); err != nil {
		cli.Error("init", fmt.Sprintf("Failed to fetch dependencies: %v", err))
	}

	// Create workspace folders
	cli.Info("init", "Creating workspace folders...")
	dirs := []string{"api", "internal/spidey", "internal/spidey/pages", "internal/spidey/routes", "pages", "components", "public"}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	// internal/spidey/spidey_base.go
	baseCodeBytes, err := starterTemplates.ReadFile("templates/base.go")
	if err == nil {
		baseCode := string(baseCodeBytes)
		baseCode = strings.Replace(baseCode, "//go:build ignore", "", 1)
		baseCode = strings.Replace(baseCode, "package pages", "package spidey", 1)
		baseCode = strings.TrimSpace(baseCode) + "\n"

		createFileIfNotExists("internal/spidey/spidey_base.go", []byte(baseCode))
	} else {
		cli.Warn("init", "Failed to create base file.")
	}

	// .gitignore
	gitignore := "internal/spidey/\nbin/\n.env\n"
	createFileIfNotExists(".gitignore", []byte(gitignore))

	// spidey.config.json
	configJsonBytes, err := starterTemplates.ReadFile("templates/config.json")
	if err == nil {
		createFileIfNotExists("spidey.config.json", configJsonBytes)
	} else {
		cli.Warn("init", "Failed to create spidey.config.json file.")
	}

	// api/main.go
	apiMainCodeBytes, err := starterTemplates.ReadFile("templates/server.txt")
	if err == nil {
		apiMainCodeTpl := string(apiMainCodeBytes)
		apiMainCode := fmt.Sprintf(apiMainCodeTpl, projectName)
		createFileIfNotExists("api/main.go", []byte(apiMainCode))
	} else {
		cli.Warn("init", "Failed to create api/main.go file.")
	}

	defaultPageBytes, err := starterTemplates.ReadFile("templates/spidey/app.html")
	if err == nil {
		createFileIfNotExists("app.spidey", defaultPageBytes)
	} else {
		cli.Warn("init", "Failed to create app.spidey template.")
	}

	indexPageBytes, err := starterTemplates.ReadFile("templates/spidey/index.html")
	if err == nil {
		createFileIfNotExists("pages/index.spidey", indexPageBytes)
	} else {
		cli.Warn("init", "Failed to create index.spidey template.")
	}

	// Create a placeholder routes file to prevent "undefined: pages.RegisterRoutes" errors
	importPath := "testapp"
	if projectName != "" {
		importPath = projectName
	}

	routesByte, err := starterTemplates.ReadFile("templates/routes.txt")
	if err == nil {
		routesCode := fmt.Sprintf(string(routesByte), importPath)
		createFileIfNotExists("internal/spidey/routes/routes.go", []byte(routesCode))
	} else {
		cli.Warn("init", "Failed to inject default route. Please run 'spidey dev' to fix it.")
	}

	cli.Ready("Workspace initialized successfully.")
	fmt.Println("Run 'spidey dev' to start the development server.")
}
