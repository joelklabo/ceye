package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/joelklabo/ceye/internal/config"
	"github.com/joelklabo/ceye/internal/core"
	"github.com/joelklabo/ceye/internal/providers"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load(os.Getenv("CEYE_CONFIG"))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	deps := providers.Dependencies{}

	var providerInstances []core.Provider
	for _, pCfg := range cfg.Providers {
		provider, err := providers.CreateProvider(pCfg, deps)
		if err != nil {
			log.Fatalf("failed to create provider: %v", err)
		}
		providerInstances = append(providerInstances, provider)
	}

	store := core.NewStore()
	events := make(chan core.RunEvent)

	for _, provider := range providerInstances {
		go func(p core.Provider) {
			if err := p.Start(ctx, events); err != nil {
				fmt.Fprintf(os.Stderr, "provider %s exited with error: %v\n", p.Name(), err)
			}
		}(provider)
	}

	go func() {
		for event := range events {
			store.Merge(event)
			// TODO: notify UI once it exists.
		}
	}()

	// Placeholder Bubble Tea program.
	if _, err := tea.NewProgram(&placeholderModel{}).Run(); err != nil {
		log.Fatalf("bubble tea failed: %v", err)
	}
}

type placeholderModel struct{}

func (m *placeholderModel) Init() tea.Cmd                           { return nil }
func (m *placeholderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *placeholderModel) View() string                            { return "CI Status Dashboard coming soon" }
