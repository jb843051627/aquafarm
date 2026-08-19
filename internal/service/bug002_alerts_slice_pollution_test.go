package service

import (
	"testing"
	"time"

	"aquafarm/internal/model"
)

// TestBug002_GetAlertsSlicePollution verifies that GetAlerts returns a copy of the internal slice,
// so modifying the returned slice does not pollute the cache.
// Bug: GetAlerts returns the internal slice directly without copy.
func TestBug002_GetAlertsSlicePollution(t *testing.T) {
	cache := NewReadingCache()

	alert := model.Alert{
		ID:        1,
		TankID:    1,
		SensorType: model.SensorTemperature,
		Value:     30.5,
		Severity:  model.SeverityWarning,
		Message:   "test alert",
		CreatedAt: time.Now(),
	}
	cache.AddAlert(alert)

	// Get alerts — should return a copy
	got := cache.GetAlerts(1)
	if len(got) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(got))
	}

	// Mutate the returned slice
	if len(got) > 0 {
		got[0].Message = "POLLUTED"
		got[0].Value = 999.0
	}

	// Get again — should be unmodified
	got2 := cache.GetAlerts(1)
	if len(got2) != 1 {
		t.Fatalf("expected 1 alert after mutation, got %d", len(got2))
	}
	if got2[0].Message != "test alert" {
		t.Errorf("cache polluted: message = %q, want %q", got2[0].Message, "test alert")
	}
	if got2[0].Value != 30.5 {
		t.Errorf("cache polluted: value = %f, want 30.5", got2[0].Value)
	}
}