package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/joelklabo/ceye/internal/providers/github"
	"github.com/joelklabo/ceye/internal/validation"
)

func validateCmd() *cobra.Command {
	var duration time.Duration
	var repos []string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Run webhook vs polling validation test",
		Long: `Runs dual-mode validation to compare webhook and polling performance.

This command starts TWO providers simultaneously:
1. Webhook Mode - Normal operation (waits for webhooks)
2. Polling Mode - Continuous polling (for validation)

Every 30 seconds, it compares the two stores and logs:
- Discrepancies (missing runs in either store)
- Timing metrics (how much faster webhooks are)
- Match rate (percentage of runs that appear in both)

Results are logged to tmp/validation-metrics.jsonl

Example:
  ceye validate --duration 10m
  ceye validate --duration 1h --repos owner/repo1,owner/repo2
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidation(duration, repos)
		},
	}

	cmd.Flags().DurationVar(&duration, "duration", 10*time.Minute, "How long to run validation")
	cmd.Flags().StringSliceVar(&repos, "repos", nil, "Repositories to monitor (format: owner/repo)")

	return cmd
}

func runValidation(duration time.Duration, repoStrings []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Set timeout
	ctx, timeoutCancel := context.WithTimeout(ctx, duration)
	defer timeoutCancel()

	// Parse repos
	var repos []github.RepoConfig
	if len(repoStrings) == 0 {
		// Load from default config
		cfg, _, _, err := loadDistributedConfigs("", ".", "", "", "")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Extract GitHub repos from config
		for _, provider := range cfg.Providers {
			if provider.Type == "github" {
				for _, repo := range provider.Repos {
					repos = append(repos, github.RepoConfig{
						Owner: repo.Owner,
						Repo:  repo.Repo,
					})
				}
			}
		}

		if len(repos) == 0 {
			return fmt.Errorf("no GitHub repositories found in config")
		}
	} else {
		// Parse from command line
		for _, repoStr := range repoStrings {
			// Parse "owner/repo" format
			parts := strings.Split(repoStr, "/")
			if len(parts) != 2 {
				return fmt.Errorf("invalid repo format: %s (expected owner/repo)", repoStr)
			}
			repos = append(repos, github.RepoConfig{
				Owner: parts[0],
				Repo:  parts[1],
			})
		}
	}

	fmt.Fprintf(os.Stderr, "🔍 Webhook Validation Mode\n")
	fmt.Fprintf(os.Stderr, "   Monitoring: %d repos\n", len(repos))
	fmt.Fprintf(os.Stderr, "   Duration: %v\n", duration)
	fmt.Fprintf(os.Stderr, "   Metrics log: tmp/validation-metrics.jsonl\n")
	fmt.Fprintf(os.Stderr, "\n")

	// Create validation harness
	harness, err := validation.NewHarness(repos)
	if err != nil {
		return fmt.Errorf("failed to create validation harness: %w", err)
	}

	// Start periodic status reporting
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		startTime := time.Now()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				elapsed := time.Since(startTime)
				metrics := harness.GetMetrics()
				discrepancies := harness.GetDiscrepancies()

				fmt.Fprintf(os.Stderr, "\n⏱️  %s elapsed\n", formatDuration(elapsed))
				fmt.Fprintf(os.Stderr, "📊 Metrics:\n")
				fmt.Fprintf(os.Stderr, "   Webhook runs: %d\n", metrics.WebhookRunCount)
				fmt.Fprintf(os.Stderr, "   Polling runs: %d\n", metrics.PollingRunCount)
				fmt.Fprintf(os.Stderr, "   Matched: %d", metrics.MatchedRuns)

				if metrics.WebhookRunCount > 0 || metrics.PollingRunCount > 0 {
					matchRate := float64(metrics.MatchedRuns) / float64(max(metrics.WebhookRunCount, metrics.PollingRunCount)) * 100
					fmt.Fprintf(os.Stderr, " (%.1f%%)", matchRate)
				}
				fmt.Fprintf(os.Stderr, "\n")

				if metrics.WebhookAdvantage > 0 {
					fmt.Fprintf(os.Stderr, "   Webhook advantage: %v ✅\n", metrics.WebhookAdvantage)
				}

				if len(discrepancies) > 0 {
					fmt.Fprintf(os.Stderr, "\n⚠️  Discrepancies: %d\n", len(discrepancies))
					// Show last 3 discrepancies
					start := max(0, len(discrepancies)-3)
					for i := start; i < len(discrepancies); i++ {
						d := discrepancies[i]
						fmt.Fprintf(os.Stderr, "   %s - RunID: %s\n", d.Type, d.RunID)
					}
				} else {
					fmt.Fprintf(os.Stderr, "   ✅ No discrepancies\n")
				}
			}
		}
	}()

	// Start validation
	log.Printf("Starting validation harness...")
	err = harness.Start(ctx)

	// Print final summary
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "Validation Complete\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	finalMetrics := harness.GetMetrics()
	finalDiscrepancies := harness.GetDiscrepancies()

	fmt.Fprintf(os.Stderr, "\n📊 Final Metrics:\n")
	fmt.Fprintf(os.Stderr, "   Total webhook runs: %d\n", finalMetrics.WebhookRunCount)
	fmt.Fprintf(os.Stderr, "   Total polling runs: %d\n", finalMetrics.PollingRunCount)
	fmt.Fprintf(os.Stderr, "   Matched runs: %d\n", finalMetrics.MatchedRuns)
	fmt.Fprintf(os.Stderr, "   Missing in webhook: %d\n", finalMetrics.MissingInWebhook)
	fmt.Fprintf(os.Stderr, "   Missing in polling: %d\n", finalMetrics.MissingInPolling)

	if finalMetrics.WebhookAdvantage > 0 {
		fmt.Fprintf(os.Stderr, "   Average webhook advantage: %v\n", finalMetrics.WebhookAdvantage)
	}

	fmt.Fprintf(os.Stderr, "\n⚠️  Total discrepancies: %d\n", len(finalDiscrepancies))

	if len(finalDiscrepancies) == 0 {
		fmt.Fprintf(os.Stderr, "\n✅ SUCCESS: 100%% match rate, no discrepancies\n")
	} else {
		fmt.Fprintf(os.Stderr, "\n⚠️  WARNING: Discrepancies detected\n")
		fmt.Fprintf(os.Stderr, "   Check tmp/validation-metrics.jsonl for details\n")
	}

	fmt.Fprintf(os.Stderr, "\n")

	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		return err
	}

	return nil
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
