package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/joelklabo/ceye/internal/config"
	"github.com/joelklabo/ceye/internal/core"
	"github.com/joelklabo/ceye/internal/alerting"
	"github.com/joelklabo/ceye/internal/providers"
	azureprovider "github.com/joelklabo/ceye/internal/providers/azure"
	githubprovider "github.com/joelklabo/ceye/internal/providers/github"
	"github.com/joelklabo/ceye/internal/providers/manager"
	"github.com/joelklabo/ceye/internal/server"
	"github.com/joelklabo/ceye/internal/storage"
	"github.com/joelklabo/ceye/internal/ui"
	"github.com/joelklabo/ceye/internal/webhooks"
)

const (
	envConfigRoot   = "CEYE_CONFIG_ROOT"
	envGithubOrg    = "CEYE_GITHUB_ORG"
	envAzureOrg     = "CEYE_AZURE_ORG"
	envAzureProject = "CEYE_AZURE_PROJECT"

	defaultGithubOrg    = "joelklabo"
	defaultAzureOrg     = "joelklabo"
	defaultAzureProject = "Big Timer"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

var (
	providerStoreFlag string
	configDirFlag     string
	githubOrgFlag     string
	azureOrgFlag      string
	azureProjectFlag  string
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var cfgPath string
	var demo bool
	var demoRuns int
	var demoDuration time.Duration
	var eventLogPath string
	var notify bool
	var historyPath string
	var webhookURL string
	var web bool
	var webPort int
	var alertDebug bool
	var enableWebhooks = true  // Enable webhooks by default
	var webhookPort int
	var webhookSecret string
	rootCmd := &cobra.Command{
		Use:     "ceye",
		Short:   "CI Status Dashboard TUI",
		Version: fmt.Sprintf("%s (%s)", Version, GitCommit),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ctx, cfgPath, configDirFlag, demo, demoRuns, demoDuration, eventLogPath, notify, historyPath, webhookURL, resolveProviderStorePath(providerStoreFlag), githubOrgFlag, azureOrgFlag, azureProjectFlag, web, webPort, enableWebhooks, webhookPort, webhookSecret)
		},
	}
	rootCmd.SetVersionTemplate("ceye version {{.Version}}\n")
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "Path to config file (defaults to ceye.yaml search paths)")
	rootCmd.PersistentFlags().BoolVar(&demo, "demo", false, "Run with the built-in demo provider (ignores config)")
	rootCmd.PersistentFlags().IntVar(&demoRuns, "demo-runs", 4, "Number of demo runs when --demo is set")
	rootCmd.PersistentFlags().DurationVar(&demoDuration, "demo-duration", 0, "Automatically exit demo mode after this duration (e.g. 5s)")
	rootCmd.PersistentFlags().StringVar(&eventLogPath, "log-events", "", "Write RunEvent JSON lines to the given file")
	rootCmd.PersistentFlags().BoolVar(&notify, "notify", false, "Emit desktop notifications when providers error")
	rootCmd.PersistentFlags().StringVar(&webhookURL, "webhook-url", "", "POST provider errors to this webhook URL")
	rootCmd.PersistentFlags().StringVar(&historyPath, "history-path", "", "Persist run history to this JSON file (defaults to ~/.config/ceye/run-history.json)")
	rootCmd.PersistentFlags().StringVar(&providerStoreFlag, "provider-store", "", "Path to provider store (defaults to ~/.config/ceye/providers.json)")
	rootCmd.PersistentFlags().StringVar(&configDirFlag, "config-dir", ".", "Root directory to scan for config files (defaults to current dir)")
	rootCmd.PersistentFlags().StringVar(&githubOrgFlag, "github-org", "", "GitHub org used when auto-discovering repos via `gh repo list`")
	rootCmd.PersistentFlags().StringVar(&azureOrgFlag, "azure-org", "", "Azure DevOps org (or URL) used when auto-discovering pipelines via `az pipelines list`")
	rootCmd.PersistentFlags().StringVar(&azureProjectFlag, "azure-project", "", "Azure DevOps project scanned when auto-discovering pipelines")
	rootCmd.PersistentFlags().BoolVar(&web, "web", false, "Start web server instead of TUI")
	rootCmd.PersistentFlags().IntVar(&webPort, "port", 8080, "Port for web server (requires --web)")
	rootCmd.PersistentFlags().BoolVar(&alertDebug, "alert-debug", false, "Enable verbose alert debugging (logs all rule evaluations)")
	rootCmd.PersistentFlags().BoolVar(&enableWebhooks, "webhooks", true, "Enable webhook receiver for push-based updates (default: true)")
	rootCmd.PersistentFlags().IntVar(&webhookPort, "webhook-port", 9090, "Port for webhook server")
	rootCmd.PersistentFlags().StringVar(&webhookSecret, "webhook-secret", "", "GitHub webhook secret for signature verification")

	rootCmd.AddCommand(providerCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show detailed version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ceye version:\n")
			fmt.Printf("  Version:    %s\n", Version)
			fmt.Printf("  Commit:     %s\n", GitCommit)
			fmt.Printf("  Build Time: %s\n", BuildTime)
			fmt.Printf("  Go Version: %s\n", runtime.Version())
			fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
			
			// Check if running from expected location
			exe, err := os.Executable()
			if err == nil {
				fmt.Printf("  Executable: %s\n", exe)
			}
			
			// Git status
			gitCmd := exec.Command("git", "status", "--porcelain")
			gitCmd.Dir = "/Users/honk/code/ceye"
			if output, err := gitCmd.Output(); err == nil && len(output) > 0 {
				fmt.Printf("  Git Status: DIRTY (uncommitted changes)\n")
			} else {
				fmt.Printf("  Git Status: clean\n")
			}
		},
	}
}

func run(parentCtx context.Context, cfgPath, configDir string, demo bool, demoRuns int, demoDuration time.Duration, eventLogPath string, notify bool, historyPath string, webhookURL string, providerStorePath string, githubOrg, azureOrg, azureProject string, web bool, webPort int, enableWebhooks bool, webhookPort int, webhookSecret string) error {
	// Print version info immediately
	exe, _ := os.Executable()
	fmt.Fprintf(os.Stderr, "🚀 ceye %s (%s) starting...\n", Version, GitCommit)
	fmt.Fprintf(os.Stderr, "   Build: %s\n", BuildTime)
	fmt.Fprintf(os.Stderr, "   Binary: %s\n", exe)
	fmt.Fprintf(os.Stderr, "\n")
	
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
	var missingConfigs []string
	var err error
	var configRoot string
	var missingConfigMu sync.RWMutex
	providerHistory := make(map[string][]string)

	if demo {
		if demoRuns <= 0 {
			demoRuns = 4
		}
		// Load config first to preserve alerting settings
		var loadedCfg *config.Config
		if cfgPath != "" {
			loadedCfg, _ = config.Load(cfgPath)
		}
		
		// Create demo config, but preserve alerting from loaded config
		cfg = &config.Config{
			Providers: []providers.ProviderConfig{{Type: "demo", Runs: demoRuns}},
		}
		if loadedCfg != nil {
			cfg.Alerting = loadedCfg.Alerting
		}
	} else {
		if cfgPath == "" {
			cfgPath = os.Getenv("CEYE_CONFIG")
		}

		cfg, configRoot, missingConfigs, err = loadDistributedConfigs(cfgPath, configDir, githubOrg, azureOrg, azureProject)
		if err != nil {
			return err
		}
		if len(missingConfigs) > 0 {
			providerHistory["missing"] = missingConfigs
		}
	}

	providerStore, err := manager.New(providerStorePath)
	if err != nil {
		return fmt.Errorf("provider store: %w", err)
	}
	if configRoot != "" {
		if list, listErr := listMissingConfigs(configRoot); listErr != nil {
			fmt.Fprintf(os.Stderr, "missing configs: %v\n", listErr)
		} else {
			missingConfigMu.Lock()
			missingConfigs = list
			missingConfigMu.Unlock()
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
	if historyPath == "" {
		historyPath = filepath.Join(os.Getenv("HOME"), ".config", "ceye", "run-history.json")
	}

	deps := providers.Dependencies{
		GitHubClient: newGitHubClient(),
		AzureClient:  newAzureClient(),
	}

	var providerInstances []core.Provider
	var providerNames []string
	var refreshers []func()
	providerStatus := make(map[string]string)
	providerTimes := make(map[string]time.Time)
	providerHealth := make(map[string]core.ProviderHealth)
	providerLastPoll := make(map[string]time.Time)
	providerLag := make(map[string]time.Duration)

	for _, candidate := range buildProviderEntries(cfg, providerStore) {
		provider, err := providers.CreateProvider(candidate.Config, deps)
		if err != nil {
			return fmt.Errorf("create provider: %w", err)
		}
		
		// Enable webhook mode for GitHub providers if webhooks are enabled
		if enableWebhooks {
			if ghProvider, ok := provider.(interface{ SetWebhookMode(bool) }); ok {
				ghProvider.SetWebhookMode(true)
				log.Printf("Enabled webhook mode for provider: %s", candidate.Alias)
			}
		}
		
		// Wrap with SafeProvider for panic recovery and validation
		safeProvider := providers.NewSafeProvider(provider)
		
		alias := candidate.Alias
		if refresher, ok := provider.(interface{ TriggerRefresh() }); ok {
			refreshers = append(refreshers, refresher.TriggerRefresh)
		}
		providerInstances = append(providerInstances, wrapProvider(safeProvider, alias))
		providerNames = append(providerNames, alias)
		providerStatus[alias] = ""
	}
	
	// Initialize storage for historical data
	var store *core.Store
	var storageBackend *storage.Storage
	storagePath := getStoragePath()
	if storagePath != "" {
		storageConfig := storage.Config{
			Path:            storagePath,
			RetentionDays:   90, // Keep 90 days of history
			CleanupInterval: 24 * time.Hour,
		}
		var err error
		storageBackend, err = storage.New(storageConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to initialize storage: %v\n", err)
			store = core.NewStore()
		} else {
			store = core.NewStoreWithStorage(storageBackend)
			defer storageBackend.Close()
		}
	} else {
		store = core.NewStore()
	}
	
	// Initialize alerting engine if configured
	var alertEngine *alerting.Engine
	if cfg != nil && cfg.Alerting != nil && cfg.Alerting.Enabled {
		if storageBackend == nil {
			fmt.Fprintf(os.Stderr, "warning: alerting requires storage, but storage initialization failed\n")
		} else {
			engine, err := config.BuildAlertEngine(cfg.Alerting, storageBackend)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to initialize alerting: %v\n", err)
			} else {
				alertEngine = engine
				alertEngine.SetStore(store) // Connect store for alert history
				fmt.Printf("alerting: loaded %d rules with %d channels\n", 
					len(cfg.Alerting.Rules), len(cfg.Alerting.Channels))
			}
		}
	}
	
	eventCh := make(chan core.RunEvent)
	
	// Start webhook server if enabled
	if enableWebhooks {
		webhookCfg := webhooks.Config{
			Port:         webhookPort,
			GitHubSecret: webhookSecret,
		}
		webhookServer := webhooks.NewServer(webhookCfg)
		
		// Start webhook server
		go func() {
			if err := webhookServer.Start(ctx); err != nil && ctx.Err() == nil {
				log.Printf("Webhook server error: %v", err)
			}
		}()
		
		// Merge webhook events into main event channel
		go func() {
			for event := range webhookServer.Events() {
				select {
				case eventCh <- event:
				case <-ctx.Done():
					return
				}
			}
		}()
		
		log.Printf("🪝 Webhook server enabled on port %d", webhookPort)
		log.Printf("   GitHub: http://localhost:%d/webhooks/github", webhookPort)
		log.Printf("   Azure:  http://localhost:%d/webhooks/azure", webhookPort)
		log.Printf("   Expose with: ngrok http %d", webhookPort)
	}
	
	// Fan out events to alerting engine if configured
	var storeEventCh chan core.RunEvent
	if alertEngine != nil {
		// Create separate channels for store and alerting
		storeEventCh = make(chan core.RunEvent, 100)
		alertEventCh := make(chan core.RunEvent, 100)
		
		// Start alerting engine
		go alertEngine.Start(ctx, alertEventCh)
		
		// Fan out events
		go func() {
			for event := range eventCh {
				// Send to store (TUI/Web UI)
				select {
				case storeEventCh <- event:
				case <-ctx.Done():
					return
				}
				
				// Send to alerting engine
				select {
				case alertEventCh <- event:
				case <-ctx.Done():
					return
				}
			}
			close(storeEventCh)
			close(alertEventCh)
		}()
	} else {
		// No alerting, events go directly to store
		storeEventCh = eventCh
	}

	log.Printf("Starting %d provider(s)", len(providerInstances))
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

	// If web mode, start HTTP server instead of TUI
	if web {
		return runWebServer(ctx, store, storageBackend, providerNames, providerStatus, providerHealth, storeEventCh, webPort, notify, webhookURL)
	}

	buildInfo := fmt.Sprintf("%s (%s)", Version, GitCommit)
	model := ui.NewModelWithBuildInfo(store, providerNames, refresh, openURL, copyToClipboard, buildInfo)
	
	// Add trend analyzer if storage is available
	if storageBackend != nil {
		trendAnalyzer := storage.NewTrendAnalyzer(storageBackend)
		model.SetTrendAnalyzer(trendAnalyzer)
	}
	
	var program *tea.Program
	snapshotMissing := func() []string {
		missingConfigMu.RLock()
		defer missingConfigMu.RUnlock()
		return copyMissing(missingConfigs)
	}

	notifyMissing := func() {
		if program == nil {
			return
		}
		program.Send(ui.RunUpdatedMsg{
			Timestamp:    time.Now(),
			MissingRepos: snapshotMissing(),
		})
	}

	model.SetMissingRepoAction(func() {
		if configRoot == "" {
			return
		}
		go func() {
			list, listErr := listMissingConfigs(configRoot)
			if listErr != nil {
				fmt.Fprintf(os.Stderr, "missing configs: %v\n", listErr)
				return
			}
			missingConfigMu.Lock()
			missingConfigs = list
			missingConfigMu.Unlock()
			notifyMissing()
		}()
	})
	model.SetProviderStoreAction(func(entry manager.ProviderRecord, action ui.ProviderStoreActionType) {
		go func() {
			switch action {
			case ui.ProviderStoreActionToggle:
				if err := providerStore.SetEnabled("ui", entry.ID, entry.Enabled); err != nil {
					fmt.Fprintf(os.Stderr, "provider store: %v\n", err)
					if program != nil {
						program.Send(ui.RunUpdatedMsg{
							Timestamp: time.Now(),
							Message:   fmt.Sprintf("store update failed: %v", err),
							Level:     "error",
							Store:     copyProviderRecords(providerStore.List()),
							Audit:     copyAudit(providerStore.Audit()),
						})
					}
					return
				}
				msg := fmt.Sprintf("%s %s", providers.DisplayName(entry.Config), map[bool]string{true: "enabled", false: "disabled"}[entry.Enabled])
				if program != nil {
					program.Send(ui.RunUpdatedMsg{
						Timestamp: time.Now(),
						Message:   msg,
						Level:     "info",
						Store:     copyProviderRecords(providerStore.List()),
						Audit:     copyAudit(providerStore.Audit()),
					})
				}
			case ui.ProviderStoreActionEdit:
				if err := providerStore.Update("ui", entry.ID, entry.Config); err != nil {
					fmt.Fprintf(os.Stderr, "provider store: %v\n", err)
					if program != nil {
						program.Send(ui.RunUpdatedMsg{
							Timestamp: time.Now(),
							Message:   fmt.Sprintf("store update failed: %v", err),
							Level:     "error",
							Store:     copyProviderRecords(providerStore.List()),
							Audit:     copyAudit(providerStore.Audit()),
						})
					}
					return
				}
				if program != nil {
					program.Send(ui.RunUpdatedMsg{
						Timestamp: time.Now(),
						Message:   fmt.Sprintf("%s updated", providers.DisplayName(entry.Config)),
						Level:     "info",
						Store:     copyProviderRecords(providerStore.List()),
						Audit:     copyAudit(providerStore.Audit()),
					})
				}
			case ui.ProviderStoreActionRemove:
				if err := providerStore.Remove("ui", entry.ID); err != nil {
					fmt.Fprintf(os.Stderr, "provider store: %v\n", err)
					if program != nil {
						program.Send(ui.RunUpdatedMsg{
							Timestamp: time.Now(),
							Message:   fmt.Sprintf("store removal failed: %v", err),
							Level:     "error",
							Store:     copyProviderRecords(providerStore.List()),
							Audit:     copyAudit(providerStore.Audit()),
						})
					}
					return
				}
				if program != nil {
					program.Send(ui.RunUpdatedMsg{
						Timestamp: time.Now(),
						Message:   fmt.Sprintf("%s removed", providers.DisplayName(entry.Config)),
						Level:     "info",
						Store:     copyProviderRecords(providerStore.List()),
						Audit:     copyAudit(providerStore.Audit()),
					})
				}
			case ui.ProviderStoreActionDuplicate:
				record, err := providerStore.Add("ui", entry.Config,
					manager.WithAuditAction("duplicate"),
					manager.WithAuditDetails(fmt.Sprintf("from %s", shortID(entry.ID))),
				)
				if err != nil {
					fmt.Fprintf(os.Stderr, "provider store duplicate: %v\n", err)
					if program != nil {
						program.Send(ui.RunUpdatedMsg{
							Timestamp: time.Now(),
							Message:   fmt.Sprintf("store duplication failed: %v", err),
							Level:     "error",
							Store:     copyProviderRecords(providerStore.List()),
							Audit:     copyAudit(providerStore.Audit()),
						})
					}
					return
				}
				if program != nil {
					program.Send(ui.RunUpdatedMsg{
						Timestamp: time.Now(),
						Message:   fmt.Sprintf("%s duplicated (%s)", providers.DisplayName(entry.Config), shortID(record.ID)),
						Level:     "info",
						Store:     copyProviderRecords(providerStore.List()),
						Audit:     copyAudit(providerStore.Audit()),
					})
				}
			}
		}()
	})
	program = tea.NewProgram(model)

	runErr := make(chan error, 1)
	go func() {
		_, err := program.Run()
		runErr <- err
	}()

	notifyMissing()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-storeEventCh:
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
						if len(event.Runs) > 0 {
							hist := providerHistory[event.Provider]
							for _, run := range event.Runs {
								summary := fmt.Sprintf("%s • %s • %s", run.WorkflowName, run.Status, run.Branch)
								if run.Conclusion != "" {
									summary = fmt.Sprintf("%s (%s)", summary, run.Conclusion)
								}
								hist = append([]string{summary}, hist...)
							}
							if len(hist) > 5 {
								hist = hist[:5]
							}
							providerHistory[event.Provider] = hist
						}
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
					if historyPath != "" {
						appendHistory(providerHistory, event.Provider, event.Runs, ts)
						if err := saveHistory(historyPath, providerHistory); err != nil {
							fmt.Fprintf(os.Stderr, "history: %v\n", err)
						}
					}
				} else if event.Message != "" {
					message = event.Message
					if event.Err != nil {
						level = "error"
					} else {
						level = "info"
					}
				}
				if notify && event.Err != nil {
					if err := sendNotification(event.Provider, event.Err.Error()); err != nil {
						fmt.Fprintf(os.Stderr, "notify: %v\n", err)
					}
				}
				if webhookURL != "" && event.Err != nil {
					if err := sendWebhook(ctx, webhookURL, event.Provider, event.Err.Error(), ts); err != nil {
						fmt.Fprintf(os.Stderr, "webhook: %v\n", err)
					}
				}
				// Check if this is a webhook event (provider name ends with "-webhook")
				isWebhook := strings.HasSuffix(event.Provider, "-webhook")
				var webhookRun *core.Run
				if isWebhook && len(event.Runs) > 0 {
					webhookRun = &event.Runs[0]
				}
				
				program.Send(ui.RunUpdatedMsg{
					Timestamp:    ts,
					Status:       copyStatus(providerStatus),
					Times:        copyTimes(providerTimes),
					Message:      message,
					Level:        level,
					Health:       copyHealth(providerHealth),
					Lag:          copyLag(providerLag),
					History:      copyHistory(providerHistory),
					Store:        copyProviderRecords(providerStore.List()),
					MissingRepos: snapshotMissing(),
					IsWebhook:    isWebhook,
					WebhookRun:   webhookRun,
				})
			}
		}
	}()

	select {
	case err := <-runErr:
		if err != nil {
			return fmt.Errorf("run UI: %w", err)
		}
	case <-ctx.Done():
		program.Quit()
		if err := <-runErr; err != nil {
			return fmt.Errorf("run UI: %w", err)
		}
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

func copyHistory(in map[string][]string) map[string][]string {
	out := make(map[string][]string, len(in))
	for k, v := range in {
		out[k] = append([]string(nil), v...)
	}
	return out
}

func copyMissing(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func copyProviderRecords(in []manager.ProviderRecord) []manager.ProviderRecord {
	if len(in) == 0 {
		return nil
	}
	out := make([]manager.ProviderRecord, len(in))
	copy(out, in)
	return out
}

func copyAudit(in []manager.StoreAuditEntry) []manager.StoreAuditEntry {
	if len(in) == 0 {
		return nil
	}
	out := make([]manager.StoreAuditEntry, len(in))
	copy(out, in)
	return out
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

func appendHistory(history map[string][]string, provider string, runs []core.Run, ts time.Time) {
	if len(runs) == 0 {
		return
	}
	for _, run := range runs {
		summary := fmt.Sprintf("%s • %s • %s", run.WorkflowName, run.Status, run.Branch)
		if run.Conclusion != "" {
			summary = fmt.Sprintf("%s (%s)", summary, run.Conclusion)
		}
		entry := fmt.Sprintf("%s @ %s", summary, ts.Format("15:04:05"))
		history[provider] = append([]string{entry}, history[provider]...)
		if len(history[provider]) > 20 {
			history[provider] = history[provider][:20]
		}
	}
}

func saveHistory(path string, history map[string][]string) error {
	if path == "" {
		return nil
	}
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}

func sendNotification(title, message string) error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`display notification %q with title %q`, message, title))
		return cmd.Start()
	case "linux":
		cmd := exec.Command("notify-send", title, message)
		return cmd.Start()
	default:
		// no-op on unsupported platforms
		return nil
	}
}

func sendWebhook(ctx context.Context, url, provider, message string, timestamp time.Time) error {
	if url == "" {
		return nil
	}
	event := map[string]interface{}{
		"provider":  provider,
		"message":   message,
		"timestamp": timestamp.Format(time.RFC3339),
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook %s returned %d: %s", url, resp.StatusCode, string(body))
	}
	return nil
}

func loadDistributedConfigs(cfgPath, configDir, githubOrg, azureOrg, azureProject string) (*config.Config, string, []string, error) {
	root, err := resolveConfigRoot(configDir)
	if err != nil {
		return nil, "", nil, err
	}
	if cfgPath != "" {
		cfg, err := config.Load(cfgPath)
		return cfg, root, nil, err
	}
	if def := defaultUserConfigPath(); def != "" {
		if _, err := os.Stat(def); err == nil {
			cfg, err := config.Load(def)
			return cfg, root, nil, err
		}
	}
	matches, err := config.DiscoverConfigs(root)
	if err != nil {
		return nil, root, nil, fmt.Errorf("discover configs: %w", err)
	}
	if len(matches) == 0 {
		autoCfg, err := autoConfigFromCLIs(githubOrg, azureOrg, azureProject)
		if err == nil {
			githubCount, azureCount := countAutoProviders(autoCfg)
			fmt.Fprintf(os.Stderr, "ceye: auto-discovered %d GitHub repos and %d Azure pipelines via gh/az CLI\n", githubCount, azureCount)
			return autoCfg, root, nil, nil
		}
		fmt.Fprintf(os.Stderr, "ceye: CLI discovery failed: %v\n", err)
		repos, _ := discoverGitRepos(root)
		if len(repos) > 0 {
			return &config.Config{}, root, repos, nil
		}
		fmt.Fprintf(os.Stderr, "ceye: no config files found under %s, falling back to demo provider\n", root)
		return &config.Config{Providers: []providers.ProviderConfig{{Type: "demo", Runs: 4}}}, root, nil, nil
	}
	var merged config.Config
	for _, match := range matches {
		cfg, err := config.Load(match)
		if err != nil {
			return nil, root, nil, fmt.Errorf("load config %s: %w", match, err)
		}
		merged.Providers = append(merged.Providers, cfg.Providers...)
	}
	return &merged, root, nil, nil
}

func autoConfigFromCLIs(githubOrg, azureOrg, azureProject string) (*config.Config, error) {
	githubOrg = resolveConfigValue(githubOrg, envGithubOrg, defaultGithubOrg)
	azureOrg = resolveConfigValue(azureOrg, envAzureOrg, defaultAzureOrg)
	azureProject = resolveConfigValue(azureProject, envAzureProject, defaultAzureProject)

	repos, err := listGithubRepos(githubOrg)
	if err != nil {
		return nil, fmt.Errorf("github discovery: %w", err)
	}
	pipelines, err := listAzurePipelines(azureOrg, azureProject)
	if err != nil {
		return nil, fmt.Errorf("azure discovery: %w", err)
	}

	return &config.Config{
		Providers: []providers.ProviderConfig{
			{Type: "github", Repos: repos},
			{Type: "azure", Org: azureOrg, Project: azureProject, Pipelines: pipelines},
		},
	}, nil
}

func resolveConfigValue(flagValue, envName, fallback string) string {
	if flagValue != "" {
		return flagValue
	}
	if env := os.Getenv(envName); env != "" {
		return env
	}
	return fallback
}

func listGithubRepos(org string) ([]githubprovider.RepoConfig, error) {
	cmd := exec.Command("gh", "repo", "list", org, "--limit", "1000", "--json", "nameWithOwner")
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var entries []struct {
		NameWithOwner string `json:"nameWithOwner"`
	}
	if err := json.Unmarshal(output, &entries); err != nil {
		return nil, fmt.Errorf("parse gh repo list: %w", err)
	}
	repos := make([]githubprovider.RepoConfig, 0, len(entries))
	seen := make(map[string]struct{})
	for _, entry := range entries {
		parts := strings.Split(entry.NameWithOwner, "/")
		if len(parts) != 2 {
			continue
		}
		if parts[0] != org {
			continue
		}
		repo := parts[1]
		if repo == "" {
			continue
		}
		if _, ok := seen[repo]; ok {
			continue
		}
		seen[repo] = struct{}{}
		repos = append(repos, githubprovider.RepoConfig{Owner: org, Repo: repo})
	}
	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories discovered for %s", org)
	}
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Repo < repos[j].Repo
	})
	return repos, nil
}

func listAzurePipelines(org, project string) ([]int, error) {
	orgURL := normalizeAzureOrgURL(org)
	cmd := exec.Command("az", "pipelines", "list", "--org", orgURL, "--project", project, "--query", "[].id", "-o", "tsv")
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	ids := make([]int, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		id, err := strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("parse pipeline id %q: %w", line, err)
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no pipelines discovered for %s/%s", orgURL, project)
	}
	sort.Ints(ids)
	return ids, nil
}

func normalizeAzureOrgURL(org string) string {
	org = strings.TrimSpace(org)
	org = strings.TrimSuffix(org, "/")
	if strings.HasPrefix(org, "http://") || strings.HasPrefix(org, "https://") {
		return org
	}
	return fmt.Sprintf("https://dev.azure.com/%s", org)
}

func countAutoProviders(cfg *config.Config) (int, int) {
	githubCount, azureCount := 0, 0
	for _, p := range cfg.Providers {
		switch p.Type {
		case "github":
			githubCount += len(p.Repos)
		case "azure":
			azureCount += len(p.Pipelines)
		}
	}
	return githubCount, azureCount
}

func resolveConfigRoot(configDir string) (string, error) {
	if override := os.Getenv(envConfigRoot); override != "" {
		return override, nil
	}
	if configDir != "" {
		return configDir, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return workspaceRootBase(cwd), nil
}

func workspaceRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return workspaceRootBase(cwd), nil
}

func workspaceRootBase(cwd string) string {
	root := cwd
	for {
		if filepath.Base(root) == "code" {
			return root
		}
		parent := filepath.Dir(root)
		if parent == root {
			return cwd
		}
		root = parent
	}
}

func discoverGitRepos(root string) ([]string, error) {
	var repos []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip directories we don't have permission to read
			if os.IsPermission(err) {
				return fs.SkipDir
			}
			return err
		}
		if !d.IsDir() {
			return nil
		}
		// Skip common directories that shouldn't be scanned
		name := d.Name()
		if name == ".Trash" || name == "node_modules" || name == ".git" && filepath.Base(filepath.Dir(path)) != "" {
			if name == ".git" {
				repos = append(repos, filepath.Dir(path))
			}
			return fs.SkipDir
		}
		return nil
	})
	return repos, err
}

func listMissingConfigs(root string) ([]string, error) {
	repos, err := discoverGitRepos(root)
	if err != nil {
		return nil, err
	}
	var missing []string
	for _, repo := range repos {
		if !repoHasConfig(repo) {
			missing = append(missing, repo)
		}
	}
	return missing, nil
}

func repoHasConfig(repo string) bool {
	for _, name := range []string{"ceye.yaml", "ceye.yml", "ceye.json", "ceye.toml"} {
		if _, err := os.Stat(filepath.Join(repo, name)); err == nil {
			return true
		}
	}
	return false
}

type providerEntry struct {
	Config providers.ProviderConfig
	Alias  string
}

func buildProviderEntries(cfg *config.Config, store *manager.Store) []providerEntry {
	used := make(map[string]struct{})
	entries := make([]providerEntry, 0, len(cfg.Providers)+len(store.EnabledRecords()))
	for idx, pCfg := range cfg.Providers {
		fallback := fmt.Sprintf("%s-%d", strings.ToLower(pCfg.Type), idx+1)
		alias := uniqueAlias(pCfg.DisplayName, fallback, used)
		entries = append(entries, providerEntry{Config: pCfg, Alias: alias})
	}
	for _, record := range store.EnabledRecords() {
		shortID := record.ID
		if len(shortID) > 6 {
			shortID = shortID[:6]
		}
		fallback := fmt.Sprintf("%s-%s", strings.ToLower(record.Config.Type), shortID)
		alias := uniqueAlias(record.Config.DisplayName, fallback, used)
		entries = append(entries, providerEntry{Config: record.Config, Alias: alias})
	}
	return entries
}

func uniqueAlias(preferred, fallback string, used map[string]struct{}) string {
	alias := strings.TrimSpace(preferred)
	if alias == "" {
		alias = fallback
	}
	if alias == "" {
		alias = "provider"
	}
	base := alias
	counter := 1
	for {
		if _, ok := used[alias]; !ok {
			used[alias] = struct{}{}
			return alias
		}
		alias = fmt.Sprintf("%s-%d", base, counter)
		counter++
	}
}

func wrapProvider(p core.Provider, alias string) core.Provider {
	if alias == "" {
		return p
	}
	return &namedProvider{Provider: p, alias: alias}
}

type namedProvider struct {
	core.Provider
	alias string
}

func (n *namedProvider) Name() string {
	if n.alias != "" {
		return n.alias
	}
	return n.Provider.Name()
}

func resolveProviderStorePath(flag string) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv("CEYE_PROVIDER_STORE"); env != "" {
		return env
	}
	return defaultProviderStorePath()
}

func defaultUserConfigPath() string {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "ceye", "ceye.yaml")
	}
	return ""
}

func newGitHubClient() githubprovider.GitHubClient {
	token := githubToken()
	if token != "" {
		return githubprovider.NewHTTPClient(token)
	}
	path, err := exec.LookPath("gh")
	fmt.Fprintf(os.Stderr, "ceye: GH lookup -> path=%q err=%v\n", path, err)
	if err == nil {
		fmt.Fprintf(os.Stderr, "ceye: using gh CLI-based GitHub client (%s)\n", path)
		return githubprovider.NewCLIClient()
	}
	fmt.Fprintln(os.Stderr, "ceye: no GitHub token found, falling back to unauthenticated REST client (requests may be rate-limited)")
	return githubprovider.NewHTTPClient("")
}

func newAzureClient() azureprovider.AzureClient {
	pat := azureToken()
	// NewClient requires org, but we'll get that from config later
	// For now, just return a client with PAT for the factory to use
	return azureprovider.NewClient("", pat)
}

func defaultProviderStorePath() string {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "ceye", "providers.json")
	}
	return "providers.json"
}

func getStoragePath() string {
	// Check environment variable first
	if path := os.Getenv("CEYE_STORAGE_PATH"); path != "" {
		return path
	}
	
	// Use default path in user config directory
	if home := os.Getenv("HOME"); home != "" {
		configDir := filepath.Join(home, ".config", "ceye")
		// Ensure directory exists
		if err := os.MkdirAll(configDir, 0755); err == nil {
			return filepath.Join(configDir, "runs.db")
		}
	}
	
	// Fallback to current directory
	return "runs.db"
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

func runWebServer(ctx context.Context, store *core.Store, storageBackend *storage.Storage, providerNames []string, providerStatus map[string]string, providerHealth map[string]core.ProviderHealth, eventCh chan core.RunEvent, port int, notify bool, webhookURL string) error {
	srv := server.New(store, providerNames, port)
	
	// Set version information
	srv.SetVersion(Version, GitCommit, BuildTime)
	
	// Add trend analyzer if storage backend is available
	if storageBackend != nil {
		trendAnalyzer := storage.NewTrendAnalyzer(storageBackend)
		srv.SetTrendAnalyzer(trendAnalyzer)
	}
	
	// Start web server in background
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Start(ctx)
	}()
	
	// Check if server started successfully
	select {
	case err := <-serverErr:
		// Server failed immediately (port in use, etc)
		return fmt.Errorf("web server failed to start: %w", err)
	case <-time.After(200 * time.Millisecond):
		// Server started successfully
		log.Printf("✓ Web server ready at http://localhost:%d", port)
	}
	
	// Open browser
	time.Sleep(400 * time.Millisecond)
	openURL(fmt.Sprintf("http://localhost:%d", port))
	
	// Process events and broadcast to connected clients
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
						health := providerHealth[event.Provider]
						health.ErrorCount++
						health.LastError = event.Timestamp
						providerHealth[event.Provider] = health
						
						if notify {
							sendNotification(event.Provider, event.Err.Error())
						}
						if webhookURL != "" {
							sendWebhook(ctx, webhookURL, event.Provider, event.Err.Error(), event.Timestamp)
						}
					} else {
						providerStatus[event.Provider] = ""
						health := providerHealth[event.Provider]
						health.LastSuccess = event.Timestamp
						health.ErrorCount = 0
						providerHealth[event.Provider] = health
					}
				}
				
				srv.UpdateStatus(providerStatus, providerHealth)
				srv.BroadcastUpdate()
			}
		}
	}()
	
	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		return nil
	}
}
