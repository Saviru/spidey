package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func createFileIfNotExists(path string, content []byte) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.WriteFile(path, content, 0644)
	} else {
		fmt.Printf("Skipped %s (already exists)\n", path)
	}
}

func initProject(projectName string) {
	if projectName != "" {
		fmt.Printf("Initializing Go module: %s\n", projectName)
		cmd := exec.Command("go", "mod", "init", projectName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to initialize Go module: %v\n", err)
		}
	}

	// Create workspace folders
	dirs := []string{"api", "pages", "components", "public"}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	os.MkdirAll("lib/pages", 0755)
	os.MkdirAll("lib/router", 0755)

	// lib/router/router.go
	routerCodeBytes, err := starterTemplates.ReadFile("templates/router.go")
	if err == nil {
		routerCode := string(routerCodeBytes)
		routerCode = strings.Replace(routerCode, "//go:build ignore", "", 1)

		routerCode = strings.TrimSpace(routerCode) + "\n"

		createFileIfNotExists("lib/router/router.go", []byte(routerCode))
	} else {
		fmt.Println("Warning: Failed to inject router.")
	}

	// lib/pages/base.go
	baseCodeBytes, err := starterTemplates.ReadFile("templates/base.go")
	if err == nil {
		baseCode := string(baseCodeBytes)
		baseCode = strings.Replace(baseCode, "//go:build ignore", "", 1)
		baseCode = strings.TrimSpace(baseCode) + "\n"

		createFileIfNotExists("lib/pages/base.go", []byte(baseCode))
	} else {
		fmt.Println("Warning: Failed to create base file.")
	}

	// .gitignore
	gitignore := "lib/\nbin/\n.env\n"
	createFileIfNotExists(".gitignore", []byte(gitignore))

	// spidey.config.json
	configJsonBytes, err := starterTemplates.ReadFile("templates/config.json")
	if err == nil {
		createFileIfNotExists("spidey.config.json", configJsonBytes)
	} else {
		fmt.Println("Warning: Failed to create spidey.config.json file.")
	}

	// api/main.go
	apiMainCodeBytes, err := starterTemplates.ReadFile("templates/server.txt")
	if err == nil {
		apiMainCodeTpl := string(apiMainCodeBytes)
		apiMainCode := fmt.Sprintf(apiMainCodeTpl, projectName, projectName)
		createFileIfNotExists("api/main.go", []byte(apiMainCode))
	} else {
		fmt.Println("Warning: Failed to create api/main.go file.")
	}

	defaultPageBytes, err := starterTemplates.ReadFile("templates/spidey/app.html")
	if err == nil {
		createFileIfNotExists("app.spidey", defaultPageBytes)
	} else {
		fmt.Println("Warning: Failed to create app.spidey template.")
	}

	indexPageBytes, err := starterTemplates.ReadFile("templates/spidey/index.html")
	if err == nil {
		createFileIfNotExists("pages/index.spidey", indexPageBytes)
	} else {
		fmt.Println("Warning: Failed to create index.spidey template.")
	}

	// Create a placeholder routes file to prevent "undefined: pages.RegisterRoutes" errors
	importPath := "testapp"
	if projectName != "" {
		importPath = projectName
	}

	routesByte, err := starterTemplates.ReadFile("templates/routes.txt")
	if err == nil {
		routesCode := fmt.Sprintf(string(routesByte), importPath)
		createFileIfNotExists("lib/pages/routes.go", []byte(routesCode))
	} else {
		fmt.Println("Warning: Failed to inject default route. Please run 'spidey dev' to fix it.")
	}

	fmt.Println("Spidey workspace created successfully")
}
