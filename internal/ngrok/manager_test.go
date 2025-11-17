package ngrok

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Helper to check if ngrok is installed
func isNgrokInstalled() bool {
	_, err := exec.LookPath("ngrok")
	return err == nil
}

// Helper to kill any running ngrok processes
func killNgrokProcesses() {
	cmd := exec.Command("pkill", "-f", "ngrok")
	_ = cmd.Run() // Ignore error if no ngrok processes are running
	time.Sleep(100 * time.Millisecond) // Give it a moment to terminate
}

func TestMain(m *testing.M) {
	// Ensure a clean state before and after tests
	killNgrokProcesses()
	code := m.Run()
	killNgrokProcesses()
	os.Exit(code)
}

func TestManager_Start_NewTunnel(t *testing.T) {
	if !isNgrokInstalled() {
		t.Skip("ngrok is not installed, skipping test")
	}

	// Use a unique port for this test to avoid conflicts
	testPort := 8081

	manager := NewManager()
	defer manager.Stop() // Ensure ngrok is stopped after the test

	url, err := manager.Start(testPort)
	if err != nil {
		t.Fatalf("Failed to start ngrok: %v", err)
	}

	if url == "" {
		t.Fatal("Expected ngrok URL, got empty string")
	}
	if !strings.HasPrefix(url, "https://") {
		t.Errorf("Expected URL to start with 'https://', got %s", url)
	}

	// Verify ngrok process is running
	cmd := exec.Command("pgrep", "-f", fmt.Sprintf("ngrok http %d", testPort))
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		t.Errorf("ngrok process for port %d not found running", testPort)
	}

	// Optionally, try to hit the ngrok URL to ensure it's live
	// This requires a local server to be running on testPort, which we don't have here.
	// For now, just checking the URL format and process existence is sufficient.
}

func TestManager_Start_ExistingTunnel(t *testing.T) {
	if !isNgrokInstalled() {
		t.Skip("ngrok is not installed, skipping test")
	}

	testPort := 8082
	
	// Manually start ngrok to simulate an existing tunnel
	cmd := exec.Command("ngrok", "http", strconv.Itoa(testPort))
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to manually start ngrok for existing tunnel test: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		time.Sleep(100 * time.Millisecond)
	}()
	
	// Give ngrok time to start and expose its API
	time.Sleep(2 * time.Second)

	// Verify the manually started ngrok is running
	pgrepCmd := exec.Command("pgrep", "-f", fmt.Sprintf("ngrok http %d", testPort))
	output, err := pgrepCmd.Output()
	if err != nil || len(output) == 0 {
		t.Fatalf("Manually started ngrok process for port %d not found running", testPort)
	}

	manager := NewManager()
	defer manager.Stop() // This stop should not affect the manually started ngrok

	url, err := manager.Start(testPort)
	if err != nil {
		t.Fatalf("Failed to start manager with existing ngrok: %v", err)
	}

	if url == "" {
		t.Fatal("Expected ngrok URL from existing tunnel, got empty string")
	}
	if !strings.HasPrefix(url, "https://") {
		t.Errorf("Expected URL to start with 'https://', got %s", url)
	}

	// Ensure the manager did NOT start a new ngrok process
	// We expect only one ngrok process for testPort
	pgrepCmd = exec.Command("pgrep", "-f", fmt.Sprintf("ngrok http %d", testPort))
	output, err = pgrepCmd.Output()
	if err != nil {
		t.Fatalf("Failed to find ngrok process after manager start: %v", err)
	}
	pids := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(pids) != 1 {
		t.Errorf("Expected 1 ngrok process for port %d, found %d: %v", testPort, len(pids), pids)
	}
}

func TestManager_Stop(t *testing.T) {
	if !isNgrokInstalled() {
		t.Skip("ngrok is not installed, skipping test")
	}

	testPort := 8083
	manager := NewManager()

	_, err := manager.Start(testPort)
	if err != nil {
		t.Fatalf("Failed to start ngrok: %v", err)
	}

	// Verify ngrok process is running
	cmd := exec.Command("pgrep", "-f", fmt.Sprintf("ngrok http %d", testPort))
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		t.Fatalf("ngrok process for port %d not found running before stop", testPort)
	}

	err = manager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop ngrok: %v", err)
	}

	// Verify ngrok process is no longer running
	time.Sleep(500 * time.Millisecond) // Give it a moment to terminate
	cmd = exec.Command("pgrep", "-f", fmt.Sprintf("ngrok http %d", testPort))
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		t.Errorf("ngrok process for port %d is still running after stop", testPort)
	}
}

func TestManager_GetURL(t *testing.T) {
	if !isNgrokInstalled() {
		t.Skip("ngrok is not installed, skipping test")
	}

	testPort := 8084
	manager := NewManager()
	defer manager.Stop()

	expectedURL := "https://test-url.ngrok.io"
	manager.url = expectedURL // Manually set for testing GetURL without starting ngrok

	if manager.GetURL() != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, manager.GetURL())
	}

	// Test after starting ngrok
	url, err := manager.Start(testPort)
	if err != nil {
		t.Fatalf("Failed to start ngrok: %v", err)
	}
	if manager.GetURL() != url {
		t.Errorf("Expected URL %s, got %s after start", url, manager.GetURL())
	}
}

// Test for getTunnelsFromAPI when ngrok API is not available
func TestManager_getTunnelsFromAPI_NoNgrokAPI(t *testing.T) {
	// Ensure ngrok is not running
	killNgrokProcesses()

	manager := NewManager()
	_, err := manager.getTunnelsFromAPI()
	if err == nil {
		t.Fatal("Expected an error when ngrok API is not available, got nil")
	}
	if !strings.Contains(err.Error(), "failed to query ngrok API") {
		t.Errorf("Expected 'failed to query ngrok API' error, got: %v", err)
	}
}

// Test for getTunnelURL when no tunnel for the specific port is found
func TestManager_getTunnelURL_NoMatchingTunnel(t *testing.T) {
	if !isNgrokInstalled() {
		t.Skip("ngrok is not installed, skipping test")
	}

	// Start ngrok on a different port
	cmd := exec.Command("ngrok", "http", "8089")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to manually start ngrok: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		time.Sleep(100 * time.Millisecond)
	}()
	time.Sleep(2 * time.Second) // Give ngrok time to start

	manager := NewManager()
	_, err := manager.getTunnelURL() // Should look for 0 by default, or whatever m.port is set to
	if err == nil {
		t.Fatal("Expected an error when no matching tunnel is found, got nil")
	}
	if !strings.Contains(err.Error(), "ngrok tunnel for port") {
		t.Errorf("Expected 'ngrok tunnel for port' error, got: %v", err)
	}
}

// Test for detectExisting when a tunnel exists
func TestManager_detectExisting_TunnelExists(t *testing.T) {
	if !isNgrokInstalled() {
		t.Skip("ngrok is not installed, skipping test")
	}

	testPort := 8085
	// Manually start ngrok to simulate an existing tunnel
	cmd := exec.Command("ngrok", "http", strconv.Itoa(testPort))
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to manually start ngrok for detectExisting test: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		time.Sleep(100 * time.Millisecond)
	}()
	time.Sleep(2 * time.Second) // Give ngrok time to start

	manager := NewManager()
	manager.port = testPort // Set the port for detectExisting to look for
	
	detectedURL := manager.detectExisting()
	if detectedURL == "" {
		t.Fatal("Expected to detect an existing ngrok tunnel, got empty URL")
	}
	if !strings.HasPrefix(detectedURL, "https://") {
		t.Errorf("Expected detected URL to start with 'https://', got %s", detectedURL)
	}
}

// Test for detectExisting when no tunnel exists for the target port
func TestManager_detectExisting_NoTunnel(t *testing.T) {
	if !isNgrokInstalled() {
		t.Skip("ngrok is not installed, skipping test")
	}

	// Ensure no ngrok is running on the test port
	killNgrokProcesses()

	manager := NewManager()
	manager.port = 8086 // A port where no ngrok is running

	detectedURL := manager.detectExisting()
	if detectedURL != "" {
		t.Errorf("Expected no existing ngrok tunnel, but detected: %s", detectedURL)
	}
}

// Test for detectExisting when ngrok API is not available
func TestManager_detectExisting_NoNgrokAPI(t *testing.T) {
	// Ensure ngrok is not running
	killNgrokProcesses()

	manager := NewManager()
	manager.port = 8087

	detectedURL := manager.detectExisting()
	if detectedURL != "" {
		t.Errorf("Expected no existing ngrok tunnel when API is down, but detected: %s", detectedURL)
	}
}

// Test for Start when ngrok is not installed
func TestManager_Start_NgrokNotInstalled(t *testing.T) {
	// Temporarily rename ngrok to simulate it not being installed
	ngrokPath, err := exec.LookPath("ngrok")
	if err != nil {
		t.Skip("ngrok is not installed, cannot test 'not installed' scenario")
	}
	tempNgrokPath := ngrokPath + ".bak"
	err = os.Rename(ngrokPath, tempNgrokPath)
	if err != nil {
		t.Fatalf("Failed to rename ngrok: %v", err)
	}
	defer func() {
		_ = os.Rename(tempNgrokPath, ngrokPath) // Restore ngrok
	}()

	manager := NewManager()
	_, err = manager.Start(8088)
	if err == nil {
		t.Fatal("Expected an error when ngrok is not installed, got nil")
	}
	if !strings.Contains(err.Error(), "failed to start ngrok") {
		t.Errorf("Expected 'failed to start ngrok' error, got: %v", err)
	}
}
