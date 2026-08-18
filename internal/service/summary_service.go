package service

import (
	"context"
	"fmt"
	"time"

	"aquafarm/internal/model"
)

// GetDailySummaryForToday returns today's daily summary for a tank.
func (s *Service) GetDailySummaryForToday(ctx context.Context, tankID int64) (*model.DailySummary, error) {
	return s.GetDailySummary(ctx, tankID, time.Now())
}

// GetWeeklySummary returns a 7-day summary for a tank.
func (s *Service) GetWeeklySummary(ctx context.Context, tankID int64) ([]model.DailySummary, error) {
	result := make([]model.DailySummary, 0, 7)
	now := time.Now()

	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		summary, err := s.GetDailySummary(ctx, tankID, day)
		if err != nil {
			return nil, fmt.Errorf("get daily summary for %s: %w", day.Format("2006-01-02"), err)
		}
		result = append(result, *summary)
	}
	return result, nil
}

// GetTankReadingsInRange returns all readings for a tank within a time range.
func (s *Service) GetTankReadingsInRange(ctx context.Context, tankID int64, start, end time.Time) ([]model.SensorReading, error) {
	return s.repo.Readings().ListByTankAndTimeRange(tankID, start, end)
}

// GetReadingStats returns min/max/avg stats for a tank since a given time.
func (s *Service) GetReadingStats(ctx context.Context, tankID int64, since time.Time) (map[string]model.DailySummary, error) {
	return s.repo.Readings().GetStatsByTank(tankID, since)
}

// ComputeSurvivalRate returns the survival rate percentage for the active batch of a tank.
func (s *Service) ComputeSurvivalRate(ctx context.Context, tankID int64) (float64, error) {
	batches, err := s.repo.Batches().ListByTank(tankID)
	if err != nil {
		return 0, fmt.Errorf("list batches: %w", err)
	}
	for _, b := range batches {
		if b.CurrentCount > 0 {
			return model.ComputeSurvivalRate(&b), nil
		}
	}
	return 0, nil
}

// ComputeStockDensity returns fish-per-liter for a tank.
func (s *Service) ComputeStockDensity(ctx context.Context, tankID int64) (float64, error) {
	tank, err := s.repo.Tanks().GetByID(tankID)
	if err != nil {
		return 0, fmt.Errorf("get tank: %w", err)
	}
	if tank == nil {
		return 0, model.ErrNotFound
	}
	return model.ComputeStockDensity(tank), nil
}

// GetFeedEfficiency returns feed amount per fish in grams.
func (s *Service) GetFeedEfficiency(ctx context.Context, tankID int64) (float64, error) {
	tank, err := s.repo.Tanks().GetByID(tankID)
	if err != nil {
		return 0, fmt.Errorf("get tank: %w", err)
	}
	if tank == nil {
		return 0, model.ErrNotFound
	}
	if tank.StockQty == 0 {
		return 0, nil
	}
	logs, err := s.repo.Feeds().ListFeedLogs(tankID, 100)
	if err != nil {
		return 0, fmt.Errorf("get feed logs: %w", err)
	}
	return model.ComputeFeedEfficiency(logs, tank.StockQty), nil
}

// GetAlertRate returns alerts per hour for the last N hours.
func (s *Service) GetAlertRate(ctx context.Context, hours int) (float64, error) {
	if hours <= 0 {
		hours = 24
	}
	alerts, err := s.repo.Alerts().List(1000)
	if err != nil {
		return 0, fmt.Errorf("list alerts: %w", err)
	}
	return model.ComputeAlertRate(alerts, time.Duration(hours)*time.Hour), nil
}

// DetectAnomalies returns readings that deviate significantly from the mean.
func (s *Service) DetectAnomalies(ctx context.Context, tankID int64, sensorType string, threshold float64) ([]model.SensorReading, error) {
	if threshold <= 0 {
		threshold = 5.0
	}
	readings, err := s.repo.Readings().ListByTankAndType(tankID, sensorType, 100)
	if err != nil {
		return nil, fmt.Errorf("list readings: %w", err)
	}
	return model.DetectAnomalies(readings, threshold), nil
}

// GetMovingAverage returns the moving average of the last N readings.
func (s *Service) GetMovingAverage(ctx context.Context, tankID int64, sensorType string, window int) ([]float64, error) {
	if window <= 0 {
		window = 5
	}
	readings, err := s.repo.Readings().ListByTankAndType(tankID, sensorType, 50)
	if err != nil {
		return nil, fmt.Errorf("list readings: %w", err)
	}
	return model.ComputeMovingAverage(readings, window), nil
}

// GetEquipmentOverview returns health scores for all equipment in a tank.
func (s *Service) GetEquipmentOverview(ctx context.Context, tankID int64) (map[int64]int, error) {
	equipment, err := s.repo.Equipment().ListByTank(tankID)
	if err != nil {
		return nil, fmt.Errorf("list equipment: %w", err)
	}
	result := make(map[int64]int)
	for _, eq := range equipment {
		tasks, _ := s.repo.Equipment().ListTasksByEquipment(eq.ID)
		result[eq.ID] = model.EquipmentHealthScore(&eq, tasks)
	}
	return result, nil
}

// CleanupOldReadings removes readings older than the specified duration.
func (s *Service) CleanupOldReadings(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	return s.repo.Readings().DeleteOlderThan(cutoff)
}

// GetSystemHealthScore returns an overall system health score 0-100.
func (s *Service) GetSystemHealthScore(ctx context.Context) (int, error) {
	overview, err := s.repo.GetSystemOverview()
	if err != nil {
		return 0, err
	}

	score := 100

	// Deduct for unresolved alerts
	score -= overview.UnresolvedAlerts * 5

	// Deduct for faulted equipment
	score -= overview.EquipmentFault * 10

	// Deduct for overdue tasks
	score -= overview.OverdueTasks * 3

	// Bonus for running equipment
	if overview.EquipmentRunning > 0 && overview.EquipmentFault == 0 {
		score += 5
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score, nil
}
