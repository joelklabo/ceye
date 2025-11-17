package server

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// LogMessage represents a single log entry for the unified log stream.
type LogMessage struct {
	Type      string    `json:"type"` // "log_entry"
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`     // e.g., "INFO", "WARN", "ERROR"
	Component string    `json:"component"` // e.g., "server", "provider", "websocket"
	Message   string    `json:"message"`
}

// LogBroadcaster manages WebSocket connections for log streaming and broadcasts log messages.
type LogBroadcaster struct {
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex
	writeMu   map[*websocket.Conn]*sync.Mutex // Mutex for each client's write operations
	writeMuMu sync.RWMutex                    // Mutex to protect the writeMu map
	logger    *CustomLogger
}

// NewLogBroadcaster creates and returns a new LogBroadcaster.
func NewLogBroadcaster(logger *CustomLogger) *LogBroadcaster {
	return &LogBroadcaster{
		clients: make(map[*websocket.Conn]bool),
		writeMu: make(map[*websocket.Conn]*sync.Mutex),
		logger:  logger,
	}
}

// AddClient adds a new WebSocket connection to the broadcaster.
func (lb *LogBroadcaster) AddClient(conn *websocket.Conn) {
	lb.clientsMu.Lock()
	lb.clients[conn] = true
	lb.clientsMu.Unlock()

	lb.writeMuMu.Lock()
	lb.writeMu[conn] = &sync.Mutex{}
	lb.writeMuMu.Unlock()
	if lb.logger != nil {
		lb.logger.Info("LogBroadcaster: Client added. Total clients: %d", len(lb.clients))
	}
}

// RemoveClient removes a WebSocket connection from the broadcaster.
func (lb *LogBroadcaster) RemoveClient(conn *websocket.Conn) {
	lb.clientsMu.Lock()
	delete(lb.clients, conn)
	lb.clientsMu.Unlock()

	lb.writeMuMu.Lock()
	delete(lb.writeMu, conn)
	lb.writeMuMu.Unlock()
	if lb.logger != nil {
		lb.logger.Info("LogBroadcaster: Client removed. Total clients: %d", len(lb.clients))
	}
}

// Broadcast sends a LogMessage to all connected clients.
func (lb *LogBroadcaster) Broadcast(msg LogMessage) {
	lb.clientsMu.RLock()
	clients := make([]*websocket.Conn, 0, len(lb.clients))
	for conn := range lb.clients {
		clients = append(clients, conn)
	}
	lb.clientsMu.RUnlock()

	for _, conn := range clients {
		err := lb.writeJSON(conn, msg)
		if err != nil {
			if lb.logger != nil {
				lb.logger.Error("LogBroadcaster: Failed to send log message to client: %v", err)
			}
			lb.RemoveClient(conn) // Remove client if write fails
		}
	}
}

// writeJSON is a helper to safely write JSON to a WebSocket connection.
func (lb *LogBroadcaster) writeJSON(conn *websocket.Conn, msg interface{}) error {
	lb.writeMuMu.RLock()
	mu, ok := lb.writeMu[conn]
	lb.writeMuMu.RUnlock()
	if !ok {
		// Client might have been removed concurrently - this is not an error
		return nil
	}

	mu.Lock()
	defer mu.Unlock()
	return conn.WriteJSON(msg)
}

// SetLogger sets the CustomLogger for the LogBroadcaster.
func (lb *LogBroadcaster) SetLogger(logger *CustomLogger) {
	lb.logger = logger
}
