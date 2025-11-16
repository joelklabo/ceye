package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joelklabo/ceye/internal/core"
)

// TestFullWorkflow simulates a complete user workflow
func TestFullWorkflow(t *testing.T) {
	store := core.NewStore()
	srv := New(store, []string{"github", "azure"}, 8080)

	// Setup test server
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", srv.handleWebSocket)
	webFS, err := srv.getWebFS()
	if err != nil {
		t.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	server := httptest.NewServer(mux)
	defer server.Close()

	// Step 1: Load index page
	t.Run("load index page", func(t *testing.T) {
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	// Step 2: Connect WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer conn.Close()

	// Step 3: Receive initial snapshot
	t.Run("receive initial snapshot", func(t *testing.T) {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("read initial snapshot failed: %v", err)
		}

		if msg.Type != "runs_update" {
			t.Errorf("msg.Type = %q, want runs_update", msg.Type)
		}

		if len(msg.Providers) != 2 {
			t.Errorf("len(Providers) = %d, want 2", len(msg.Providers))
		}
	})

	// Step 4: Add runs and verify broadcast
	t.Run("add runs and receive broadcast", func(t *testing.T) {
		// Add some test runs
		store.Merge(core.RunEvent{
			Provider: "github",
			Runs: []core.Run{
				{
					ID:           "run1",
					Provider:     "github",
					Repo:         "test/repo",
					WorkflowName: "CI",
					Status:       core.RunStatusInProgress,
					Branch:       "main",
					CommitSHA:    "abc123",
					URL:          "https://github.com/test/repo/actions/runs/1",
				},
			},
			Timestamp: time.Now(),
		})

		// Trigger broadcast
		srv.BroadcastUpdate()

		// Should receive the update
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("read broadcast failed: %v", err)
		}

		if len(msg.Runs) != 1 {
			t.Errorf("len(Runs) = %d, want 1", len(msg.Runs))
		}

		if msg.Runs[0].WorkflowName != "CI" {
			t.Errorf("WorkflowName = %q, want CI", msg.Runs[0].WorkflowName)
		}

		if msg.Totals["running"] != 1 {
			t.Errorf("Totals[running] = %d, want 1", msg.Totals["running"])
		}
	})

	// Step 5: Update run status and verify
	t.Run("update run status", func(t *testing.T) {
		store.Merge(core.RunEvent{
			Provider: "github",
			Runs: []core.Run{
				{
					ID:           "run1",
					Provider:     "github",
					Repo:         "test/repo",
					WorkflowName: "CI",
					Status:       core.RunStatusCompleted,
					Conclusion:   "success",
					Branch:       "main",
					CommitSHA:    "abc123",
					URL:          "https://github.com/test/repo/actions/runs/1",
					Duration:     5 * time.Minute,
				},
			},
			Timestamp: time.Now(),
		})

		srv.BroadcastUpdate()

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("read update failed: %v", err)
		}

		if msg.Runs[0].Status != core.RunStatusCompleted {
			t.Errorf("Status = %q, want completed", msg.Runs[0].Status)
		}

		if msg.Totals["success"] != 1 {
			t.Errorf("Totals[success] = %d, want 1", msg.Totals["success"])
		}

		if msg.Totals["running"] != 0 {
			t.Errorf("Totals[running] = %d, want 0", msg.Totals["running"])
		}
	})

	// Step 6: Test provider health updates
	t.Run("provider health updates", func(t *testing.T) {
		srv.UpdateStatus(
			map[string]string{"github": "", "azure": "connection failed"},
			map[string]core.ProviderHealth{
				"github": {LastSuccess: time.Now(), ErrorCount: 0},
				"azure":  {LastError: time.Now(), ErrorCount: 5},
			},
		)

		srv.BroadcastUpdate()

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("read health update failed: %v", err)
		}

		if msg.Status["azure"] != "connection failed" {
			t.Errorf("Status[azure] = %q, want connection failed", msg.Status["azure"])
		}

		if msg.Health["azure"].ErrorCount != 5 {
			t.Errorf("Health[azure].ErrorCount = %d, want 5", msg.Health["azure"].ErrorCount)
		}
	})

	// Step 7: Add more runs from different provider
	t.Run("multiple providers", func(t *testing.T) {
		store.Merge(core.RunEvent{
			Provider: "azure",
			Runs: []core.Run{
				{
					ID:           "build-123",
					Provider:     "azure",
					Repo:         "azure/project",
					WorkflowName: "Build Pipeline",
					Status:       core.RunStatusQueued,
					Branch:       "develop",
				},
			},
			Timestamp: time.Now(),
		})

		srv.BroadcastUpdate()

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("read multi-provider update failed: %v", err)
		}

		if len(msg.Runs) != 2 {
			t.Errorf("len(Runs) = %d, want 2", len(msg.Runs))
		}

		if msg.Totals["queued"] != 1 {
			t.Errorf("Totals[queued] = %d, want 1", msg.Totals["queued"])
		}
	})
}

// TestMultipleClients tests concurrent WebSocket connections
func TestMultipleClients(t *testing.T) {
	store := core.NewStore()
	srv := New(store, []string{"test"}, 8080)

	server := httptest.NewServer(http.HandlerFunc(srv.handleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect 3 clients
	clients := make([]*websocket.Conn, 3)
	for i := range clients {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("client %d dial failed: %v", i, err)
		}
		defer conn.Close()
		clients[i] = conn

		// Read initial snapshot
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("client %d initial read failed: %v", i, err)
		}
	}

	// Add data
	store.Merge(core.RunEvent{
		Provider: "test",
		Runs: []core.Run{
			{
				ID:           "test-run",
				Provider:     "test",
				WorkflowName: "Test",
				Status:       core.RunStatusInProgress,
			},
		},
		Timestamp: time.Now(),
	})

	// Broadcast to all
	srv.BroadcastUpdate()

	// All clients should receive the update
	for i, conn := range clients {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			t.Errorf("client %d broadcast read failed: %v", i, err)
			continue
		}

		if len(msg.Runs) != 1 {
			t.Errorf("client %d: len(Runs) = %d, want 1", i, len(msg.Runs))
		}
	}
}

// TestStaticAssets verifies all static files are accessible
func TestStaticAssets(t *testing.T) {
	store := core.NewStore()
	srv := New(store, []string{}, 8080)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", srv.handleWebSocket)
	webFS, err := srv.getWebFS()
	if err != nil {
		t.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	server := httptest.NewServer(mux)
	defer server.Close()

	assets := []struct {
		path        string
		contentType string
		contains    string
	}{
		{"/", "text/html", "<title>ceye</title>"},
		{"/style.css", "text/css", ".container"},
		{"/app.js", "text/javascript", "function connect()"},
	}

	for _, asset := range assets {
		t.Run(asset.path, func(t *testing.T) {
			resp, err := http.Get(server.URL + asset.path)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
			}

			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, asset.contentType) {
				t.Errorf("Content-Type = %q, want %q", contentType, asset.contentType)
			}
		})
	}
}

// TestWebSocketReconnection tests handling of disconnections
func TestWebSocketReconnection(t *testing.T) {
	store := core.NewStore()
	srv := New(store, []string{"test"}, 8080)

	server := httptest.NewServer(http.HandlerFunc(srv.handleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// First connection
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Read initial snapshot
	var msg Message
	if err := conn1.ReadJSON(&msg); err != nil {
		t.Fatal(err)
	}

	// Close connection
	conn1.Close()

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Reconnect (simulating client reconnection)
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("reconnection failed: %v", err)
	}
	defer conn2.Close()

	// Should receive initial snapshot again
	if err := conn2.ReadJSON(&msg); err != nil {
		t.Fatalf("read after reconnection failed: %v", err)
	}

	if msg.Type != "runs_update" {
		t.Errorf("msg.Type = %q, want runs_update", msg.Type)
	}
}

// TestConcurrentUpdates tests thread safety
func TestConcurrentUpdates(t *testing.T) {
	store := core.NewStore()
	srv := New(store, []string{"test"}, 8080)

	server := httptest.NewServer(http.HandlerFunc(srv.handleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Read initial snapshot
	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatal(err)
	}

	// Fire multiple concurrent updates
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			srv.UpdateStatus(
				map[string]string{"test": "update"},
				map[string]core.ProviderHealth{
					"test": {ErrorCount: n},
				},
			)
			done <- true
		}(i)
	}

	// Wait for all updates to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Broadcast once after all updates
	srv.BroadcastUpdate()

	// Should be able to read the update without deadlock
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err := conn.ReadJSON(&msg); err != nil {
		t.Errorf("concurrent updates caused issues: %v", err)
	}

	if msg.Type != "runs_update" {
		t.Errorf("msg.Type = %q, want runs_update", msg.Type)
	}
}
