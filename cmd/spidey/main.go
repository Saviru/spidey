package main

import (
	"embed"
	"fmt"
	"os"

	"spidey/pkg/bundler"
)

//go:embed templates/*
var starterTemplates embed.FS

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gox [init|dev|build]")
		os.Exit(1)
	}

	command := os.Args[1]
	currentDir, _ := os.Getwd()

	switch command {
	case "init":
		initProject()
	case "dev":
		fmt.Println("Starting dev server...")
		// Watch files and trigger bundler.ProcessPages() on change
	case "build":
		fmt.Println("Gox: Transpiling pages...")
		if err := bundler.ProcessPages(currentDir); err != nil {
			fmt.Printf("Engine Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Gox: Compiling final binary...")
		if err := bundler.CompileBinary(currentDir); err != nil {
			fmt.Printf("Compilation Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Build successful! Executable is in ./bin/server")
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

func initProject() {
	// Create standard workspace folders
	dirs := []string{"api", "pages", "components", "public"}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	// Create lib file to store generated code
	os.MkdirAll("lib/pages", 0755)
	os.MkdirAll("lib/router", 0755)

	// Inject the Router Code
	routerCode, err := starterTemplates.ReadFile("templates/router.go.txt")
	if err == nil {
		os.WriteFile("lib/router/router.go", routerCode, 0644)
	} else {
		fmt.Println("Warning: Failed to inject router code.")
	}

	// Inject the Dummy Page for first-time compilation
	dummyCode := "package pages\n\nfunc RenderIndex(data interface{}) (string, error) { return \"\", nil }"
	os.WriteFile("lib/pages/dummy.go", []byte(dummyCode), 0644)

	// Setup configs
	gitignore := "lib/\nbin/\n.env\n"
	os.WriteFile(".gitignore", []byte(gitignore), 0644)

	

	fmt.Println("Jet workspace created successfully")
}
