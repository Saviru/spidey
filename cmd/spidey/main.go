package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"time"

	"github.com/saviru/spidey/internal/bundler"
	"github.com/saviru/spidey/internal/config"
	"github.com/saviru/spidey/internal/dev"
)

//go:embed templates/*
var starterTemplates embed.FS

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: spidey [init|dev|build|version|update]")
		os.Exit(1)
	}

	command := os.Args[1]
	currentDir, _ := os.Getwd()

	if command != "version" && command != "update" {
		go checkLatestVersion()
	}

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
		cmd := exec.Command("go", "install", "github.com/saviru/spidey/cmd/spidey@latest")
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

func checkLatestVersion() {
	// Get current version
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return
	}
	currentVersion := info.Main.Version

	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/saviru/spidey/releases/latest")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	if release.TagName != "" && release.TagName != currentVersion {
		fmt.Printf("\nA new version of Spidey is available! (%s -> %s)\n", currentVersion, release.TagName)
		fmt.Println("Run 'spidey update' to upgrade instantly.")
	}
}
