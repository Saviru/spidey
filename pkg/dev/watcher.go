package dev

import (
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"spidey/pkg/bundler"

	"github.com/fsnotify/fsnotify"
)

var clients []chan struct{}
var clientsMu sync.Mutex

func startLiveReloadServer() {
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
	go http.ListenAndServe(":3001", nil)
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
func getPort(projectDir string) string {
	mainPath := filepath.Join(projectDir, "api", "main.go")
	content, err := os.ReadFile(mainPath)
	if err == nil {
		re := regexp.MustCompile(`app\.Run\(":(.*?)"\)`)
		matches := re.FindStringSubmatch(string(content))
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return "3000"
}

// Accept templates as the second argument
func StartWatcher(projectDir string, templates embed.FS) {
	startLiveReloadServer()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	fmt.Println("Spidey: Running initial sync...")
	if err := bundler.ProcessPages(projectDir, templates, true); err != nil {
		fmt.Println("Sync error:", err)
	}

	pagesDir := filepath.Join(projectDir, "pages")
	componentsDir := filepath.Join(projectDir, "components")
	appFile := filepath.Join(projectDir, "app.spidey")

	err = watcher.Add(pagesDir)
	if err != nil {
		log.Fatal("Could not watch pages folder:", err)
	}

	watcher.Add(componentsDir)
	watcher.Add(appFile)

	fmt.Println("Spidey Watcher: monitoring for changes.")

	restartServer(projectDir)
	port := getPort(projectDir)
	openBrowser("http://localhost:" + port)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// ignore metadata changes (like chmod)
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				fmt.Printf("File changed: %s | Syncing...\n", filepath.Base(event.Name))
				if err := bundler.ProcessPages(projectDir, templates, true); err != nil {
					fmt.Println("Sync error:", err)
				}
				restartServer(projectDir)
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

func restartServer(projectDir string) {
	if serverCmd != nil && serverCmd.Process != nil {
		serverCmd.Process.Kill()
		serverCmd.Wait()
	}

	if err := bundler.CompileBinary(projectDir); err != nil {
		fmt.Printf("Compilation Error: %v\n", err)
		return
	}

	serverPath := filepath.Join(projectDir, "bin", "server")
	if runtime.GOOS == "windows" {
		serverPath += ".exe"
	}

	serverCmd = exec.Command(serverPath)
	serverCmd.Dir = projectDir
	serverCmd.Stdout = filterWriter{os.Stdout}
	serverCmd.Stderr = filterWriter{os.Stderr}

	if err := serverCmd.Start(); err != nil {
		fmt.Println("Failed to start server:", err)
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
		fmt.Printf("Failed to open browser: %v\n", err)
	}
}
