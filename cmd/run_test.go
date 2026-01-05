package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestInjectHotReload tests the hot reload injection middleware
func TestInjectHotReload(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		acceptHeader   string
		urlPath        string
		expectInjected bool
	}{
		{
			name:           "HTML with body tag",
			responseBody:   "<html><body><h1>Test</h1></body></html>",
			acceptHeader:   "text/html",
			urlPath:        "/",
			expectInjected: true,
		},
		{
			name:           "HTML file extension",
			responseBody:   "<html><body>Content</body></html>",
			acceptHeader:   "text/plain",
			urlPath:        "/index.html",
			expectInjected: true,
		},
		{
			name:           "Non-HTML content",
			responseBody:   "console.log('test');",
			acceptHeader:   "application/javascript",
			urlPath:        "/script.js",
			expectInjected: false,
		},
		{
			name:           "HTML without body tag",
			responseBody:   "<html><head><title>Test</title></head></html>",
			acceptHeader:   "text/html",
			urlPath:        "/",
			expectInjected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tt.responseBody))
			})

			handler := injectHotReload(mockHandler)
			req := httptest.NewRequest("GET", tt.urlPath, nil)
			req.Header.Set("Accept", tt.acceptHeader)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)
			body := rec.Body.String()

			hasScript := strings.Contains(body, "new WebSocket")
			if tt.expectInjected && !hasScript {
				t.Error("Expected script to be injected")
			}
			if !tt.expectInjected && hasScript {
				t.Error("Script should not be injected")
			}

			// Verify injection happens before </body>
			if tt.expectInjected && strings.Contains(tt.responseBody, "</body>") {
				scriptIdx := strings.Index(body, "new WebSocket")
				bodyIdx := strings.Index(body, "</body>")
				if scriptIdx >= bodyIdx {
					t.Error("Script should be injected before </body> tag")
				}
			}
		})
	}
}

// TestWebSocketConnection tests WebSocket handling
func TestWebSocketConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(handleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Verify connection is registered
	time.Sleep(50 * time.Millisecond)
	clientsMu.Lock()
	count := len(clients)
	clientsMu.Unlock()

	if count != 1 {
		t.Errorf("Expected 1 client, got %d", count)
	}
}

// TestNotifyReload tests reload notification to clients
func TestNotifyReload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(handleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)
	NotifyReload()

	// Read the reload message
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	if string(msg) != "reload" {
		t.Errorf("Expected 'reload' message, got %q", msg)
	}
}

// TestRunCommandFlags tests flag configuration
func TestRunCommandFlags(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		checkPort int
		checkOpen bool
		checkDev  bool
	}{
		{"port flag", []string{"--port", "9000"}, 9000, false, false},
		{"open flag", []string{"--open"}, 8080, true, false},
		{"dev flag", []string{"--dev"}, 8080, false, true},
		{"short flags", []string{"-p", "7000", "-o", "-d"}, 7000, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port = 8080
			open = false
			dev = false

			runCmd.ParseFlags(tt.args)

			if port != tt.checkPort {
				t.Errorf("Expected port %d, got %d", tt.checkPort, port)
			}
			if open != tt.checkOpen {
				t.Errorf("Expected open %v, got %v", tt.checkOpen, open)
			}
			if dev != tt.checkDev {
				t.Errorf("Expected dev %v, got %v", tt.checkDev, dev)
			}
		})
	}
}

// TestReloadScriptContent verifies the reload script
func TestReloadScriptContent(t *testing.T) {
	required := []string{"WebSocket", "window.location.reload", "/ws"}
	for _, str := range required {
		if !strings.Contains(reloadScript, str) {
			t.Errorf("Reload script should contain %q", str)
		}
	}
}

// TestWebSocketUpgrader tests CORS configuration
func TestWebSocketUpgrader(t *testing.T) {
	req := &http.Request{Header: http.Header{}}
	req.Header.Set("Origin", "http://example.com")

	if !upgrader.CheckOrigin(req) {
		t.Error("CheckOrigin should allow all origins")
	}
}
