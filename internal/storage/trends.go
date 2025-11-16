package storage

import (
	"context"
	"fmt"
	"time"
)

// TrendDirection indicates whether a metric is improving, declining, or stable
type TrendDirection int

const (
	TrendUnknown TrendDirection = iota
	TrendUp
	TrendDown
	TrendStable
)

func (d TrendDirection) String() string {
	switch d {
	case TrendUp:
		return "up"
	case TrendDown:
		return "down"
	case TrendStable:
		return "stable"
	default:
		return "unknown"
	}
}

// Trend represents a metric trend over time
type Trend struct {
	Metric    string         // "success_rate", "duration", "frequency"
	Period    time.Duration  // Time period analyzed
	Current   float64        // Current value
	Previous  float64        // Previous period value
	Change    float64        // Percentage change
	Direction TrendDirection // Up, Down, Stable
}

// TrendAnalyzer calculates trends from historical data
type TrendAnalyzer struct {
	storage *Storage
}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer(storage *Storage) *TrendAnalyzer {
	return &TrendAnalyzer{storage: storage}
}

// GetSuccessRateTrend calculates success rate trend over a period
func (ta *TrendAnalyzer) GetSuccessRateTrend(ctx context.Context, provider string, period time.Duration) (*Trend, error) {
	now := time.Now()
	
	// Get current period metrics
	currentStart := now.Add(-period)
	currentMetrics, err := ta.storage.GetProviderMetricsRange(ctx, provider, currentStart, now)
	if err != nil {
		return nil, fmt.Errorf("get current metrics: %w", err)
	}

	// Get previous period metrics (same duration, shifted back)
	previousStart := currentStart.Add(-period)
	previousMetrics, err := ta.storage.GetProviderMetricsRange(ctx, provider, previousStart, currentStart)
	if err != nil {
		return nil, fmt.Errorf("get previous metrics: %w", err)
	}

	current := currentMetrics.SuccessRate
	previous := previousMetrics.SuccessRate
	change := 0.0
	if previous > 0 {
		change = ((current - previous) / previous) * 100
	}

	direction := TrendStable
	if change > 1 {
		direction = TrendUp
	} else if change < -1 {
		direction = TrendDown
	}

	return &Trend{
		Metric:    "success_rate",
		Period:    period,
		Current:   current,
		Previous:  previous,
		Change:    change,
		Direction: direction,
	}, nil
}

// GetDurationTrend calculates average duration trend over a period
func (ta *TrendAnalyzer) GetDurationTrend(ctx context.Context, provider string, period time.Duration) (*Trend, error) {
	now := time.Now()
	
	// Get current period metrics
	currentStart := now.Add(-period)
	currentMetrics, err := ta.storage.GetProviderMetricsRange(ctx, provider, currentStart, now)
	if err != nil {
		return nil, fmt.Errorf("get current metrics: %w", err)
	}

	// Get previous period metrics
	previousStart := currentStart.Add(-period)
	previousMetrics, err := ta.storage.GetProviderMetricsRange(ctx, provider, previousStart, currentStart)
	if err != nil {
		return nil, fmt.Errorf("get previous metrics: %w", err)
	}

	// Convert to seconds for easier interpretation
	current := currentMetrics.AvgDuration.Seconds()
	previous := previousMetrics.AvgDuration.Seconds()
	change := 0.0
	if previous > 0 {
		change = ((current - previous) / previous) * 100
	}

	// For duration, down is good (faster builds)
	direction := TrendStable
	if change < -5 {
		direction = TrendUp // Improvement (faster)
	} else if change > 5 {
		direction = TrendDown // Degradation (slower)
	}

	return &Trend{
		Metric:    "duration",
		Period:    period,
		Current:   current,
		Previous:  previous,
		Change:    change,
		Direction: direction,
	}, nil
}

// GetFrequencyTrend calculates build frequency trend over a period
func (ta *TrendAnalyzer) GetFrequencyTrend(ctx context.Context, provider string, period time.Duration) (*Trend, error) {
	now := time.Now()
	
	// Get current period count
	currentStart := now.Add(-period)
	currentMetrics, err := ta.storage.GetProviderMetricsRange(ctx, provider, currentStart, now)
	if err != nil {
		return nil, fmt.Errorf("get current metrics: %w", err)
	}

	// Get previous period count
	previousStart := currentStart.Add(-period)
	previousMetrics, err := ta.storage.GetProviderMetricsRange(ctx, provider, previousStart, currentStart)
	if err != nil {
		return nil, fmt.Errorf("get previous metrics: %w", err)
	}

	// Adjust for time periods to get rate (builds per day)
	days := period.Hours() / 24
	current := float64(currentMetrics.Total) / days
	previous := float64(previousMetrics.Total) / days
	
	change := 0.0
	if previous > 0 {
		change = ((current - previous) / previous) * 100
	}

	direction := TrendStable
	if change > 10 {
		direction = TrendUp
	} else if change < -10 {
		direction = TrendDown
	}

	return &Trend{
		Metric:    "frequency",
		Period:    period,
		Current:   current,
		Previous:  previous,
		Change:    change,
		Direction: direction,
	}, nil
}

// GetFailureRateTrend calculates failure rate trend over a period
func (ta *TrendAnalyzer) GetFailureRateTrend(ctx context.Context, provider string, period time.Duration) (*Trend, error) {
	now := time.Now()
	
	// Get current period metrics
	currentStart := now.Add(-period)
	currentMetrics, err := ta.storage.GetProviderMetricsRange(ctx, provider, currentStart, now)
	if err != nil {
		return nil, fmt.Errorf("get current metrics: %w", err)
	}

	// Get previous period metrics
	previousStart := currentStart.Add(-period)
	previousMetrics, err := ta.storage.GetProviderMetricsRange(ctx, provider, previousStart, currentStart)
	if err != nil {
		return nil, fmt.Errorf("get previous metrics: %w", err)
	}

	current := 0.0
	if currentMetrics.Total > 0 {
		current = float64(currentMetrics.Failed) / float64(currentMetrics.Total)
	}

	previous := 0.0
	if previousMetrics.Total > 0 {
		previous = float64(previousMetrics.Failed) / float64(previousMetrics.Total)
	}

	change := 0.0
	if previous > 0 {
		change = ((current - previous) / previous) * 100
	}

	// For failure rate, down is good
	direction := TrendStable
	if change < -10 {
		direction = TrendUp // Improvement (fewer failures)
	} else if change > 10 {
		direction = TrendDown // Degradation (more failures)
	}

	return &Trend{
		Metric:    "failure_rate",
		Period:    period,
		Current:   current,
		Previous:  previous,
		Change:    change,
		Direction: direction,
	}, nil
}

// GetAllTrends returns all trends for a provider
func (ta *TrendAnalyzer) GetAllTrends(ctx context.Context, provider string, period time.Duration) (map[string]*Trend, error) {
	trends := make(map[string]*Trend)

	successTrend, err := ta.GetSuccessRateTrend(ctx, provider, period)
	if err != nil {
		return nil, fmt.Errorf("get success rate trend: %w", err)
	}
	trends["success_rate"] = successTrend

	durationTrend, err := ta.GetDurationTrend(ctx, provider, period)
	if err != nil {
		return nil, fmt.Errorf("get duration trend: %w", err)
	}
	trends["duration"] = durationTrend

	frequencyTrend, err := ta.GetFrequencyTrend(ctx, provider, period)
	if err != nil {
		return nil, fmt.Errorf("get frequency trend: %w", err)
	}
	trends["frequency"] = frequencyTrend

	failureTrend, err := ta.GetFailureRateTrend(ctx, provider, period)
	if err != nil {
		return nil, fmt.Errorf("get failure rate trend: %w", err)
	}
	trends["failure_rate"] = failureTrend

	return trends, nil
}

// TimeSeriesPoint represents a single data point in a time series
type TimeSeriesPoint struct {
	Timestamp time.Time
	Value     float64
}

// GetSuccessRateTimeSeries returns success rate over time with specified granularity
func (ta *TrendAnalyzer) GetSuccessRateTimeSeries(ctx context.Context, provider string, start, end time.Time, bucketSize time.Duration) ([]TimeSeriesPoint, error) {
	var points []TimeSeriesPoint

	current := start
	for current.Before(end) {
		next := current.Add(bucketSize)
		if next.After(end) {
			next = end
		}

		metrics, err := ta.storage.GetProviderMetrics(ctx, provider, current)
		if err != nil {
			return nil, fmt.Errorf("get metrics for bucket: %w", err)
		}

		points = append(points, TimeSeriesPoint{
			Timestamp: current,
			Value:     metrics.SuccessRate,
		})

		current = next
	}

	return points, nil
}

// GetDurationTimeSeries returns average duration over time with specified granularity
func (ta *TrendAnalyzer) GetDurationTimeSeries(ctx context.Context, provider string, start, end time.Time, bucketSize time.Duration) ([]TimeSeriesPoint, error) {
	var points []TimeSeriesPoint

	current := start
	for current.Before(end) {
		next := current.Add(bucketSize)
		if next.After(end) {
			next = end
		}

		metrics, err := ta.storage.GetProviderMetrics(ctx, provider, current)
		if err != nil {
			return nil, fmt.Errorf("get metrics for bucket: %w", err)
		}

		points = append(points, TimeSeriesPoint{
			Timestamp: current,
			Value:     metrics.AvgDuration.Seconds(),
		})

		current = next
	}

	return points, nil
}

// DetectAnomalies identifies unusual patterns in run data
func (ta *TrendAnalyzer) DetectAnomalies(ctx context.Context, provider string, period time.Duration) ([]Anomaly, error) {
	var anomalies []Anomaly

	// Get baseline metrics from the period
	start := time.Now().Add(-period)
	metrics, err := ta.storage.GetProviderMetrics(ctx, provider, start)
	if err != nil {
		return nil, fmt.Errorf("get metrics: %w", err)
	}

	// Check for anomalous failure rate (> 20%)
	failureRate := 0.0
	if metrics.Total > 0 {
		failureRate = float64(metrics.Failed) / float64(metrics.Total)
	}
	if failureRate > 0.2 {
		anomalies = append(anomalies, Anomaly{
			Type:        "high_failure_rate",
			Severity:    "warning",
			Description: fmt.Sprintf("High failure rate: %.1f%%", failureRate*100),
			Value:       failureRate,
			Threshold:   0.2,
		})
	}

	// Check for very long duration (> 2x average for recent period)
	if metrics.MaxDuration > metrics.AvgDuration*2 && metrics.AvgDuration > 0 {
		anomalies = append(anomalies, Anomaly{
			Type:        "duration_spike",
			Severity:    "info",
			Description: fmt.Sprintf("Build duration spike detected: %v (avg: %v)", metrics.MaxDuration, metrics.AvgDuration),
			Value:       metrics.MaxDuration.Seconds(),
			Threshold:   (metrics.AvgDuration * 2).Seconds(),
		})
	}

	return anomalies, nil
}

// Anomaly represents an unusual pattern detected in the data
type Anomaly struct {
	Type        string  // "high_failure_rate", "duration_spike", "frequency_drop"
	Severity    string  // "info", "warning", "critical"
	Description string
	Value       float64
	Threshold   float64
}
