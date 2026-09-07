package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/saviru/spidey/internal/bundler"
	"github.com/saviru/spidey/internal/cli"
	"github.com/saviru/spidey/internal/config"
	"github.com/saviru/spidey/internal/dev"
)

//go:embed templates/*
var starterTemplates embed.FS

func main() {
	if len(os.Args) < 2 {
		cli.Error("init", "No command provided\n\nUsage: spidey [init|dev|build|version|update]")
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
			cli.Ready(fmt.Sprintf("Spidey CLI %s", info.Main.Version))
		} else {
			cli.Ready("Spidey CLI (development build)")
		}
	case "init", "hatch":
		var projectName string
		if len(os.Args) > 2 {
			projectName = os.Args[2]
		}
		initProject(projectName)
	case "dev", "weave":
		cli.Info("dev", "Starting development server...")
		cfg := config.LoadConfig(currentDir)
		// Pass starterTemplates to the watcher
		dev.StartWatcher(currentDir, starterTemplates, cfg)
	case "build", "wrap":
		cli.Info("build", "Transpiling pages...")
		cfg := config.LoadConfig(currentDir)
		// Pass starterTemplates to the build engine
		if err := bundler.ProcessPages(currentDir, starterTemplates, "", cfg); err != nil {
			cli.Error("build", fmt.Sprintf("Engine Error: %v", err))
			os.Exit(1)
		}

		cli.Info("build", "Compiling final binary...")
		if err := bundler.CompileBinary(currentDir, cfg); err != nil {
			cli.Error("build", fmt.Sprintf("Compilation Error: %v", err))
			os.Exit(1)
		}

		outPath := filepath.Join(currentDir, cfg.Directories.OutputDir+cli.DetectOS())
		displayPath := "./" + cfg.Directories.OutputDir + cli.DetectOS()

		sizeStr := ""
		if fi, err := os.Stat(outPath); err == nil {
			sizeStr = fmt.Sprintf(" (%s)", cli.FormatFileSize(fi.Size()))
		}

		cli.Success("build", fmt.Sprintf("Compilation successful! Output: %s%s", displayPath, sizeStr))
	case "export", "shed":
		cli.Info("build", "Transpiling pages for static export...")
		cfg := config.LoadConfig(currentDir)
		if err := bundler.ProcessPages(currentDir, starterTemplates, "", cfg); err != nil {
			cli.Error("build", fmt.Sprintf("Engine Error: %v", err))
			os.Exit(1)
		}

		cli.Info("build", "Compiling temporary SSG binary...")
		if err := bundler.CompileBinary(currentDir, cfg); err != nil {
			cli.Error("build", fmt.Sprintf("Compilation Error: %v", err))
			os.Exit(1)
		}

		// Run the generated binary with --export
		exportCmd := exec.Command(fmt.Sprintf("./%s", cfg.Directories.OutputDir), "--export")
		exportCmd.Dir = currentDir
		exportCmd.Stdout = os.Stdout
		exportCmd.Stderr = os.Stderr
		if err := exportCmd.Run(); err != nil {
			cli.Error("build", fmt.Sprintf("Export Error: %v", err))
			os.Exit(1)
		}
	case "update":
		cli.Info("update", "Downloading the latest version...")
		cmd := exec.Command("go", "install", "github.com/saviru/spidey/cmd/spidey@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			cli.Error("update", fmt.Sprintf("Update Error: %v", err))
			os.Exit(1)
		}
		cli.Ready("Update successful!")
	default:
		cli.Error("init", fmt.Sprintf("Unknown command: %s\n\n  Usage: spidey [init|dev|build|version|update]"))
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
		cli.Warn("update", fmt.Sprintf("A new version of Spidey is available! (%s -> %s)", currentVersion, release.TagName))
		cli.Warn("update", "Run 'spidey update' to upgrade instantly.")
	}
}
