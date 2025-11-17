package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joelklabo/ceye/internal/core"
)

//go:embed web
var webAssets embed.FS

type Server struct {
	store          *core.Store
	port           int
	clients        map[*websocket.Conn]bool
	clientsMu      sync.RWMutex
	upgrader       websocket.Upgrader
	trendAnalyzer  interface{} // *storage.TrendAnalyzer (optional)
	
	providerStatus map[string]string
	providerHealth map[string]core.ProviderHealth
	providerNames  []string
	statusMu       sync.RWMutex
}

type Message struct {
	Type       string                         `json:"type"`
	Timestamp  time.Time                      `json:"timestamp"`
	Runs       []core.Run                     `json:"runs,omitempty"`
	Providers  []string                       `json:"providers,omitempty"`
	Status     map[string]string              `json:"status,omitempty"`
	Health     map[string]core.ProviderHealth `json:"health,omitempty"`
	Totals     map[string]int                 `json:"totals,omitempty"`
	AlertCount int                            `json:"alert_count,omitempty"`
}

func New(store *core.Store, providerNames []string, port int) *Server {
	return &Server{
		store:          store,
		port:           port,
		clients:        make(map[*websocket.Conn]bool),
		providerNames:  providerNames,
		providerStatus: make(map[string]string),
		providerHealth: make(map[string]core.ProviderHealth),
		trendAnalyzer:  nil,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// SetTrendAnalyzer sets the optional trend analyzer for analytics
func (s *Server) SetTrendAnalyzer(analyzer interface{}) {
	s.trendAnalyzer = analyzer
}

func (s *Server) getWebFS() (fs.FS, error) {
	return fs.Sub(webAssets, "web")
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	
	// WebSocket endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)
	
	// Analytics API endpoint
	mux.HandleFunc("/api/analytics/trends", s.handleAnalyticsTrends)
	
	// Alerts API endpoint
	mux.HandleFunc("/api/alerts/history", s.handleAlertsHistory)
	
	// Static files
	webFS, err := s.getWebFS()
	if err != nil {
		return fmt.Errorf("web assets: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webFS)))
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}
	
	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()
	
	log.Printf("Web server starting on http://localhost:%d", s.port)
	
	// Start listening
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Printf("Web server error: %v", err)
	}
	return err
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	
	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()
	
	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		conn.Close()
	}()
	
	// Send initial snapshot
	s.sendSnapshot(conn)
	
	// Keep connection alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (s *Server) sendSnapshot(conn *websocket.Conn) {
	runs := s.store.ListRuns("")
	
	s.statusMu.RLock()
	status := make(map[string]string)
	for k, v := range s.providerStatus {
		status[k] = v
	}
	health := make(map[string]core.ProviderHealth)
	for k, v := range s.providerHealth {
		health[k] = v
	}
	s.statusMu.RUnlock()
	
	totals := make(map[string]int)
	for _, run := range runs {
		switch run.Status {
		case core.RunStatusInProgress:
			totals["running"]++
		case core.RunStatusQueued:
			totals["queued"]++
		case core.RunStatusCompleted:
			if run.Conclusion == "success" {
				totals["success"]++
			} else {
				totals["failed"]++
			}
		case core.RunStatusFailed:
			totals["failed"]++
		}
	}
	
	msg := Message{
		Type:       "runs_update",
		Timestamp:  time.Now(),
		Runs:       runs,
		Providers:  s.providerNames,
		Status:     status,
		Health:     health,
		Totals:     totals,
		AlertCount: s.store.GetAlertCount(),
	}
	
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send snapshot: %v", err)
	}
}

func (s *Server) UpdateStatus(providerStatus map[string]string, providerHealth map[string]core.ProviderHealth) {
	s.statusMu.Lock()
	for k, v := range providerStatus {
		s.providerStatus[k] = v
	}
	for k, v := range providerHealth {
		s.providerHealth[k] = v
	}
	s.statusMu.Unlock()
}

func (s *Server) BroadcastUpdate() {
	s.clientsMu.RLock()
	clients := make([]*websocket.Conn, 0, len(s.clients))
	for conn := range s.clients {
		clients = append(clients, conn)
	}
	s.clientsMu.RUnlock()
	
	for _, conn := range clients {
		s.sendSnapshot(conn)
	}
}

// handleAnalyticsTrends serves trend data as JSON for the analytics dashboard
func (s *Server) handleAnalyticsTrends(w http.ResponseWriter, r *http.Request) {
	if s.trendAnalyzer == nil {
		http.Error(w, `{"error":"Analytics not available - storage not enabled"}`, http.StatusServiceUnavailable)
		return
	}

	// Get query parameters
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "all"
		// Use first available provider if "all" is requested
		if len(s.providerNames) > 0 {
			provider = s.providerNames[0]
		}
	}

	// Import storage to use TrendAnalyzer
	// This uses dynamic type assertion since we stored as interface{}
	type TrendGetter interface {
		GetAllTrends(ctx context.Context, provider string, period time.Duration) (map[string]interface{}, error)
	}

	analyzer, ok := s.trendAnalyzer.(TrendGetter)
	if !ok {
		http.Error(w, `{"error":"Invalid trend analyzer"}`, http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Get trends for last 7 days
	trends, err := analyzer.GetAllTrends(ctx, provider, 7*24*time.Hour)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Failed to get trends: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Convert to JSON-friendly format
	response := map[string]interface{}{
		"provider": provider,
		"period":   "7d",
		"trends":   trends,
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(response); err != nil {
		log.Printf("Failed to write analytics response: %v", err)
	}
}

func (s *Server) handleAlertsHistory(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
return
}

// Parse query parameters
query := r.URL.Query()
limitStr := query.Get("limit")
limit := 50 // default
if limitStr != "" {
if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
limit = parsed
if limit > 200 {
limit = 200 // max
}
}
}

// Get alerts from store
alerts := s.store.GetRecentAlerts(limit)

// Convert to JSON-friendly format
response := make([]map[string]interface{}, len(alerts))
for i, alert := range alerts {
response[i] = map[string]interface{}{
"rule_name":    alert.RuleName,
"condition":    alert.Condition,
"message":      alert.Message,
"severity":     alert.Severity,
"triggered_at": alert.TriggeredAt.Format(time.RFC3339),
"run": map[string]interface{}{
"id":            alert.Run.ID,
"provider":      alert.Run.Provider,
"repo":          alert.Run.Repo,
"workflow_name": alert.Run.WorkflowName,
"status":        alert.Run.Status,
"conclusion":    alert.Run.Conclusion,
"branch":        alert.Run.Branch,
"url":           alert.Run.URL,
},
}
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(response)
}
