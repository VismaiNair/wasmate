package hotreload

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type WatcherOption func(*Watcher)

func WithDebounce(d time.Duration) WatcherOption {
	return func(w *Watcher) {
		w.debounceTime = d
	}
}

type Watcher struct {
	watcher      *fsnotify.Watcher
	paths        []string
	onChange     func()
	debounceTime time.Duration
}

func NewWatcher(paths []string, onChange func(), opts ...WatcherOption) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	w := &Watcher{
		watcher:      watcher,
		paths:        paths,
		onChange:     onChange,
		debounceTime: 500 * time.Millisecond,
	}

	for _, opt := range opts {
		opt(w)
	}

	for _, path := range paths {
		if err := w.WatchRecursive(path); err != nil {
			log.Printf("Warning: Failed to watch path %s: %v", path, err)
		}
	}

	return w, nil
}

// Start begins watching with a trailing-edge debounce timer
func (w *Watcher) Start() {
	go func() {
		var timer *time.Timer

		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}

				// 1. Handle new directories immediately
				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						log.Printf("New directory: %s. Adding to watcher...", filepath.Base(event.Name))
						w.WatchRecursive(event.Name)
						continue
					}
				}

				// 2. Filter for relevant Go file changes
				// Added: Remove and Rename to catch deletions or editor 'atomic saves'
				if isRelevantFileOp(event) {
					// 3. Trailing Edge Debounce
					// If a new event comes in, stop the existing timer and start a new one.
					// This ensures we only trigger ONCE after the user stops saving files.
					if timer != nil {
						timer.Stop()
					}

					timer = time.AfterFunc(w.debounceTime, func() {
						log.Printf("Change detected in %s, rebuilding...", filepath.Base(event.Name))
						w.onChange()
					})
				}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()
}

func isRelevantFileOp(event fsnotify.Event) bool {
	// Include Create, Write, Remove, and Rename
	const relevantOps = fsnotify.Write | fsnotify.Create | fsnotify.Remove | fsnotify.Rename

	if event.Op&relevantOps == 0 {
		return false
	}

	// Ignore temporary files (common in Vim/Emacs/GoLand)
	base := filepath.Base(event.Name)
	if strings.HasPrefix(base, ".") || strings.HasSuffix(base, "~") {
		return false
	}

	return strings.HasSuffix(event.Name, ".go")
}

func (w *Watcher) Close() error {
	return w.watcher.Close()
}

func (w *Watcher) WatchRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if info.IsDir() {
			name := filepath.Base(path)
			// Ignore common high-volume or hidden directories
			if (strings.HasPrefix(name, ".") && len(name) > 1) || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}

			if err := w.watcher.Add(path); err != nil {
				// Ignore 'already exists' errors
				if !strings.Contains(err.Error(), "already exists") {
					return fmt.Errorf("failed to watch %s: %w", path, err)
				}
			}
		}
		return nil
	})
}
