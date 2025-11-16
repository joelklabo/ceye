package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
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
	"github.com/joelklabo/ceye/internal/providers/manager"
	"github.com/joelklabo/ceye/internal/ui"
)

const envConfigRoot = "CEYE_CONFIG_ROOT"

var providerStoreFlag string
var configDirFlag string

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
	rootCmd := &cobra.Command{
		Use:   "ci-dash",
		Short: "CI Status Dashboard TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ctx, cfgPath, configDirFlag, demo, demoRuns, demoDuration, eventLogPath, notify, historyPath, webhookURL, resolveProviderStorePath(providerStoreFlag))
		},
	}
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

	rootCmd.AddCommand(providerCmd())

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func run(parentCtx context.Context, cfgPath, configDir string, demo bool, demoRuns int, demoDuration time.Duration, eventLogPath string, notify bool, historyPath string, webhookURL string, providerStorePath string) error {
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
	providerHistory := make(map[string][]string)
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

		cfg, missingConfigs, err = loadDistributedConfigs(cfgPath, configDir)
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

	for _, candidate := range buildProviderEntries(cfg, providerStore) {
		provider, err := providers.CreateProvider(candidate.Config, deps)
		if err != nil {
			return fmt.Errorf("create provider: %w", err)
		}
		alias := candidate.Alias
		if refresher, ok := provider.(interface{ TriggerRefresh() }); ok {
			refreshers = append(refreshers, refresher.TriggerRefresh)
		}
		providerInstances = append(providerInstances, wrapProvider(provider, alias))
		providerNames = append(providerNames, alias)
		providerStatus[alias] = ""
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
	var program *tea.Program
	model.SetProviderStoreAction(func(entry manager.ProviderRecord, action ui.ProviderStoreActionType) {
		go func() {
			switch action {
			case ui.ProviderStoreActionToggle:
				if err := providerStore.SetEnabled(entry.ID, entry.Enabled); err != nil {
					fmt.Fprintf(os.Stderr, "provider store: %v\n", err)
					if program != nil {
						program.Send(ui.RunUpdatedMsg{
							Timestamp: time.Now(),
							Message:   fmt.Sprintf("store update failed: %v", err),
							Level:     "error",
							Store:     copyProviderRecords(providerStore.List()),
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
					})
				}
			case ui.ProviderStoreActionEdit:
				if err := providerStore.Update(entry.ID, entry.Config); err != nil {
					fmt.Fprintf(os.Stderr, "provider store: %v\n", err)
					if program != nil {
						program.Send(ui.RunUpdatedMsg{
							Timestamp: time.Now(),
							Message:   fmt.Sprintf("store update failed: %v", err),
							Level:     "error",
							Store:     copyProviderRecords(providerStore.List()),
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
					})
				}
			case ui.ProviderStoreActionRemove:
				if err := providerStore.Remove(entry.ID); err != nil {
					fmt.Fprintf(os.Stderr, "provider store: %v\n", err)
					if program != nil {
						program.Send(ui.RunUpdatedMsg{
							Timestamp: time.Now(),
							Message:   fmt.Sprintf("store removal failed: %v", err),
							Level:     "error",
							Store:     copyProviderRecords(providerStore.List()),
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
					})
				}
			case ui.ProviderStoreActionDuplicate:
				record, err := providerStore.Add(entry.Config)
				if err != nil {
					fmt.Fprintf(os.Stderr, "provider store duplicate: %v\n", err)
					if program != nil {
						program.Send(ui.RunUpdatedMsg{
							Timestamp: time.Now(),
							Message:   fmt.Sprintf("store duplication failed: %v", err),
							Level:     "error",
							Store:     copyProviderRecords(providerStore.List()),
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
					})
				}
			}
		}()
	})
	program = tea.NewProgram(model)

	// if we discovered repos but no configs, alert the user immediately
	if len(missingConfigs) > 0 {
		if program != nil {
			program.Send(ui.RunUpdatedMsg{
				Timestamp:    time.Now(),
				Message:      fmt.Sprintf("missing configs for: %s", strings.Join(missingConfigs, ", ")),
				Level:        "warn",
				Store:        copyProviderRecords(providerStore.List()),
				MissingRepos: copyMissing(missingConfigs),
			})
		}
	}

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
					MissingRepos: copyMissing(missingConfigs),
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

func loadDistributedConfigs(cfgPath, configDir string) (*config.Config, []string, error) {
	if cfgPath != "" {
		cfg, err := config.Load(cfgPath)
		return cfg, nil, err
	}
	root, err := resolveConfigRoot(configDir)
	if err != nil {
		return nil, nil, err
	}
	configDir = root
	matches, err := config.DiscoverConfigs(configDir)
	if err != nil {
		return nil, nil, fmt.Errorf("discover configs: %w", err)
	}
	if len(matches) == 0 {
		repos, repoErr := discoverGitRepos(configDir)
		msg := fmt.Sprintf("no config files found under %s", configDir)
		if repoErr == nil && len(repos) > 0 {
			msg = fmt.Sprintf("%s (found git repos: %s)", msg, strings.Join(repos, ", "))
		}
		if len(repos) > 0 {
			return &config.Config{}, repos, nil
		}
		return nil, nil, fmt.Errorf("%s", msg)
	}
	var merged config.Config
	for _, match := range matches {
		cfg, err := config.Load(match)
		if err != nil {
			return nil, nil, fmt.Errorf("load config %s: %w", match, err)
		}
		merged.Providers = append(merged.Providers, cfg.Providers...)
	}
	return &merged, nil, nil
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
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if d.Name() == ".git" {
			repos = append(repos, filepath.Dir(path))
			return fs.SkipDir
		}
		return nil
	})
	return repos, err
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

func defaultProviderStorePath() string {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "ceye", "providers.json")
	}
	return "providers.json"
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
