package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"aquafarm/internal/model"
)

// IngestReading ingests a sensor reading and checks thresholds.
func (s *Service) IngestReading(ctx context.Context, r *model.SensorReading) (*model.SensorReading, error) {
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now()
	}

	if err := model.ValidateSensorReading(r); err != nil {
		return nil, err
	}

	// Verify tank exists
	tank, err := s.repo.Tanks().GetByID(r.TankID)
	if err != nil {
		return nil, fmt.Errorf("check tank: %w", err)
	}
	if tank == nil {
		return nil, model.ErrNotFound
	}

	if err := s.repo.Readings().Create(r); err != nil {
		return nil, fmt.Errorf("create reading: %w", err)
	}

	// Check thresholds and create alerts if needed
	if ctxErr := ctx.Err(); ctxErr != nil {
		return r, nil // reading stored, but skip threshold check
	}

	s.checkThresholds(ctx, tank, r)

	return r, nil
}

// BatchIngestReadings ingests multiple sensor readings.
func (s *Service) BatchIngestReadings(ctx context.Context, readings []model.SensorReading) (int, error) {
	if len(readings) == 0 {
		return 0, nil
	}

	// Validate all readings
	for i := range readings {
		if readings[i].Timestamp.IsZero() {
			readings[i].Timestamp = time.Now()
		}
		if err := model.ValidateSensorReading(&readings[i]); err != nil {
			return 0, err
		}
	}

	if err := s.repo.Readings().BatchInsert(readings); err != nil {
		return 0, fmt.Errorf("batch insert readings: %w", err)
	}

	// Check thresholds for all readings
	for _, r := range readings {
		tank, err := s.repo.Tanks().GetByID(r.TankID)
		if err != nil || tank == nil {
			continue
		}
		s.checkThresholds(ctx, tank, &r)
	}

	return len(readings), nil
}

// checkThresholds evaluates a reading against configured thresholds and creates alerts.
func (s *Service) checkThresholds(ctx context.Context, tank *model.Tank, r *model.SensorReading) {
	thresholds, err := s.repo.Config().GetThresholdsByTank(tank.ID)
	if err != nil {
		return
	}

	for _, th := range thresholds {
		if !th.Enabled || th.SensorType != r.Type {
			continue
		}

		var severity string
		if r.Value < th.MinValue || r.Value > th.MaxValue {
			severity = model.SeverityWarning
		}
		// Critical if 50% beyond range
		margin := th.MaxValue - th.MinValue
		if r.Value < th.MinValue-margin*0.5 || r.Value > th.MaxValue+margin*0.5 {
			severity = model.SeverityCritical
		}

		if severity == "" {
			continue
		}

		msg := model.FormatAlertMessage(tank.Name, r.Type, r.Value, th.MinValue, th.MaxValue, severity)
		alert := &model.Alert{
			TankID:      tank.ID,
			SensorType:  r.Type,
			Value:       r.Value,
			Severity:    severity,
			Message:     msg,
			CreatedAt:   time.Now(),
		}
		_ = s.repo.Alerts().Create(alert)
	}
}

// GetLatestReadings returns the latest reading for each sensor type on a tank.
func (s *Service) GetLatestReadings(ctx context.Context, tankID int64) (map[string]*model.SensorReading, error) {
	result := make(map[string]*model.SensorReading)
	sensorTypes := []string{
		model.SensorTemperature, model.SensorPH, model.SensorOxygen,
		model.SensorAmmonia, model.SensorNitrite,
	}
	for _, st := range sensorTypes {
		reading, err := s.repo.Readings().GetLatestByTankAndType(tankID, st)
		if err != nil && !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
		if reading != nil {
			result[st] = reading
		}
	}
	return result, nil
}

// ListReadingsByTank returns readings for a tank.
func (s *Service) ListReadingsByTank(ctx context.Context, tankID int64, limit int) ([]model.SensorReading, error) {
	return s.repo.Readings().ListByTank(tankID, limit)
}

// ListReadingsByTankAndType returns readings for a tank+sensor type.
func (s *Service) ListReadingsByTankAndType(ctx context.Context, tankID int64, sensorType string, limit int) ([]model.SensorReading, error) {
	return s.repo.Readings().ListByTankAndType(tankID, sensorType, limit)
}

// GetReadingTrend computes a trend analysis for a tank+sensor type.
func (s *Service) GetReadingTrend(ctx context.Context, tankID int64, sensorType string) (*model.TrendAnalysis, error) {
	readings, err := s.repo.Readings().ListByTankAndType(tankID, sensorType, 50)
	if err != nil {
		return nil, err
	}
	return model.AnalyzeTrend(readings), nil
}

// ListAlerts returns all alerts.
func (s *Service) ListAlerts(ctx context.Context, limit int) ([]model.Alert, error) {
	return s.repo.Alerts().List(limit)
}

// ListAlertsByTank returns alerts for a tank.
func (s *Service) ListAlertsByTank(ctx context.Context, tankID int64, limit int) ([]model.Alert, error) {
	alerts, err := s.repo.Alerts().ListByTank(tankID, limit)
	if err != nil {
		return nil, err
	}
	return model.SortAlertsBySeverity(alerts), nil
}

// ListUnresolvedAlerts returns unresolved alerts.
func (s *Service) ListUnresolvedAlerts(ctx context.Context, limit int) ([]model.Alert, error) {
	return s.repo.Alerts().ListUnresolved(limit)
}

// ResolveAlert marks an alert as resolved.
func (s *Service) ResolveAlert(ctx context.Context, id int64) error {
	alert, err := s.repo.Alerts().GetByID(id)
	if err != nil {
		return fmt.Errorf("get alert %d: %w", id, err)
	}
	if alert == nil {
		return model.ErrNotFound
	}
	if alert.Resolved {
		return &model.ValidationError{Field: "resolved", Message: "alert already resolved"}
	}
	return s.repo.Alerts().Resolve(id)
}

// ResolveAlertsByTank resolves all unresolved alerts for a tank.
func (s *Service) ResolveAlertsByTank(ctx context.Context, tankID int64) (int64, error) {
	return s.repo.Alerts().ResolveByTank(tankID)
}

// GetAlertStats returns alert statistics.
func (s *Service) GetAlertStats(ctx context.Context) (map[string]int, error) {
	return s.repo.Alerts().CountBySeverity()
}

// SaveThreshold saves a threshold config.
func (s *Service) SaveThreshold(ctx context.Context, t *model.ThresholdConfig) (*model.ThresholdConfig, error) {
	if err := model.ValidateThreshold(t); err != nil {
		return nil, err
	}
	// Verify tank exists
	tank, err := s.repo.Tanks().GetByID(t.TankID)
	if err != nil {
		return nil, fmt.Errorf("check tank: %w", err)
	}
	if tank == nil {
		return nil, model.ErrNotFound
	}
	if err := s.repo.Config().SaveThreshold(t); err != nil {
		return nil, fmt.Errorf("save threshold: %w", err)
	}
	return t, nil
}

// GetThresholdsByTank returns threshold configs for a tank.
func (s *Service) GetThresholdsByTank(ctx context.Context, tankID int64) ([]model.ThresholdConfig, error) {
	return s.repo.Config().GetThresholdsByTank(tankID)
}

// DeleteThreshold removes a threshold config.
func (s *Service) DeleteThreshold(ctx context.Context, id int64) error {
	return s.repo.Config().DeleteThreshold(id)
}

// readingCache is a thread-safe in-memory cache for recent readings.
type readingCache struct {
	mu       sync.RWMutex
	latest   map[int64]map[string]model.SensorReading // tank_id -> sensor_type -> reading
	alerts   map[int64][]model.Alert                   // tank_id -> recent alerts
	maxAlert int
}

// NewReadingCache creates a new reading cache.
func NewReadingCache() *readingCache {
	return &readingCache{
		latest:   make(map[int64]map[string]model.SensorReading),
		alerts:   make(map[int64][]model.Alert),
		maxAlert: 50,
	}
}

// Update stores the latest reading in the cache.
func (c *readingCache) Update(r model.SensorReading) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.latest[r.TankID]; !ok {
		c.latest[r.TankID] = make(map[string]model.SensorReading)
	}
	c.latest[r.TankID][r.Type] = r
}

// GetLatest returns the cached latest reading.
func (c *readingCache) GetLatest(tankID int64, sensorType string) (model.SensorReading, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if tank, ok := c.latest[tankID]; ok {
		if r, ok := tank[sensorType]; ok {
			return r, true
		}
	}
	return model.SensorReading{}, false
}

// AddAlert adds an alert to the cache.
func (c *readingCache) AddAlert(a model.Alert) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.alerts[a.TankID] = append(c.alerts[a.TankID], a)
	if len(c.alerts[a.TankID]) > c.maxAlert {
		c.alerts[a.TankID] = c.alerts[a.TankID][1:]
	}
}

// GetAlerts returns cached alerts for a tank.
func (c *readingCache) GetAlerts(tankID int64) []model.Alert {
	c.mu.RLock()
	defer c.mu.RUnlock()
	alerts := c.alerts[tankID]
	result := make([]model.Alert, len(alerts))
	copy(result, alerts)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result
}
