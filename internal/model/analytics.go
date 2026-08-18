package model

import (
	"fmt"
	"sort"
	"time"
)

// TrendAnalysis computes trend direction for sensor readings.
type TrendAnalysis struct {
	TankID    int64
	SensorType string
	Trend     string // rising, falling, stable
	Slope     float64
	Readings  int
}

// AnalyzeTrend computes a simple linear regression slope on readings.
func AnalyzeTrend(readings []SensorReading) *TrendAnalysis {
	if len(readings) < 2 {
		return &TrendAnalysis{Trend: "insufficient"}
	}

	// Sort by timestamp ascending
	sorted := make([]SensorReading, len(readings))
	copy(sorted, readings)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	n := float64(len(sorted))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	for i, r := range sorted {
		x := float64(i)
		sumX += x
		sumY += r.Value
		sumXY += x * r.Value
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return &TrendAnalysis{
			TankID:     sorted[0].TankID,
			SensorType: sorted[0].Type,
			Trend:      "stable",
			Slope:      0,
			Readings:   len(sorted),
		}
	}

	slope := (n*sumXY - sumX*sumY) / denominator

	trend := "stable"
	if slope > 0.1 {
		trend = "rising"
	} else if slope < -0.1 {
		trend = "falling"
	}

	return &TrendAnalysis{
		TankID:     sorted[0].TankID,
		SensorType: sorted[0].Type,
		Trend:      trend,
		Slope:      slope,
		Readings:   len(sorted),
	}
}

// ComputeMovingAverage returns the moving average of the last N values.
func ComputeMovingAverage(readings []SensorReading, window int) []float64 {
	if window <= 0 || len(readings) == 0 {
		return nil
	}

	sorted := make([]SensorReading, len(readings))
	copy(sorted, readings)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	result := make([]float64, 0, len(sorted))
	for i := 0; i < len(sorted); i++ {
		start := i - window + 1
		if start < 0 {
			start = 0
		}
		sum := 0.0
		count := 0
		for j := start; j <= i; j++ {
			sum += sorted[j].Value
			count++
		}
		result = append(result, sum/float64(count))
	}
	return result
}

// DetectAnomalies returns readings that deviate more than threshold from the mean.
func DetectAnomalies(readings []SensorReading, threshold float64) []SensorReading {
	if len(readings) < 3 || threshold <= 0 {
		return nil
	}

	mean := 0.0
	for _, r := range readings {
		mean += r.Value
	}
	mean /= float64(len(readings))

	var anomalies []SensorReading
	for _, r := range readings {
		deviation := r.Value - mean
		if deviation < 0 {
			deviation = -deviation
		}
		if deviation > threshold {
			anomalies = append(anomalies, r)
		}
	}
	return anomalies
}

// EquipmentHealthScore computes a health score 0-100 for equipment.
func EquipmentHealthScore(eq *Equipment, tasks []MaintenanceTask) int {
	if eq == nil {
		return 0
	}

	score := 100

	// Deduct for fault status
	if eq.Status == EquipmentFault {
		score -= 50
	} else if eq.Status == EquipmentStopped {
		score -= 20
	}

	// Deduct for overdue maintenance
	daysSinceService := time.Since(eq.LastService).Hours() / 24
	if daysSinceService > 180 {
		score -= 30
	} else if daysSinceService > 90 {
		score -= 15
	}

	// Deduct for overdue tasks
	for _, t := range tasks {
		if t.EquipmentID == eq.ID && t.Status == TaskOverdue {
			score -= 10
		}
	}

	if score < 0 {
		score = 0
	}
	return score
}

// ComputeAlertRate returns alerts per hour for a given time window.
func ComputeAlertRate(alerts []Alert, window time.Duration) float64 {
	if len(alerts) == 0 || window <= 0 {
		return 0
	}
	cutoff := time.Now().Add(-window)
	count := 0
	for _, a := range alerts {
		if a.CreatedAt.After(cutoff) {
			count++
		}
	}
	hours := window.Hours()
	if hours == 0 {
		return 0
	}
	return float64(count) / hours
}

// SummarizeAlertsBySeverity returns counts by severity.
func SummarizeAlertsBySeverity(alerts []Alert) map[string]int {
	result := map[string]int{
		SeverityWarning:  0,
		SeverityCritical: 0,
	}
	for _, a := range alerts {
		if a.Resolved {
			continue
		}
		result[a.Severity]++
	}
	return result
}

// FilterReadingsByTimeRange returns readings within [start, end).
func FilterReadingsByTimeRange(readings []SensorReading, start, end time.Time) []SensorReading {
	var result []SensorReading
	for _, r := range readings {
		if (r.Timestamp.Equal(start) || r.Timestamp.After(start)) && r.Timestamp.Before(end) {
			result = append(result, r)
		}
	}
	return result
}

// GroupReadingsByTank returns a map of tank_id -> readings.
func GroupReadingsByTank(readings []SensorReading) map[int64][]SensorReading {
	result := make(map[int64][]SensorReading)
	for _, r := range readings {
		result[r.TankID] = append(result[r.TankID], r)
	}
	return result
}

// FormatDuration human-formats a duration.
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

// ComputeSystemUptime returns percentage uptime from equipment status logs.
func ComputeSystemUptime(statuses map[string]int) float64 {
	total := 0
	running := 0
	for status, count := range statuses {
		total += count
		if status == EquipmentRunning {
			running += count
		}
	}
	if total == 0 {
		return 100
	}
	return float64(running) / float64(total) * 100
}

// PaginateResults returns a page of results from a slice.
func PaginateResults[T any](items []T, page, pageSize int) ([]T, int) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}
	total := len(items)
	start := (page - 1) * pageSize
	if start >= total {
		return []T{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return items[start:end], total
}

// SortAlertsBySeverity sorts alerts: critical first, then warning, then by time descending.
func SortAlertsBySeverity(alerts []Alert) []Alert {
	sort.Slice(alerts, func(i, j int) bool {
		// Unresolved first
		if alerts[i].Resolved != alerts[j].Resolved {
			return !alerts[i].Resolved
		}
		// Critical before warning
		if alerts[i].Severity != alerts[j].Severity {
			return alerts[i].Severity == SeverityCritical
		}
		// Newer first
		return alerts[i].CreatedAt.After(alerts[j].CreatedAt)
	})
	return alerts
}

// SortReadingsByTimestamp sorts readings ascending by timestamp.
func SortReadingsByTimestamp(readings []SensorReading) []SensorReading {
	sorted := make([]SensorReading, len(readings))
	copy(sorted, readings)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})
	return sorted
}

// ChunkReadings splits readings into chunks of the given size.
func ChunkReadings(readings []SensorReading, size int) [][]SensorReading {
	if size <= 0 {
		return nil
	}
	var chunks [][]SensorReading
	for i := 0; i < len(readings); i += size {
		end := i + size
		if end > len(readings) {
			end = len(readings)
		}
		chunks = append(chunks, readings[i:end])
	}
	return chunks
}

// MergeOverlappingAlerts merges alerts that share the same tank+sensor and overlap in time.
func MergeOverlappingAlerts(alerts []Alert) []Alert {
	if len(alerts) < 2 {
		return alerts
	}
	sorted := SortAlertsBySeverity(alerts)
	var merged []Alert
	for _, a := range sorted {
		if len(merged) == 0 {
			merged = append(merged, a)
			continue
		}
		last := &merged[len(merged)-1]
		if last.TankID == a.TankID && last.SensorType == a.SensorType && !a.CreatedAt.After(last.CreatedAt.Add(5*time.Minute)) {
			// Merge into last
			if a.Severity == SeverityCritical {
				last.Severity = SeverityCritical
			}
		} else {
			merged = append(merged, a)
		}
	}
	return merged
}
