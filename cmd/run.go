package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/vismainair/wasmate/browser"
	"github.com/vismainair/wasmate/hotreload"
)

var port int
var open bool
var dev bool

// WebSocket configuration
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Thread-safe storage for active browser connections
var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

// Script to be injected into HTML files
const reloadScript = `
<script>
  (function() {
    // Wait for the page to fully load first
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', connectWebSocket);
    } else {
      connectWebSocket();
    }
    
    function connectWebSocket() {
      const socket = new WebSocket('ws://' + window.location.host + '/ws');
      socket.onmessage = function(msg) {
        if (msg.data === 'reload') {
          console.log('Wasmate: Change detected, reloading...');
          window.location.reload();
        }
      };
      socket.onopen = function() {
        console.log('Wasmate: Hot reload connected.');
      };
      socket.onclose = function() {
        console.log('Wasmate: Hot reload disconnected.');
      };
      socket.onerror = function(err) {
        console.error('Wasmate: WebSocket error:', err);
      };
    }
  })();
</script>
`

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "wasmate run serves all static files over a webserver.",
	RunE: func(cmd *cobra.Command, args []string) error {
		mux := http.NewServeMux()

		// 1. Setup the standard file server
		fs := http.FileServer(http.Dir("."))
		mux.Handle("/", fs)

		// 2. Setup the WebSocket endpoint
		mux.HandleFunc("/ws", handleWebSocket)

		// 3. Setup Middleware and Watcher if dev mode is enabled
		var handler http.Handler = mux
		if dev {
			handler = injectHotReload(mux)

			// Find the WASM file to watch and rebuild
			wasmFile := findWasmFile()
			if wasmFile == "" {
				fmt.Fprintf(os.Stderr, "⚠ Warning: No .wasm file found in current directory. Hot reload will not rebuild.\n")
			}

			watcher, err := hotreload.NewWatcher(
				[]string{"."},
				func() {
					fmt.Fprintf(os.Stdout, "Change detected, rebuilding...\n")

					// Rebuild the WASM file
					if wasmFile != "" {
						err := RunBuild(".", wasmFile, false)
						if err != nil {
							fmt.Fprintf(os.Stderr, "✗ Build failed: %v\n", err)
							return
						}
						fmt.Fprintf(os.Stdout, "✓ Rebuild complete\n")
					}

					// Small delay to ensure file is written
					time.Sleep(100 * time.Millisecond)

					// Notify browser to reload
					NotifyReload()
				},
			)
			if err != nil {
				return fmt.Errorf("failed to create watcher: %w", err)
			}
			defer watcher.Close()
			watcher.Start()
			fmt.Fprintf(os.Stdout, "Hot reload enabled. Watching for changes...\n")
		}

		if open {
			browser.Open(fmt.Sprintf("http://localhost:%d", port))
			fmt.Fprintf(os.Stdout, "Opening the file in your browser\n")
		}

		fmt.Fprintf(os.Stdout, "The WASM static webserver is starting on http://localhost:%d\n", port)
		fmt.Fprintf(os.Stdout, "Press Ctrl+C or Command+C to stop the server.\n")

		err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
		if err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}

		return nil
	},
}

// findWasmFile looks for a .wasm file in the current directory
func findWasmFile() string {
	files, err := filepath.Glob("*.wasm")
	if err != nil || len(files) == 0 {
		return ""
	}
	// Return the first .wasm file found
	return files[0]
}

// --- Middleware & Injection Logic ---

func injectHotReload(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only intercept HTML requests
		isHtml := strings.Contains(r.Header.Get("Accept"), "text/html") ||
			strings.HasSuffix(r.URL.Path, ".html") ||
			r.URL.Path == "/"

		if !isHtml {
			next.ServeHTTP(w, r)
			return
		}

		// Record the response to modify the HTML
		recorder := &responseRecorder{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(recorder, r)

		htmlContent := recorder.body.String()
		if strings.Contains(htmlContent, "</body>") {
			// Inject script before closing body tag
			modifiedHtml := strings.Replace(htmlContent, "</body>", reloadScript+"</body>", 1)

			// Write status code first
			w.WriteHeader(recorder.statusCode)

			// Update headers and write the modified content
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Content-Length", fmt.Sprint(len(modifiedHtml)))
			w.Write([]byte(modifiedHtml))
		} else {
			// If no body tag, write original content
			w.WriteHeader(recorder.statusCode)
			w.Write(recorder.body.Bytes())
		}
	})
}

type responseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

// --- WebSocket Handling ---

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func NotifyReload() {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte("reload"))
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().IntVarP(&port, "port", "p", 8080, "The port to run the web server on")
	runCmd.Flags().BoolVarP(&open, "open", "o", false, "Open the app on the default browser")
	runCmd.Flags().BoolVarP(&dev, "dev", "d", false, "Enable hot-reloading for development")
}
