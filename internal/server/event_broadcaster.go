package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

// TimelineEvent represents a single event for the timeline stream.
type TimelineEvent struct {
	Timestamp string `json:"timestamp"` // ISO string
	Type      string `json:"type"`      // e.g., "websocket_message", "webhook_received", "provider_poll"
	Category  string `json:"category"`  // e.g., "WebSocket", "Webhook", "Polling", "UI", "Error"
	Message   string `json:"message"`   // Short description
	Details   any    `json:"details,omitempty"` // Full event payload
	ID        string `json:"id"`        // Unique ID for React keys
}

// EventBroadcaster manages WebSocket connections for event streaming and broadcasts event messages.
type EventBroadcaster struct {
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex
	writeMu   map[*websocket.Conn]*sync.Mutex // Mutex for each client's write operations
	writeMuMu sync.RWMutex                    // Mutex to protect the writeMu map
	logger    *CustomLogger
}

// NewEventBroadcaster creates and returns a new EventBroadcaster.
func NewEventBroadcaster(logger *CustomLogger) *EventBroadcaster {
	return &EventBroadcaster{
		clients: make(map[*websocket.Conn]bool),
		writeMu: make(map[*websocket.Conn]*sync.Mutex),
		logger:  logger,
	}
}

// SetLogger sets the CustomLogger for the EventBroadcaster.
func (eb *EventBroadcaster) SetLogger(logger *CustomLogger) {
	eb.logger = logger
}

// AddClient adds a new WebSocket connection to the broadcaster.
func (eb *EventBroadcaster) AddClient(conn *websocket.Conn) {
	eb.clientsMu.Lock()
	eb.clients[conn] = true
	eb.clientsMu.Unlock()

	eb.writeMuMu.Lock()
	eb.writeMu[conn] = &sync.Mutex{}
	eb.writeMuMu.Unlock()
	if eb.logger != nil {
		eb.logger.Info("EventBroadcaster: Client added. Total clients: %d", len(eb.clients))
	}
}

// RemoveClient removes a WebSocket connection from the broadcaster.
func (eb *EventBroadcaster) RemoveClient(conn *websocket.Conn) {
	eb.clientsMu.Lock()
	delete(eb.clients, conn)
	eb.clientsMu.Unlock()

	eb.writeMuMu.Lock()
	delete(eb.writeMu, conn)
	eb.writeMuMu.Unlock()
	if eb.logger != nil {
		eb.logger.Info("EventBroadcaster: Client removed. Total clients: %d", len(eb.clients))
	}
}

// Broadcast sends a TimelineEvent to all connected clients.
func (eb *EventBroadcaster) Broadcast(event TimelineEvent) {
	eb.clientsMu.RLock()
	clients := make([]*websocket.Conn, 0, len(eb.clients))
	for conn := range eb.clients {
		clients = append(clients, conn)
	}
	eb.clientsMu.RUnlock()

	for _, conn := range clients {
		if err := eb.writeJSON(conn, event); err != nil {
			if eb.logger != nil {
				eb.logger.Error("EventBroadcaster: Failed to send event message to client: %v", err)
			}
			eb.RemoveClient(conn) // Remove client if write fails
		}
	}
}

// writeJSON is a helper to safely write JSON to a WebSocket connection.
func (eb *EventBroadcaster) writeJSON(conn *websocket.Conn, msg interface{}) error {
	eb.writeMuMu.RLock()
	mu, ok := eb.writeMu[conn]
	eb.writeMuMu.RUnlock()
	if !ok {
		return nil // Client might have been removed concurrently
	}

	mu.Lock()
	defer mu.Unlock()
	return conn.WriteJSON(msg)
}
