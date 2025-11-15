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
	"github.com/joelklabo/ceye/internal/ui"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfgPath := os.Getenv("CEYE_CONFIG")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	deps := providers.Dependencies{}

	var providerInstances []core.Provider
	var providerNames []string
	for _, pCfg := range cfg.Providers {
		provider, err := providers.CreateProvider(pCfg, deps)
		if err != nil {
			log.Fatalf("failed to create provider: %v", err)
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
		log.Fatalf("bubble tea failed: %v", err)
	}
}
