package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/joelklabo/ceye/internal/storage"
)

// TrendsPanel renders trend information for a provider
func renderTrendsPanel(analyzer *storage.TrendAnalyzer, provider string, width int) string {
	if analyzer == nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Get trends for the last 7 days
	trends, err := analyzer.GetAllTrends(ctx, provider, 7*24*time.Hour)
	if err != nil {
		return renderTrendsPanelError(width)
	}

	return renderTrendsPanelContent(trends, width)
}

func renderTrendsPanelError(width int) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width - 2)

	content := "Trends (Last 7 Days)\n\n" +
		"No trend data available"

	return style.Render(content)
}

func renderTrendsPanelContent(trends map[string]*storage.Trend, width int) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Width(width - 2)

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Trends (Last 7 Days)"))
	lines = append(lines, "")

	// Success Rate
	if trend, ok := trends["success_rate"]; ok {
		lines = append(lines, formatTrendLine("Success Rate:", trend.Current*100, "%", trend.Change, trend.Direction))
	}

	// Duration
	if trend, ok := trends["duration"]; ok {
		durationStr := formatTrendDuration(time.Duration(trend.Current * float64(time.Second)))
		lines = append(lines, formatTrendLine("Avg Duration:", durationStr, "", trend.Change, trend.Direction))
	}

	// Frequency
	if trend, ok := trends["frequency"]; ok {
		lines = append(lines, formatTrendLine("Builds/Day:", trend.Current, "", trend.Change, trend.Direction))
	}

	// Failure Rate
	if trend, ok := trends["failure_rate"]; ok {
		lines = append(lines, formatTrendLine("Failure Rate:", trend.Current*100, "%", trend.Change, trend.Direction))
	}

	content := ""
	for _, line := range lines {
		content += line + "\n"
	}

	return style.Render(content)
}

func formatTrendLine(label string, value interface{}, unit string, change float64, direction storage.TrendDirection) string {
	// Format the value
	var valueStr string
	switch v := value.(type) {
	case float64:
		if v == 0 {
			valueStr = "N/A"
		} else {
			valueStr = fmt.Sprintf("%.1f%s", v, unit)
		}
	case string:
		valueStr = v
	default:
		valueStr = fmt.Sprintf("%v%s", value, unit)
	}

	// Format the change indicator
	var indicator string
	var changeColor lipgloss.Color
	if change == 0 || direction == storage.TrendStable {
		indicator = "→"
		changeColor = lipgloss.Color("240")
	} else if direction == storage.TrendUp {
		indicator = "↑"
		changeColor = lipgloss.Color("42") // Green
	} else {
		indicator = "↓"
		changeColor = lipgloss.Color("196") // Red
	}

	changeStr := ""
	if change != 0 {
		changeStr = fmt.Sprintf(" (%s %.1f%%)", indicator, abs(change))
	} else {
		changeStr = fmt.Sprintf(" (%s stable)", indicator)
	}

	labelStyle := lipgloss.NewStyle().Width(16).Align(lipgloss.Left)
	valueStyle := lipgloss.NewStyle().Width(12).Align(lipgloss.Right).Bold(true)
	changeStyle := lipgloss.NewStyle().Foreground(changeColor)

	return labelStyle.Render(label) + " " + 
		valueStyle.Render(valueStr) + " " +
		changeStyle.Render(changeStr)
}

func formatTrendDuration(d time.Duration) string {
	if d == 0 {
		return "N/A"
	}

	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}

	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	if minutes < 60 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}

	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
