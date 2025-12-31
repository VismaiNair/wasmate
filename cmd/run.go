package cmd

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket" // Run: go get github.com/gorilla/websocket
	"github.com/spf13/cobra"
	"github.com/vismainair/wasmate/browser"
)

var port int
var open bool

// WebSocket configuration
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Thread-safe storage for active browser connections
var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "wasmate run serves all static files over a webserver.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {

		mux := http.NewServeMux()

		// 1. The standard file server
		fs := http.FileServer(http.Dir("."))
		mux.Handle("/", fs)

		// 2. The WebSocket endpoint for hot-reload signals
		mux.HandleFunc("/ws", handleWebSocket)

		if open {
			browser.Open(fmt.Sprintf("http://localhost:%d", port))
			fmt.Fprintf(os.Stdout, "Opening the file in your browser\n")
		}

		fmt.Fprintf(os.Stdout, "The WASM static webserver is starting on http://localhost:%d\n", port)
		fmt.Fprintf(os.Stdout, "Press Ctrl+C or Command+C to stop the server.\n")

		err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
		if err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}

		return nil
	},
}

// handleWebSocket upgrades the HTTP connection and tracks the client
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	// Keep connection alive until client disconnects
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

// NotifyReload sends a reload message to all connected browser tabs
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
}
