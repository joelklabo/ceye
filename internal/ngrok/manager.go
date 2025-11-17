package ngrok

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

// Manager handles starting and managing an ngrok tunnel.
type Manager struct {
	cmd    *exec.Cmd
	url    string
	port   int
}

// NewManager creates a new ngrok Manager.
func NewManager() *Manager {
	return &Manager{}
}

// Start attempts to start an ngrok tunnel for the given port.
// It returns the public URL of the tunnel or an error.
func (m *Manager) Start(port int) (string, error) {
	m.port = port

	// Check if ngrok is already running and a tunnel exists
	if url := m.detectExisting(); url != "" {
		m.url = url
		return url, nil
	}

	// Start ngrok
	m.cmd = exec.Command("ngrok", "http", strconv.Itoa(port))
	// We don't want ngrok's output to pollute ceye's stdout/stderr,
	// so we'll redirect it to /dev/null or a log file later if needed.
	// For now, we'll let it run in the background.
	if err := m.cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ngrok: %w", err)
	}

	// Wait for ngrok to start and expose its API
	// This is a heuristic; a more robust solution would poll the ngrok API
	time.Sleep(2 * time.Second)

	url, err := m.getTunnelURL()
	if err != nil {
		// If we failed to get the tunnel URL, it's likely ngrok didn't start correctly
		// or the API isn't ready. Try to kill the process.
		if m.cmd != nil && m.cmd.Process != nil {
			_ = m.cmd.Process.Kill()
		}
		return "", err
	}

	m.url = url
	return url, nil
}

// detectExisting checks if an ngrok tunnel is already running for the target port
// by querying the ngrok API.
func (m *Manager) detectExisting() string {
	tunnels, err := m.getTunnelsFromAPI()
	if err != nil {
		return ""
	}

	for _, tunnel := range tunnels {
		if tunnel.Proto == "https" && tunnel.Config.Addr == fmt.Sprintf("http://localhost:%d", m.port) {
			return tunnel.PublicURL
		}
	}
	return ""
}

// getTunnelURL queries the ngrok API to get the public URL of the tunnel.
func (m *Manager) getTunnelURL() (string, error) {
	tunnels, err := m.getTunnelsFromAPI()
	if err != nil {
		return "", err
	}

	for _, tunnel := range tunnels {
		if tunnel.Proto == "https" && tunnel.Config.Addr == fmt.Sprintf("http://localhost:%d", m.port) {
			return tunnel.PublicURL, nil
		}
	}

	return "", fmt.Errorf("ngrok tunnel for port %d not found", m.port)
}

// getTunnelsFromAPI fetches the list of tunnels from the ngrok API.
func (m *Manager) getTunnelsFromAPI() ([]ngrokTunnel, error) {
	resp, err := http.Get("http://localhost:4040/api/tunnels")
	if err != nil {
		return nil, fmt.Errorf("failed to query ngrok API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ngrok API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ngrok API response: %w", err)
	}

	var apiResponse struct {
		Tunnels []ngrokTunnel `json:"tunnels"`
	}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse ngrok API response: %w", err)
	}

	return apiResponse.Tunnels, nil
}

// ngrokTunnel represents a single tunnel returned by the ngrok API.
type ngrokTunnel struct {
	PublicURL string `json:"public_url"`
	Proto     string `json:"proto"`
	Config    struct {
		Addr string `json:"addr"`
	} `json:"config"`
}

// Stop terminates the ngrok process if it was started by this manager.
func (m *Manager) Stop() error {
	if m.cmd != nil && m.cmd.Process != nil {
		return m.cmd.Process.Kill()
	}
	return nil
}

// GetURL returns the public URL of the ngrok tunnel.
func (m *Manager) GetURL() string {
	return m.url
}
