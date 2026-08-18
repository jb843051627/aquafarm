package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"aquafarm/internal/model"
	"aquafarm/internal/store"
)

// TestReproBug006 reproduces the nil-pointer panic reported when querying
// summary info for a non-existent tank. Kept as a regression guard.
func TestReproBug006(t *testing.T) {
	repo, err := store.Open(t.TempDir() + "/bug006_repro.db")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer repo.Close()
	if err := repo.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	svc := New(repo)
	ctx := context.Background()
	const ghostID int64 = 999999

	cases := []struct {
		name string
		fn   func() error
	}{
		{"GetTankSummary", func() error { _, err := svc.GetTankSummary(ctx, ghostID); return err }},
		{"ComputeStockDensity", func() error { _, err := svc.ComputeStockDensity(ctx, ghostID); return err }},
		{"GetDailySummaryForToday", func() error { _, err := svc.GetDailySummaryForToday(ctx, ghostID); return err }},
		{"GetWeeklySummary", func() error { _, err := svc.GetWeeklySummary(ctx, ghostID); return err }},
		{"GetFeedEfficiency", func() error { _, err := svc.GetFeedEfficiency(ctx, ghostID); return err }},
		{"ComputeSurvivalRate", func() error { _, err := svc.ComputeSurvivalRate(ctx, ghostID); return err }},
		{"GetReadingStats", func() error { _, err := svc.GetReadingStats(ctx, ghostID, time.Now().Add(-24*time.Hour)); return err }},
		{"GetTankReadingsInRange", func() error { _, err := svc.GetTankReadingsInRange(ctx, ghostID, time.Now().Add(-24*time.Hour), time.Now()); return err }},
		{"GetMovingAverage", func() error { _, err := svc.GetMovingAverage(ctx, ghostID, "temperature", 5); return err }},
		{"DetectAnomalies", func() error { _, err := svc.DetectAnomalies(ctx, ghostID, "temperature", 5); return err }},
		{"GetEquipmentOverview", func() error { _, err := svc.GetEquipmentOverview(ctx, ghostID); return err }},
		{"GetReadingTrend", func() error { _, err := svc.GetReadingTrend(ctx, ghostID, "temperature"); return err }},
		{"GetLatestReadings", func() error { _, err := svc.GetLatestReadings(ctx, ghostID); return err }},
		// Direct model-layer nil deref probe (regression guard for bug-006):
		{"model.ComputeStockDensity(nil)", func() error {
			if got := model.ComputeStockDensity(nil); got != 0 {
				return fmt.Errorf("expected 0 for nil tank, got %v", got)
			}
			return nil
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("PANIC on non-existent tank: %v", r)
				}
			}()
			err := c.fn()
			t.Logf("%s -> err=%v", c.name, err)
		})
	}
}
