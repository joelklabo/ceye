package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/joelklabo/ceye/internal/config"
	"github.com/joelklabo/ceye/internal/core"
	"github.com/joelklabo/ceye/internal/providers"
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

	deps := providers.Dependencies{}

	var providerInstances []core.Provider
	var providerNames []string
	for _, pCfg := range cfg.Providers {
		provider, err := providers.CreateProvider(pCfg, deps)
		if err != nil {
			return fmt.Errorf("create provider: %w", err)
		}
		providerInstances = append(providerInstances, provider)
		providerNames = append(providerNames, provider.Name())
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

	model := ui.NewModel(store, providerNames)
	program := tea.NewProgram(model)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventCh:
				store.Merge(event)
				program.Send(ui.RunUpdatedMsg{})
			}
		}
	}()

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run UI: %w", err)
	}
	return nil
}
