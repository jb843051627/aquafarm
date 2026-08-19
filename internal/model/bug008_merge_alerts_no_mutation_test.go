package model

import (
	"testing"
	"time"
)

// TestBug008_MergeAlertsNoMutation verifies that MergeOverlappingAlerts does not modify
// the elements of the input slice.
// Bug: MergeOverlappingAlerts mutates input slice elements via pointer to merged slice.
func TestBug008_MergeAlertsNoMutation(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(1 * time.Minute)

	alerts := []Alert{
		{ID: 1, TankID: 1, SensorType: SensorTemperature, Severity: SeverityWarning, CreatedAt: t1},
		{ID: 2, TankID: 1, SensorType: SensorTemperature, Severity: SeverityWarning, CreatedAt: t2},
	}

	// Snapshot original severities
	origSeverity0 := alerts[0].Severity
	origSeverity1 := alerts[1].Severity
	origID0 := alerts[0].ID
	origID1 := alerts[1].ID

	_ = MergeOverlappingAlerts(alerts)

	// Verify input slice elements are not mutated
	if alerts[0].Severity != origSeverity0 {
		t.Errorf("input slice[0] severity mutated: got %q, want %q", alerts[0].Severity, origSeverity0)
	}
	if alerts[1].Severity != origSeverity1 {
		t.Errorf("input slice[1] severity mutated: got %q, want %q", alerts[1].Severity, origSeverity1)
	}
	if alerts[0].ID != origID0 {
		t.Errorf("input slice[0] ID mutated: got %d, want %d", alerts[0].ID, origID0)
	}
	if alerts[1].ID != origID1 {
		t.Errorf("input slice[1] ID mutated: got %d, want %d", alerts[1].ID, origID1)
	}
}