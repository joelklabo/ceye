package webhooks

import (
"context"
"encoding/json"
"fmt"
"log"
"net/http"
"time"

"github.com/joelklabo/ceye/internal/core"
)

type Config struct {
Port         int
GitHubSecret string
AzureUser    string
AzurePass    string
}

type Server struct {
config  Config
server  *http.Server
events  chan core.RunEvent
mux     *http.ServeMux
}

func NewServer(cfg Config) *Server {
if cfg.Port == 0 {
cfg.Port = 9090
}

s := &Server{
config: cfg,
events: make(chan core.RunEvent, 100),
mux:    http.NewServeMux(),
}

s.setupRoutes()

s.server = &http.Server{
Addr:    fmt.Sprintf(":%d", cfg.Port),
Handler: s.mux,
}

return s
}

func (s *Server) setupRoutes() {
s.mux.HandleFunc("/webhooks/github", s.handleGitHub)
s.mux.HandleFunc("/webhooks/azure", s.handleAzure)
s.mux.HandleFunc("/webhooks/health", s.handleHealth)
s.mux.HandleFunc("/", s.handleRoot)
}

func (s *Server) Start(ctx context.Context) error {
go func() {
<-ctx.Done()
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
s.server.Shutdown(shutdownCtx)
}()

log.Printf("Webhook server starting on port %d", s.config.Port)
log.Printf("  GitHub endpoint: http://localhost:%d/webhooks/github", s.config.Port)
log.Printf("  Azure endpoint:  http://localhost:%d/webhooks/azure", s.config.Port)
log.Printf("  Health check:    http://localhost:%d/webhooks/health", s.config.Port)

if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
return err
}
return nil
}

func (s *Server) Events() <-chan core.RunEvent {
return s.events
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "text/html")
fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>ceye Webhook Server</title></head>
<body>
<h1>ceye Webhook Server</h1>
<p>Server is running and ready to receive webhooks.</p>
<h2>Endpoints:</h2>
<ul>
  <li>POST /webhooks/github - GitHub webhook receiver</li>
  <li>POST /webhooks/azure - Azure DevOps webhook receiver</li>
  <li>GET /webhooks/health - Health check</li>
</ul>
<p>Time: %s</p>
</body>
</html>`, time.Now().Format(time.RFC3339))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(map[string]interface{}{
"status":    "ok",
"timestamp": time.Now().Unix(),
"endpoints": []string{"/webhooks/github", "/webhooks/azure"},
})
}

func (s *Server) handleGitHub(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
return
}

eventType := r.Header.Get("X-GitHub-Event")
deliveryID := r.Header.Get("X-GitHub-Delivery")

log.Printf("Received GitHub webhook: %s (delivery: %s)", eventType, deliveryID)

// Only process workflow_run events
if eventType != "workflow_run" {
log.Printf("Ignoring event type: %s", eventType)
w.WriteHeader(http.StatusOK)
fmt.Fprintf(w, "Event type %s not processed", eventType)
return
}

// Read and parse body
var buf []byte
var err error
if r.Body != nil {
buf, err = json.Marshal(map[string]interface{}{})
decoder := json.NewDecoder(r.Body)
var temp interface{}
if err := decoder.Decode(&temp); err != nil {
log.Printf("Error reading body: %v", err)
http.Error(w, "Error reading body", http.StatusBadRequest)
return
}
buf, _ = json.Marshal(temp)
}

// Parse the webhook payload
run, err := ParseGitHubWebhook(buf)
if err != nil {
log.Printf("Error parsing GitHub webhook: %v", err)
http.Error(w, "Invalid payload", http.StatusBadRequest)
return
}

log.Printf("✅ Parsed GitHub webhook: %s/%s (status: %s, conclusion: %s)",
run.Repo, run.WorkflowName, run.Status, run.Conclusion)

// Emit RunEvent
event := core.RunEvent{
Provider:  "github", // Use same provider name as GitHub provider
Runs:      []core.Run{run},
Timestamp: time.Now(),
}

select {
case s.events <- event:
log.Printf("Sent RunEvent to channel")
default:
log.Printf("Warning: event channel full, dropping event")
}

w.WriteHeader(http.StatusOK)
fmt.Fprintf(w, "Webhook processed at %s", time.Now().Format(time.RFC3339))
}

func (s *Server) handleAzure(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
return
}

log.Printf("Received Azure DevOps webhook")

// Read and parse body
var buf []byte
var err error
if r.Body != nil {
decoder := json.NewDecoder(r.Body)
var temp interface{}
if err := decoder.Decode(&temp); err != nil {
log.Printf("Error reading body: %v", err)
http.Error(w, "Error reading body", http.StatusBadRequest)
return
}
buf, _ = json.Marshal(temp)
}

// Parse the webhook payload
run, err := ParseAzureWebhook(buf)
if err != nil {
log.Printf("Error parsing Azure webhook: %v", err)
http.Error(w, "Invalid payload", http.StatusBadRequest)
return
}

log.Printf("✅ Parsed Azure webhook: %s/%s (status: %s, conclusion: %s)",
run.Repo, run.WorkflowName, run.Status, run.Conclusion)

// Emit RunEvent
event := core.RunEvent{
Provider:  "azure-webhook",
Runs:      []core.Run{run},
Timestamp: time.Now(),
}

select {
case s.events <- event:
log.Printf("Sent RunEvent to channel")
default:
log.Printf("Warning: event channel full, dropping event")
}

w.WriteHeader(http.StatusOK)
fmt.Fprintf(w, "Webhook processed at %s", time.Now().Format(time.RFC3339))
}
