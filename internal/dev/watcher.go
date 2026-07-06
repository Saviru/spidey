package dev

import (
	"embed"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"spidey/internal/bundler"
	"spidey/internal/config"

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
			fmt.Fprintf(w, "data: reload\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	})

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Println("Engine Error: Could not start livereload server.", err)
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

type filterWriter struct {
	w io.Writer
}

func (fw filterWriter) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "Spidey Server running on") {
		return len(p), nil
	}
	return fw.w.Write(p)
}

// Find difined port no
func getPort(projectDir string, cfg *config.Config) string {
	mainPath := filepath.Join(projectDir, "api", "main.go")
	content, err := os.ReadFile(mainPath)
	if err == nil {
		re := regexp.MustCompile(`app\.Listen\("(.*?)"\)`)
		matches := re.FindStringSubmatch(string(content))
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return fmt.Sprintf("%d", cfg.Port)
}

// Accept templates as the second argument
func StartWatcher(projectDir string, templates embed.FS, cfg *config.Config) {
	liveReloadPort := startLiveReloadServer()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	fmt.Println("Spidey: Running initial sync...")
	if err := bundler.ProcessPages(projectDir, templates, liveReloadPort, cfg); err != nil {
		fmt.Println("Sync error:", err)
	}

	pagesDir := filepath.Join(projectDir, "pages")
	componentsDir := filepath.Join(projectDir, "components")
	appFile := filepath.Join(projectDir, "app.spidey")

	err = watcher.Add(pagesDir)
	if err != nil {
		log.Fatal("Engine Error: Could not watch pages folder.", err)
	}

	watcher.Add(componentsDir)
	watcher.Add(appFile)

	port := getPort(projectDir, cfg)

	fmt.Printf("Spidey is running on %s\n", "http://localhost:"+port)

	restartServer(projectDir, cfg)
	openBrowser("http://localhost:" + port)
	fmt.Println("Spidey: monitoring for changes.")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// ignore metadata changes (like chmod)
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				fmt.Printf("File changed: %s | Syncing...\n", filepath.Base(event.Name))
				if err := bundler.ProcessPages(projectDir, templates, liveReloadPort, cfg); err != nil {
					fmt.Println("Sync error:", err)
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
			log.Println("Watcher error:", err)
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
		fmt.Printf("Compilation Error: %v\n", err)
		return
	}

	serverPath := filepath.Join(projectDir, cfg.Directories.OutputDir)
	if runtime.GOOS == "windows" {
		serverPath += ".exe"
	}

	serverCmd = exec.Command(serverPath)
	serverCmd.Dir = projectDir
	serverCmd.Stdout = filterWriter{os.Stdout}
	serverCmd.Stderr = filterWriter{os.Stderr}

	if err := serverCmd.Start(); err != nil {
		fmt.Println("Engine Error: Failed to start server, ", err)
	}
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("Engine Error: Failed to open browser: %v\n", err)
	}
}
