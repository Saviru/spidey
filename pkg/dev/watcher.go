package dev

import (
	"fmt"
	"log"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func StartWatcher(projectDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					if strings.HasSuffix(event.Name, ".spidey") || strings.HasSuffix(event.Name, ".jsx") {
						fmt.Println("File changed:", event.Name)
						// Trigger pkg/bundler to rebuild
						// Send signal over WebSocket to reload browser
						fmt.Println("Rebuilding and reloading...")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

	// Watch specific directories
	watcher.Add(projectDir + "/pages")
	watcher.Add(projectDir + "/components")

	// Block forever (or start WebSocket server here)
	<-make(chan struct{})
}
