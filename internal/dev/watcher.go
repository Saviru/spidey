package dev

import (
	"embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/saviru/spidey/internal/bundler"
	"github.com/saviru/spidey/internal/cli"
	"github.com/saviru/spidey/internal/config"

	"github.com/fsnotify/fsnotify"
)

var clients []chan struct{}
var clientsMu sync.Mutex

func startLiveReloadServer() string {
	http.HandleFunc("/livereload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		reload := make(chan struct{}, 1)

		clientsMu.Lock()
		clients = append(clients, reload)
		clientsMu.Unlock()

		select {
		case <-r.Context().Done():
			return
		case <-reload:
			cli.Info("dev", "Reloading...")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	})

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		cli.Error("dev", "Could not start live reload server.")
		return "3001" // fallback
	}

	port := fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port)
	go http.Serve(listener, nil)

	return port
}

func triggerReload() {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for _, c := range clients {
		select {
		case c <- struct{}{}:
		default:
		}
	}
	// Clear clients list to reconnect
	clients = nil
}

// Find difined port no
func getPort(projectDir string, cfg *config.Config) string {
	mainPath := filepath.Join(projectDir, "api", "main.go")
	content, err := os.ReadFile(mainPath)
	if err == nil {
		re := regexp.MustCompile(`app\.Listen\("?([0-9]+)"?\)`)
		matches := re.FindStringSubmatch(string(content))
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return fmt.Sprintf("%d", cfg.Port)
}

// Accept templates as the second argument
func StartWatcher(projectDir string, templates embed.FS, cfg *config.Config) {
	startTime := time.Now()
	liveReloadPort := startLiveReloadServer()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		cli.Error("dev", "Could not start live reload server.")
		os.Exit(1)
	}
	defer watcher.Close()

	cli.Info("dev", "Running initial sync...")
	if err := bundler.ProcessPages(projectDir, templates, liveReloadPort, cfg); err != nil {
		cli.Error("dev", "Sync error.")
	}

	pagesDir := filepath.Join(projectDir, "pages")
	componentsDir := filepath.Join(projectDir, "components")
	appFile := filepath.Join(projectDir, "app.spidey")

	err = watcher.Add(pagesDir)
	if err != nil {
		cli.Error("dev", "Could not watch pages folder.")
		os.Exit(1)
	}

	watcher.Add(componentsDir)
	watcher.Add(appFile)

	port := getPort(projectDir, cfg)

	restartServer(projectDir, cfg)
	cli.DevBanner("Development build", port, time.Since(startTime))

	openBrowser("http://localhost:" + port)
	cli.Info("dev", "Monitoring for changes...")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// ignore metadata changes (like chmod)
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				cli.Info("dev", fmt.Sprintf("File changed: %s | Syncing...", filepath.Base(event.Name)))
				if err := bundler.ProcessPages(projectDir, templates, liveReloadPort, cfg); err != nil {
					cli.Error("dev", "Sync error.")
				}
				restartServer(projectDir, cfg)
				go func() {
					time.Sleep(250 * time.Millisecond)
					triggerReload()
				}()
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			cli.Error("dev", fmt.Sprintf("Watcher error: %v", err))
		}
	}
}

var serverCmd *exec.Cmd

func restartServer(projectDir string, cfg *config.Config) {
	if serverCmd != nil && serverCmd.Process != nil {
		serverCmd.Process.Kill()
		serverCmd.Wait()
	}

	if err := bundler.CompileBinary(projectDir, cfg); err != nil {
		cli.Error("dev", fmt.Sprintf("Compilation Error: %v", err))
		return
	}

	serverPath := filepath.Join(projectDir, cfg.Directories.OutputDir)
	serverPath += cli.DetectOS()

	serverCmd = exec.Command(serverPath)
	serverCmd.Dir = projectDir
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr

	if err := serverCmd.Start(); err != nil {
		cli.Error("dev", "Failed to start server.")
	}
}

func openBrowser(url string) {
	var err error
	switch cli.DetectOS() {
	case ".bin":
		err = exec.Command("xdg-open", url).Start()
	case ".exe":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case ".app":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		cli.Error("dev", "Failed to open browser.")
	}
}
