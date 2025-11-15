package azure

import (
	"context"

	"github.com/joelklabo/ceye/internal/core"
)

// AzureClient abstracts Azure DevOps API interactions used by the provider.
type AzureClient interface {
	ListBuilds(org, project string, pipelines []int) ([]core.Run, error)
}

// Config describes the Azure DevOps project/pipeline selection.
type Config struct {
	Org       string
	Project   string
	Pipelines []int
}

// Provider polls Azure DevOps builds for configured pipelines.
type Provider struct {
	client AzureClient
	cfg    Config
}

// NewProvider constructs an Azure provider.
func NewProvider(client AzureClient, cfg Config) *Provider {
	return &Provider{client: client, cfg: cfg}
}

// Name implements core.Provider.
func (p *Provider) Name() string {
	return "azure"
}

// Start begins polling Azure for builds. Implementation to follow in Step 12.
func (p *Provider) Start(ctx context.Context, out chan<- core.RunEvent) error {
	<-ctx.Done()
	return ctx.Err()
}
