package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/joelklabo/ceye/internal/config"
	"github.com/joelklabo/ceye/internal/core"
	"github.com/joelklabo/ceye/internal/providers"
	azureprovider "github.com/joelklabo/ceye/internal/providers/azure"
	githubprovider "github.com/joelklabo/ceye/internal/providers/github"
	"github.com/joelklabo/ceye/internal/ui"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var cfgPath string
	var demo bool
	var demoRuns int
	var demoDuration time.Duration
	var eventLogPath string
	rootCmd := &cobra.Command{
		Use:   "ci-dash",
		Short: "CI Status Dashboard TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ctx, cfgPath, demo, demoRuns, demoDuration, eventLogPath)
		},
	}
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "Path to config file (defaults to ceye.yaml search paths)")
	rootCmd.PersistentFlags().BoolVar(&demo, "demo", false, "Run with the built-in demo provider (ignores config)")
	rootCmd.PersistentFlags().IntVar(&demoRuns, "demo-runs", 4, "Number of demo runs when --demo is set")
	rootCmd.PersistentFlags().DurationVar(&demoDuration, "demo-duration", 0, "Automatically exit demo mode after this duration (e.g. 5s)")
	rootCmd.PersistentFlags().StringVar(&eventLogPath, "log-events", "", "Write RunEvent JSON lines to the given file")

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func run(parentCtx context.Context, cfgPath string, demo bool, demoRuns int, demoDuration time.Duration, eventLogPath string) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	if demoDuration > 0 && demo {
		timer := time.NewTimer(demoDuration)
		defer timer.Stop()
		go func() {
			select {
			case <-timer.C:
				cancel()
			case <-ctx.Done():
			}
		}()
	}

	var cfg *config.Config
	var err error
	if demo {
		if demoRuns <= 0 {
			demoRuns = 4
		}
		cfg = &config.Config{
			Providers: []providers.ProviderConfig{{Type: "demo", Runs: demoRuns}},
		}
	} else {
		if cfgPath == "" {
			cfgPath = os.Getenv("CEYE_CONFIG")
		}

		cfg, err = config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
	}

	var eventLog io.WriteCloser
	if eventLogPath != "" {
		f, err := os.Create(eventLogPath)
		if err != nil {
			return fmt.Errorf("open event log: %w", err)
		}
		eventLog = f
		defer eventLog.Close()
	}

	deps := providers.Dependencies{
		GitHubClient: githubprovider.NewHTTPClient(githubToken()),
		AzureClient:  azureprovider.NewHTTPClient(azureToken()),
	}

	var providerInstances []core.Provider
	var providerNames []string
	var refreshers []func()
	providerStatus := make(map[string]string)
	providerTimes := make(map[string]time.Time)
	providerHealth := make(map[string]core.ProviderHealth)
	providerLastPoll := make(map[string]time.Time)
	providerLag := make(map[string]time.Duration)
	for _, pCfg := range cfg.Providers {
		provider, err := providers.CreateProvider(pCfg, deps)
		if err != nil {
			return fmt.Errorf("create provider: %w", err)
		}
		providerInstances = append(providerInstances, provider)
		providerNames = append(providerNames, provider.Name())
		if refresher, ok := provider.(interface{ TriggerRefresh() }); ok {
			refreshers = append(refreshers, refresher.TriggerRefresh)
		}
		providerStatus[provider.Name()] = ""
	}

	store := core.NewStore()
	eventCh := make(chan core.RunEvent)

	for _, provider := range providerInstances {
		go func(p core.Provider) {
			if err := p.Start(ctx, eventCh); err != nil && ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "provider %s exited with error: %v\n", p.Name(), err)
			}
		}(provider)
	}

	refresh := func() {
		for _, fn := range refreshers {
			fn()
		}
	}

	model := ui.NewModel(store, providerNames, refresh, openURL, copyToClipboard)
	program := tea.NewProgram(model)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventCh:
				if eventLog != nil {
					if err := writeEventLog(eventLog, event); err != nil {
						fmt.Fprintf(os.Stderr, "log event: %v\n", err)
					}
				}
				store.Merge(event)
				message := ""
				level := ""
				ts := event.Timestamp
				if ts.IsZero() {
					ts = time.Now()
				}
					if event.Provider != "" {
						if event.Err != nil {
							providerStatus[event.Provider] = event.Err.Error()
							message = fmt.Sprintf("%s: %v", event.Provider, event.Err)
							level = "error"
							health := providerHealth[event.Provider]
							health.ErrorCount++
							health.LastError = ts
							providerHealth[event.Provider] = health
						} else {
							providerStatus[event.Provider] = ""
							switch {
							case event.Message != "":
								message = fmt.Sprintf("%s: %s", event.Provider, event.Message)
							case len(event.Runs) > 0:
								message = fmt.Sprintf("%s refreshed %d run(s)", event.Provider, len(event.Runs))
							default:
								message = fmt.Sprintf("%s refreshed", event.Provider)
							}
							health := providerHealth[event.Provider]
							health.LastSuccess = ts
							health.ErrorCount = 0
							providerHealth[event.Provider] = health
							if last, ok := providerLastPoll[event.Provider]; ok && last.After(time.Time{}) {
								delta := ts.Sub(last)
								providerLag[event.Provider] = delta
								if delta > 10*time.Second {
									message = fmt.Sprintf("%s (slow poll %s)", message, delta.Round(time.Second))
								}
							}
							level = "info"
						}
						providerLastPoll[event.Provider] = ts
						providerTimes[event.Provider] = ts
					} else if event.Message != "" {
						message = event.Message
						if event.Err != nil {
							level = "error"
						} else {
							level = "info"
						}
					}
				program.Send(ui.RunUpdatedMsg{
					Timestamp: ts,
					Status:    copyStatus(providerStatus),
					Times:     copyTimes(providerTimes),
					Message:   message,
					Level:     level,
					Health:    copyHealth(providerHealth),
					Lag:       copyLag(providerLag),
				})
			}
		}
	}()

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run UI: %w", err)
	}
	return nil
}

func githubToken() string {
	if token := os.Getenv("CEYE_GITHUB_TOKEN"); token != "" {
		return token
	}
	return os.Getenv("GITHUB_TOKEN")
}

func azureToken() string {
	if token := os.Getenv("CEYE_AZURE_PAT"); token != "" {
		return token
	}
	return os.Getenv("AZURE_DEVOPS_PAT")
}

func copyStatus(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyTimes(in map[string]time.Time) map[string]time.Time {
	out := make(map[string]time.Time, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyHealth(in map[string]core.ProviderHealth) map[string]core.ProviderHealth {
	out := make(map[string]core.ProviderHealth, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyLag(in map[string]time.Duration) map[string]time.Duration {
	out := make(map[string]time.Duration, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func openURL(link string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", link)
	case "darwin":
		cmd = exec.Command("open", link)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", link)
	default:
		return
	}
	cmd.Start()
}

func copyToClipboard(text string) {
	if text == "" {
		return
	}
	if err := clipboard.WriteAll(text); err != nil {
		fmt.Fprintf(os.Stderr, "copy to clipboard: %v\n", err)
	}
}

func writeEventLog(w io.Writer, event core.RunEvent) error {
	entry := struct {
		Timestamp time.Time `json:"timestamp"`
		Provider  string    `json:"provider"`
		Runs      int       `json:"runs"`
		Error     string    `json:"error,omitempty"`
		Message   string    `json:"message,omitempty"`
	}{
		Timestamp: event.Timestamp,
		Provider:  event.Provider,
		Runs:      len(event.Runs),
		Message:   event.Message,
	}
	if event.Err != nil {
		entry.Error = event.Err.Error()
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if _, err := w.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}
