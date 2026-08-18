package model

import (
	"errors"
	"fmt"
	"time"
)

// ValidationError wraps a field-specific validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

func (e *ValidationError) Is(target error) bool {
	var ve *ValidationError
	return errors.As(target, &ve)
}

// ValidateTank checks tank fields before create/update.
func ValidateTank(t *Tank) error {
	if t.Name == "" {
		return &ValidationError{Field: "name", Message: "tank name is required"}
	}
	if t.Species == "" {
		return &ValidationError{Field: "species", Message: "species is required"}
	}
	if t.Capacity <= 0 {
		return &ValidationError{Field: "capacity", Message: "capacity must be positive"}
	}
	if t.StockQty < 0 {
		return &ValidationError{Field: "stock_qty", Message: "stock quantity cannot be negative"}
	}
	if !IsValidTankStatus(t.Status) {
		return &ValidationError{Field: "status", Message: "invalid tank status"}
	}
	return nil
}

// ValidateSensorReading checks sensor reading before ingestion.
func ValidateSensorReading(r *SensorReading) error {
	if r.TankID <= 0 {
		return &ValidationError{Field: "tank_id", Message: "tank_id is required"}
	}
	if !IsValidSensorType(r.Type) {
		return &ValidationError{Field: "type", Message: "invalid sensor type"}
	}
	if r.Unit == "" {
		return &ValidationError{Field: "unit", Message: "unit is required"}
	}
	if r.Timestamp.IsZero() {
		return &ValidationError{Field: "timestamp", Message: "timestamp is required"}
	}
	return nil
}

// ValidateThreshold checks threshold config.
func ValidateThreshold(t *ThresholdConfig) error {
	if t.TankID <= 0 {
		return &ValidationError{Field: "tank_id", Message: "tank_id is required"}
	}
	if !IsValidSensorType(t.SensorType) {
		return &ValidationError{Field: "sensor_type", Message: "invalid sensor type"}
	}
	if t.MinValue >= t.MaxValue {
		return &ValidationError{Field: "min_value", Message: "min must be less than max"}
	}
	return nil
}

// ValidateFeedPlan checks feed plan before create/update.
func ValidateFeedPlan(f *FeedPlan) error {
	if f.TankID <= 0 {
		return &ValidationError{Field: "tank_id", Message: "tank_id is required"}
	}
	if f.FeedType == "" {
		return &ValidationError{Field: "feed_type", Message: "feed_type is required"}
	}
	if f.Amount <= 0 {
		return &ValidationError{Field: "amount", Message: "amount must be positive"}
	}
	if f.Schedule == "" {
		return &ValidationError{Field: "schedule", Message: "schedule is required"}
	}
	return nil
}

// ValidateEquipment checks equipment fields.
func ValidateEquipment(e *Equipment) error {
	if e.TankID <= 0 {
		return &ValidationError{Field: "tank_id", Message: "tank_id is required"}
	}
	if e.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	if e.Type == "" {
		return &ValidationError{Field: "type", Message: "type is required"}
	}
	if !IsValidEquipmentStatus(e.Status) {
		return &ValidationError{Field: "status", Message: "invalid equipment status"}
	}
	if e.PowerRating < 0 {
		return &ValidationError{Field: "power_rating", Message: "power rating cannot be negative"}
	}
	return nil
}

// ValidateMaintenanceTask checks maintenance task fields.
func ValidateMaintenanceTask(t *MaintenanceTask) error {
	if t.EquipmentID <= 0 {
		return &ValidationError{Field: "equipment_id", Message: "equipment_id is required"}
	}
	if t.Type == "" {
		return &ValidationError{Field: "type", Message: "type is required"}
	}
	if t.Description == "" {
		return &ValidationError{Field: "description", Message: "description is required"}
	}
	if t.ScheduledAt.IsZero() {
		return &ValidationError{Field: "scheduled_at", Message: "scheduled_at is required"}
	}
	return nil
}

// ValidateWaterChange checks water change record.
func ValidateWaterChange(w *WaterChange) error {
	if w.TankID <= 0 {
		return &ValidationError{Field: "tank_id", Message: "tank_id is required"}
	}
	if w.Volume <= 0 {
		return &ValidationError{Field: "volume", Message: "volume must be positive"}
	}
	if w.Reason == "" {
		return &ValidationError{Field: "reason", Message: "reason is required"}
	}
	if w.Operator == "" {
		return &ValidationError{Field: "operator", Message: "operator is required"}
	}
	return nil
}

// ValidateMortalityLog checks mortality log.
func ValidateMortalityLog(m *MortalityLog) error {
	if m.TankID <= 0 {
		return &ValidationError{Field: "tank_id", Message: "tank_id is required"}
	}
	if m.Count <= 0 {
		return &ValidationError{Field: "count", Message: "count must be positive"}
	}
	if m.Cause == "" {
		return &ValidationError{Field: "cause", Message: "cause is required"}
	}
	return nil
}

// ValidateBatch checks batch record.
func ValidateBatch(b *Batch) error {
	if b.TankID <= 0 {
		return &ValidationError{Field: "tank_id", Message: "tank_id is required"}
	}
	if b.Species == "" {
		return &ValidationError{Field: "species", Message: "species is required"}
	}
	if b.InitialCount <= 0 {
		return &ValidationError{Field: "initial_count", Message: "initial_count must be positive"}
	}
	if b.CurrentCount < 0 {
		return &ValidationError{Field: "current_count", Message: "current_count cannot be negative"}
	}
	if b.CurrentCount > b.InitialCount {
		return &ValidationError{Field: "current_count", Message: "current_count cannot exceed initial_count"}
	}
	if b.ArrivalDate.IsZero() {
		return &ValidationError{Field: "arrival_date", Message: "arrival_date is required"}
	}
	if b.Source == "" {
		return &ValidationError{Field: "source", Message: "source is required"}
	}
	return nil
}

// ParseSchedule parses a cron-like schedule string "08:00,16:00" into time slots.
// Times are parsed in the local timezone so they align with the farm's
// operating hours and can be compared correctly in ComputeDailyStats.
func ParseSchedule(schedule string) ([]time.Time, error) {
	if schedule == "" {
		return nil, &ValidationError{Field: "schedule", Message: "empty schedule"}
	}
	parts := splitSchedule(schedule)
	slots := make([]time.Time, 0, len(parts))
	now := time.Now()
	localLoc := now.Location()
	for _, p := range parts {
		t, err := time.ParseInLocation("15:04", p, localLoc)
		if err != nil {
			return nil, &ValidationError{Field: "schedule", Message: fmt.Sprintf("invalid time slot: %s", p)}
		}
		slot := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, localLoc)
		slots = append(slots, slot)
	}
	return slots, nil
}

func splitSchedule(s string) []string {
	result := make([]string, 0)
	current := ""
	for _, ch := range s {
		if ch == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else if ch != ' ' {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// ComputeStockDensity returns fish per liter.
func ComputeStockDensity(tank *Tank) float64 {
	if tank.Capacity <= 0 {
		return 0
	}
	return float64(tank.StockQty) / tank.Capacity
}

// ComputeSurvivalRate returns survival rate as a percentage.
func ComputeSurvivalRate(b *Batch) float64 {
	if b.InitialCount <= 0 {
		return 0
	}
	return float64(b.CurrentCount) / float64(b.InitialCount) * 100
}

// ComputeFeedEfficiency returns feed amount per fish in grams.
func ComputeFeedEfficiency(logs []FeedLog, stockQty int) float64 {
	if stockQty <= 0 || len(logs) == 0 {
		return 0
	}
	total := 0.0
	for _, l := range logs {
		total += l.Amount
	}
	return total / float64(stockQty)
}

// IsAlertOverdue returns true if an unresolved alert exceeds the given duration.
func IsAlertOverdue(a *Alert, duration time.Duration) bool {
	if a.Resolved {
		return false
	}
	return time.Since(a.CreatedAt) > duration
}

// ShouldEscalateAlert returns true if a warning alert should escalate to critical.
func ShouldEscalateAlert(a *Alert, readings []SensorReading) bool {
	if a.Resolved || a.Severity != SeverityWarning {
		return false
	}
	for _, r := range readings {
		if r.TankID == a.TankID && r.Type == a.SensorType && r.Timestamp.After(a.CreatedAt) {
			if a.Severity == SeverityWarning && (r.Value < 0 || r.Value > 100) {
				return true
			}
		}
	}
	return false
}

// FormatAlertMessage builds a human-readable alert message.
func FormatAlertMessage(tankName, sensorType string, value, min, max float64, severity string) string {
	return fmt.Sprintf("Tank %s %s=%.2f (range %.2f-%.2f) [%s]",
		tankName, sensorType, value, min, max, severity)
}

// ComputeDailyStats aggregates readings into a daily summary.
func ComputeDailyStats(readings []SensorReading, feeds []FeedLog, mortality int) *DailySummary {
	if len(readings) == 0 {
		return &DailySummary{Mortality: mortality}
	}
	tempSum, phSum, count := 0.0, 0.0, 0
	minOxygen := 999.0
	maxAmmonia := 0.0
	date := ""
	for _, r := range readings {
		switch r.Type {
		case SensorTemperature:
			tempSum += r.Value
			count++
		case SensorPH:
			phSum += r.Value
		case SensorOxygen:
			if r.Value < minOxygen {
				minOxygen = r.Value
			}
		case SensorAmmonia:
			if r.Value > maxAmmonia {
				maxAmmonia = r.Value
			}
		}
		if date == "" {
			date = r.Timestamp.Format("2006-01-02")
		}
	}
	feedTotal := 0.0
	for _, f := range feeds {
		feedTotal += f.Amount
	}
	summary := &DailySummary{
		Date:       date,
		MaxAmmonia: maxAmmonia,
		FeedTotal:  feedTotal,
		Mortality:  mortality,
	}
	if count > 0 {
		summary.AvgTemp = tempSum / float64(count)
	}
	if len(readings) > 0 {
		summary.AvgPH = phSum / float64(len(readings))
	}
	if minOxygen < 999 {
		summary.MinOxygen = minOxygen
	}
	return summary
}
