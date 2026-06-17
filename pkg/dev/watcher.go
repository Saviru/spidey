package dev

import (
	"embed"
	"fmt"
	"log"
	"path/filepath"

	"spidey/pkg/bundler"

	"github.com/fsnotify/fsnotify"
)

// Accept templates as the second argument
func StartWatcher(projectDir string, templates embed.FS) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// 1. Pass templates here
	fmt.Println("Spidey: Running initial sync to clear IDE errors...")
	if err := bundler.ProcessPages(projectDir, templates); err != nil {
		fmt.Println("Sync error:", err)
	}

	pagesDir := filepath.Join(projectDir, "pages")

	err = watcher.Add(pagesDir)
	if err != nil {
		log.Fatal("Could not watch pages folder:", err)
	}

	fmt.Println("🕷️ Spidey Watcher running... monitoring /pages for changes.")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				fmt.Printf("File changed: %s | Syncing...\n", filepath.Base(event.Name))
				// 2. Pass templates here too
				bundler.ProcessPages(projectDir, templates)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}
