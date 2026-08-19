package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"aquafarm/internal/model"
	"aquafarm/internal/store"
)

// TestBug005_BatchIngestContextCancellation verifies that BatchIngestReadings stops processing
// when the context is already cancelled before the threshold-checking loop.
// Bug: the loop does not check ctx.Err(), so it continues processing after cancellation.
func TestBug005_BatchIngestContextCancellation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	repo, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	defer repo.Close()

	if err := repo.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Create a tank
	tank := &model.Tank{
		Name:     "TestTank",
		Species:  "Salmon",
		Capacity: 1000,
		StockQty: 50,
		Status:   model.TankStatusActive,
	}
	if err := repo.Tanks().Create(tank); err != nil {
		t.Fatalf("create tank: %v", err)
	}

	// Create a threshold config so checkThresholds does real work
	th := &model.ThresholdConfig{
		TankID:      tank.ID,
		SensorType:  model.SensorTemperature,
		MinValue:    10,
		MaxValue:    30,
		Enabled:     true,
	}
	if err := repo.Config().SaveThreshold(th); err != nil {
		t.Fatalf("save threshold: %v", err)
	}

	svc := New(repo)

	// Build many readings that exceed threshold (triggers alert creation in loop)
	readings := make([]model.SensorReading, 100)
	for i := range readings {
		readings[i] = model.SensorReading{
			TankID:    tank.ID,
			Type:      model.SensorTemperature,
			Value:     50.0, // exceeds max=30 → triggers alert
			Unit:      "C",
			Timestamp: time.Now(),
		}
	}

	// Use a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = svc.BatchIngestReadings(ctx, readings)
	if err != nil {
		t.Fatalf("BatchIngestReadings with cancelled ctx returned error: %v", err)
	}

	// If ctx was respected, the loop should have broken early and not created
	// alerts for all 100 readings. Check alert count.
	alerts, err := repo.Alerts().List(1000)
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}

	// With ctx cancellation, we expect far fewer than 100 alerts.
	// The bug processes all 100 readings without checking ctx.
	if len(alerts) == 100 {
		t.Fatalf("expected fewer than 100 alerts due to ctx cancellation, got %d (ctx not checked)", len(alerts))
	}
}