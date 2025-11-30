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

// WatcherOption is a function that configures the Watcher.
type WatcherOption func(*Watcher)

// WithDebounce sets the duration for debouncing rapid file change events.
func WithDebounce(d time.Duration) WatcherOption {
	return func(w *Watcher) {
		w.debounceTime = d
	}
}

// --- Watcher Struct (OOP State and Behavior) ---

// Watcher manages file watching and triggers a callback on relevant changes.
type Watcher struct {
	watcher      *fsnotify.Watcher // Underlying fsnotify watcher
	paths        []string
	onChange     func()
	debounceTime time.Duration
	lastChange   time.Time // State required for debouncing
}

// NewWatcher creates and initializes a new file watcher.
func NewWatcher(paths []string, onChange func(), opts ...WatcherOption) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	w := &Watcher{
		watcher:      watcher,
		paths:        paths,
		onChange:     onChange,
		debounceTime: 500 * time.Millisecond, // Default debounce
	}

	// Apply functional options to configure the Watcher object
	for _, opt := range opts {
		opt(w)
	}

	// Add all initial paths (recursively if they are directories)
	for _, path := range paths {
		if err := w.WatchRecursive(path); err != nil {
			log.Printf("Warning: Failed to watch initial path %s: %v", path, err)
		}
	}

	return w, nil
}

// --- Helper Functions (Pure Logic) ---

// isRelevantFileOp checks if the event is a Write or Create operation on a .go file.
func isRelevantFileOp(event fsnotify.Event) bool {
	// Only watch for Write and Create events
	if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
		return false
	}
	// Only trigger for .go files
	return strings.HasSuffix(event.Name, ".go")
}

// --- Watcher Methods (OOP Behavior) ---

// Start begins watching for file changes in a separate goroutine.
func (w *Watcher) Start() {
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}

				// 1. Dynamic Directory Watching: If a new directory is created, watch it recursively.
				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						log.Printf("📂 New directory created: %s. Watching recursively...", filepath.Base(event.Name))
						if err := w.WatchRecursive(event.Name); err != nil {
							log.Printf("Error watching new directory: %v", err)
						}
						continue // Skip onChange for directory creation itself
					}
				}

				// 2. Filter: Use the helper function to check for relevant file operations/types.
				if !isRelevantFileOp(event) {
					continue
				}

				// 3. Debounce: Check the state (`w.lastChange`) to prevent rapid reloads.
				now := time.Now()
				if now.Sub(w.lastChange) < w.debounceTime {
					continue
				}
				w.lastChange = now

				// 4. Trigger Action
				log.Printf("File changed: %s", filepath.Base(event.Name))
				w.onChange()

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()
}

// Close stops the underlying fsnotify watcher.
func (w *Watcher) Close() error {
	return w.watcher.Close()
}

// WatchRecursive adds the root path and all valid subdirectories to the watcher.
func (w *Watcher) WatchRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Ignore "file does not exist" errors, which can happen during rapid operations.
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if info.IsDir() {
			name := filepath.Base(path)
			// Skip hidden and common ignored directories
			if (strings.HasPrefix(name, ".") && len(name) > 1) || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}

			// Add the directory to the watcher
			if err := w.watcher.Add(path); err != nil && !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("failed to watch directory %s: %w", path, err)
			}
		}
		return nil
	})
}
