package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joelklabo/ceye/internal/core"
)

//go:embed web
var webAssets embed.FS

type Server struct {
	store     *core.Store
	port      int
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex
	upgrader  websocket.Upgrader
	
	providerStatus map[string]string
	providerHealth map[string]core.ProviderHealth
	providerNames  []string
	statusMu       sync.RWMutex
}

type Message struct {
	Type      string                        `json:"type"`
	Timestamp time.Time                     `json:"timestamp"`
	Runs      []core.Run                    `json:"runs,omitempty"`
	Providers []string                      `json:"providers,omitempty"`
	Status    map[string]string             `json:"status,omitempty"`
	Health    map[string]core.ProviderHealth `json:"health,omitempty"`
	Totals    map[string]int                `json:"totals,omitempty"`
}

func New(store *core.Store, providerNames []string, port int) *Server {
	return &Server{
		store:          store,
		port:           port,
		clients:        make(map[*websocket.Conn]bool),
		providerNames:  providerNames,
		providerStatus: make(map[string]string),
		providerHealth: make(map[string]core.ProviderHealth),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *Server) getWebFS() (fs.FS, error) {
	return fs.Sub(webAssets, "web")
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	
	// WebSocket endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)
	
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
	return server.ListenAndServe()
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
		Type:      "runs_update",
		Timestamp: time.Now(),
		Runs:      runs,
		Providers: s.providerNames,
		Status:    status,
		Health:    health,
		Totals:    totals,
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
