package main

import (
"context"
"fmt"
"log"
"os"
"os/signal"
"syscall"

"github.com/joelklabo/ceye/internal/webhooks"
)

func main() {
fmt.Println("=== ceye Webhook Server Test ===")
fmt.Println()
fmt.Println("This test server will:")
fmt.Println("1. Start a webhook server on port 9090")
fmt.Println("2. Wait for you to expose it via ngrok")
fmt.Println("3. Receive and log webhook payloads")
fmt.Println()

// Create webhook server
config := webhooks.Config{
Port:         9090,
GitHubSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
AzureUser:    os.Getenv("AZURE_WEBHOOK_USER"),
AzurePass:    os.Getenv("AZURE_WEBHOOK_PASS"),
}

server := webhooks.NewServer(config)

// Setup context for graceful shutdown
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Handle Ctrl+C
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
go func() {
<-sigChan
fmt.Println("\nShutting down...")
cancel()
}()

// Start listening for events (just log them for now)
go func() {
for event := range server.Events() {
log.Printf("Received event from %s provider with %d runs", event.Provider, len(event.Runs))
}
}()

fmt.Println()
fmt.Println("Next steps:")
fmt.Println("1. In another terminal, run: ngrok http 9090")
fmt.Println("2. Copy the https URL (e.g., https://abc123.ngrok.io)")
fmt.Println("3. Go to GitHub repo → Settings → Webhooks → Add webhook")
fmt.Println("4. Paste URL + /webhooks/github (e.g., https://abc123.ngrok.io/webhooks/github)")
fmt.Println("5. Content type: application/json")
fmt.Println("6. Select 'Let me select individual events' → workflow_run")
fmt.Println("7. Trigger a workflow and watch the logs here!")
fmt.Println()

// Start server (blocks until shutdown)
if err := server.Start(ctx); err != nil {
log.Fatalf("Server error: %v", err)
}

fmt.Println("Server stopped")
}
