package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"

	"github.com/saviru/spidey/internal/bundler"
	"github.com/saviru/spidey/internal/config"
	"github.com/saviru/spidey/internal/dev"
)

//go:embed templates/*
var starterTemplates embed.FS

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: spidey [init|dev|build|version]")
		os.Exit(1)
	}

	command := os.Args[1]
	currentDir, _ := os.Getwd()

	switch command {
	case "version", "-v", "--version":
		info, ok := debug.ReadBuildInfo()
		if ok && info.Main.Version != "" {
			fmt.Printf("Spidey CLI %s\n", info.Main.Version)
		} else {
			fmt.Println("Spidey CLI (development build)")
		}
	case "init", "hatch":
		var projectName string
		if len(os.Args) > 2 {
			projectName = os.Args[2]
		}
		initProject(projectName)
	case "dev", "weave":
		fmt.Println("Starting Spidey development environment...")
		cfg := config.LoadConfig(currentDir)
		// Pass starterTemplates to the watcher
		dev.StartWatcher(currentDir, starterTemplates, cfg)
	case "build", "wrap":
		fmt.Println("Spidey: Transpiling pages...")
		cfg := config.LoadConfig(currentDir)
		// Pass starterTemplates to the build engine
		if err := bundler.ProcessPages(currentDir, starterTemplates, "", cfg); err != nil {
			fmt.Printf("Engine Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Spidey: Compiling final binary...")
		if err := bundler.CompileBinary(currentDir, cfg); err != nil {
			fmt.Printf("Compilation Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Build successful! Executable is in ./%s\n", cfg.Directories.OutputDir)
	case "export", "shed":
		fmt.Println("Spidey: Transpiling pages for static export...")
		cfg := config.LoadConfig(currentDir)
		if err := bundler.ProcessPages(currentDir, starterTemplates, "", cfg); err != nil {
			fmt.Printf("Engine Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Spidey: Compiling temporary SSG binary...")
		if err := bundler.CompileBinary(currentDir, cfg); err != nil {
			fmt.Printf("Compilation Error: %v\n", err)
			os.Exit(1)
		}

		// Run the generated binary with --export
		exportCmd := exec.Command(fmt.Sprintf("./%s", cfg.Directories.OutputDir), "--export")
		exportCmd.Dir = currentDir
		exportCmd.Stdout = os.Stdout
		exportCmd.Stderr = os.Stderr
		if err := exportCmd.Run(); err != nil {
			fmt.Printf("Export Error: %v\n", err)
			os.Exit(1)
		}
	case "update":
		fmt.Println("Downloading the latest version...")
		cmd := exec.Command("go", "install", "github.com/saviru/spidey@latest")
		cmd.Dir = currentDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Update Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Update successful!")
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}
