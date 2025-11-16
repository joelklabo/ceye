package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joelklabo/ceye/internal/core"
)

func TestServerStaticFiles(t *testing.T) {
	store := core.NewStore()
	srv := New(store, []string{"demo"}, 8080)

	tests := []struct {
		name         string
		path         string
		wantStatus   int
		wantContains string
	}{
		{
			name:         "index page loads",
			path:         "/",
			wantStatus:   http.StatusOK,
			wantContains: "<title>ceye</title>",
		},
		{
			name:         "index contains ceye header",
			path:         "/",
			wantStatus:   http.StatusOK,
			wantContains: "<h1>🔍 ceye</h1>",
		},
		{
			name:         "css loads",
			path:         "/style.css",
			wantStatus:   http.StatusOK,
			wantContains: "background: #0a0e27",
		},
		{
			name:         "javascript loads",
			path:         "/app.js",
			wantStatus:   http.StatusOK,
			wantContains: "function connect()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			mux := http.NewServeMux()
			mux.HandleFunc("/ws", srv.handleWebSocket)
			webFS, err := srv.getWebFS()
			if err != nil {
				t.Fatal(err)
			}
			mux.Handle("/", http.FileServer(http.FS(webFS)))

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			body := rec.Body.String()
			if !strings.Contains(body, tt.wantContains) {
				t.Errorf("body does not contain %q", tt.wantContains)
			}
		})
	}
}

func TestWebSocketConnection(t *testing.T) {
	store := core.NewStore()
	providerNames := []string{"demo"}
	srv := New(store, providerNames, 8080)

	server := httptest.NewServer(http.HandlerFunc(srv.handleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusSwitchingProtocols)
	}

	// Should receive initial snapshot
	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read message failed: %v", err)
	}

	if msg.Type != "runs_update" {
		t.Errorf("msg.Type = %q, want %q", msg.Type, "runs_update")
	}

	if len(msg.Providers) != 1 || msg.Providers[0] != "demo" {
		t.Errorf("msg.Providers = %v, want [demo]", msg.Providers)
	}
}

func TestBroadcastUpdate(t *testing.T) {
	store := core.NewStore()
	srv := New(store, []string{"demo"}, 8080)

	// Add some test data
	store.Merge(core.RunEvent{
		Provider: "demo",
		Runs: []core.Run{
			{
				ID:           "run1",
				Provider:     "demo",
				Repo:         "test/repo",
				WorkflowName: "Test Workflow",
				Status:       core.RunStatusInProgress,
				Branch:       "main",
			},
		},
		Timestamp: time.Now(),
	})

	server := httptest.NewServer(http.HandlerFunc(srv.handleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect first client
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client 1 dial failed: %v", err)
	}
	defer conn1.Close()

	// Read initial snapshot
	var initialMsg Message
	if err := conn1.ReadJSON(&initialMsg); err != nil {
		t.Fatalf("client 1 initial read failed: %v", err)
	}

	// Connect second client
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client 2 dial failed: %v", err)
	}
	defer conn2.Close()

	// Read initial snapshot for second client
	if err := conn2.ReadJSON(&initialMsg); err != nil {
		t.Fatalf("client 2 initial read failed: %v", err)
	}

	// Broadcast update
	srv.BroadcastUpdate()

	// Both clients should receive the broadcast
	conn1.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg1 Message
	if err := conn1.ReadJSON(&msg1); err != nil {
		t.Errorf("client 1 broadcast read failed: %v", err)
	}

	conn2.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg2 Message
	if err := conn2.ReadJSON(&msg2); err != nil {
		t.Errorf("client 2 broadcast read failed: %v", err)
	}

	if len(msg1.Runs) != 1 {
		t.Errorf("client 1 runs = %d, want 1", len(msg1.Runs))
	}

	if len(msg2.Runs) != 1 {
		t.Errorf("client 2 runs = %d, want 1", len(msg2.Runs))
	}
}

func TestMessageFormat(t *testing.T) {
	store := core.NewStore()
	store.Merge(core.RunEvent{
		Provider: "demo",
		Runs: []core.Run{
			{
				ID:           "run1",
				Provider:     "demo",
				Repo:         "test/repo",
				WorkflowName: "Build",
				Status:       core.RunStatusInProgress,
				Branch:       "main",
				Duration:     5 * time.Minute,
			},
			{
				ID:           "run2",
				Provider:     "demo",
				Repo:         "test/repo2",
				WorkflowName: "Test",
				Status:       core.RunStatusCompleted,
				Conclusion:   "success",
				Branch:       "develop",
			},
		},
		Timestamp: time.Now(),
	})

	srv := New(store, []string{"demo"}, 8080)
	srv.UpdateStatus(
		map[string]string{"demo": ""},
		map[string]core.ProviderHealth{
			"demo": {
				LastSuccess: time.Now(),
				ErrorCount:  0,
			},
		},
	)

	server := httptest.NewServer(http.HandlerFunc(srv.handleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	// Verify message structure
	if msg.Type != "runs_update" {
		t.Errorf("Type = %q, want runs_update", msg.Type)
	}

	if len(msg.Runs) != 2 {
		t.Errorf("len(Runs) = %d, want 2", len(msg.Runs))
	}

	if len(msg.Providers) != 1 {
		t.Errorf("len(Providers) = %d, want 1", len(msg.Providers))
	}

	if msg.Status == nil {
		t.Error("Status is nil")
	}

	if msg.Health == nil {
		t.Error("Health is nil")
	}

	if msg.Totals == nil {
		t.Error("Totals is nil")
	}

	// Check totals calculation
	if msg.Totals["running"] != 1 {
		t.Errorf("Totals[running] = %d, want 1", msg.Totals["running"])
	}

	if msg.Totals["success"] != 1 {
		t.Errorf("Totals[success] = %d, want 1", msg.Totals["success"])
	}
}

func TestUpdateStatus(t *testing.T) {
	store := core.NewStore()
	srv := New(store, []string{"demo"}, 8080)

	providerStatus := map[string]string{
		"demo": "test error",
	}

	providerHealth := map[string]core.ProviderHealth{
		"demo": {
			LastError:  time.Now(),
			ErrorCount: 5,
		},
	}

	srv.UpdateStatus(providerStatus, providerHealth)

	// Verify status was updated by getting a snapshot
	server := httptest.NewServer(http.HandlerFunc(srv.handleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if msg.Status["demo"] != "test error" {
		t.Errorf("Status[demo] = %q, want %q", msg.Status["demo"], "test error")
	}

	if msg.Health["demo"].ErrorCount != 5 {
		t.Errorf("Health[demo].ErrorCount = %d, want 5", msg.Health["demo"].ErrorCount)
	}
}

func TestTotalsCalculation(t *testing.T) {
	tests := []struct {
		name  string
		runs  []core.Run
		wants map[string]int
	}{
		{
			name: "mixed statuses",
			runs: []core.Run{
				{Status: core.RunStatusInProgress},
				{Status: core.RunStatusInProgress},
				{Status: core.RunStatusQueued},
				{Status: core.RunStatusCompleted, Conclusion: "success"},
				{Status: core.RunStatusCompleted, Conclusion: "failure"},
				{Status: core.RunStatusFailed},
			},
			wants: map[string]int{
				"running": 2,
				"queued":  1,
				"success": 1,
				"failed":  2,
			},
		},
		{
			name: "all running",
			runs: []core.Run{
				{Status: core.RunStatusInProgress},
				{Status: core.RunStatusInProgress},
				{Status: core.RunStatusInProgress},
			},
			wants: map[string]int{
				"running": 3,
				"queued":  0,
				"success": 0,
				"failed":  0,
			},
		},
		{
			name:  "empty",
			runs:  []core.Run{},
			wants: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := core.NewStore()
			if len(tt.runs) > 0 {
				for i, run := range tt.runs {
					run.ID = string(rune(i))
					run.Provider = "test"
					tt.runs[i] = run
				}
				store.Merge(core.RunEvent{
					Provider:  "test",
					Runs:      tt.runs,
					Timestamp: time.Now(),
				})
			}

			srv := New(store, []string{"test"}, 8080)

			server := httptest.NewServer(http.HandlerFunc(srv.handleWebSocket))
			defer server.Close()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("dial failed: %v", err)
			}
			defer conn.Close()

			var msg Message
			if err := conn.ReadJSON(&msg); err != nil {
				t.Fatalf("read failed: %v", err)
			}

			for key, want := range tt.wants {
				if got := msg.Totals[key]; got != want {
					t.Errorf("Totals[%s] = %d, want %d", key, got, want)
				}
			}
		})
	}
}

func TestJSONSerialization(t *testing.T) {
	msg := Message{
		Type:      "runs_update",
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Runs: []core.Run{
			{
				ID:           "run1",
				Provider:     "demo",
				Repo:         "test/repo",
				WorkflowName: "Build",
				Status:       core.RunStatusInProgress,
			},
		},
		Providers: []string{"demo"},
		Status:    map[string]string{"demo": ""},
		Health: map[string]core.ProviderHealth{
			"demo": {LastSuccess: time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC)},
		},
		Totals: map[string]int{"running": 1},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Type != msg.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, msg.Type)
	}

	if len(decoded.Runs) != len(msg.Runs) {
		t.Errorf("len(Runs) = %d, want %d", len(decoded.Runs), len(msg.Runs))
	}

	if decoded.Totals["running"] != msg.Totals["running"] {
		t.Errorf("Totals[running] = %d, want %d", decoded.Totals["running"], msg.Totals["running"])
	}
}
