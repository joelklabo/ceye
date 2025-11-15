package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

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
	rootCmd := &cobra.Command{
		Use:   "ci-dash",
		Short: "CI Status Dashboard TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ctx, cfgPath)
		},
	}
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "Path to config file (defaults to ceye.yaml search paths)")

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context, cfgPath string) error {
	if cfgPath == "" {
		cfgPath = os.Getenv("CEYE_CONFIG")
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	deps := providers.Dependencies{
		GitHubClient: githubprovider.NewHTTPClient(githubToken()),
		AzureClient:  azureprovider.NewHTTPClient(azureToken()),
	}

	var providerInstances []core.Provider
	var providerNames []string
	var refreshers []func()
	providerStatus := make(map[string]string)
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

	model := ui.NewModel(store, providerNames, refresh, openURL)
	program := tea.NewProgram(model)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventCh:
				store.Merge(event)
				if event.Provider != "" {
					if event.Err != nil {
						providerStatus[event.Provider] = event.Err.Error()
					} else {
						providerStatus[event.Provider] = ""
					}
				}
				ts := event.Timestamp
				if ts.IsZero() {
					ts = time.Now()
				}
				program.Send(ui.RunUpdatedMsg{Timestamp: ts, Status: copyStatus(providerStatus)})
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
