package service

import (
	"context"
	"fmt"
	"time"

	"aquafarm/internal/model"
)

// CreateEquipment creates a new equipment record.
func (s *Service) CreateEquipment(ctx context.Context, e *model.Equipment) (*model.Equipment, error) {
	if e.InstalledAt.IsZero() {
		e.InstalledAt = time.Now()
	}
	if e.LastService.IsZero() {
		e.LastService = e.InstalledAt
	}
	if e.Status == "" {
		e.Status = model.EquipmentRunning
	}

	if err := model.ValidateEquipment(e); err != nil {
		return nil, err
	}

	// Verify tank exists
	tank, err := s.repo.Tanks().GetByID(e.TankID)
	if err != nil {
		return nil, fmt.Errorf("check tank: %w", err)
	}
	if tank == nil {
		return nil, model.ErrNotFound
	}

	if err := s.repo.Equipment().Create(e); err != nil {
		return nil, fmt.Errorf("create equipment: %w", err)
	}
	return e, nil
}

// GetEquipment retrieves equipment by ID.
func (s *Service) GetEquipment(ctx context.Context, id int64) (*model.Equipment, error) {
	eq, err := s.repo.Equipment().GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get equipment %d: %w", id, err)
	}
	if eq == nil {
		return nil, model.ErrNotFound
	}
	return eq, nil
}

// ListEquipment returns all equipment.
func (s *Service) ListEquipment(ctx context.Context) ([]model.Equipment, error) {
	return s.repo.Equipment().List()
}

// ListEquipmentByTank returns equipment for a tank.
func (s *Service) ListEquipmentByTank(ctx context.Context, tankID int64) ([]model.Equipment, error) {
	return s.repo.Equipment().ListByTank(tankID)
}

// ListEquipmentByStatus returns equipment filtered by status.
func (s *Service) ListEquipmentByStatus(ctx context.Context, status string) ([]model.Equipment, error) {
	return s.repo.Equipment().ListByStatus(status)
}

// UpdateEquipment updates equipment fields.
func (s *Service) UpdateEquipment(ctx context.Context, e *model.Equipment) (*model.Equipment, error) {
	if err := model.ValidateEquipment(e); err != nil {
		return nil, err
	}
	if err := s.repo.Equipment().Update(e); err != nil {
		return nil, fmt.Errorf("update equipment: %w", err)
	}
	return e, nil
}

// UpdateEquipmentStatus updates only the status.
func (s *Service) UpdateEquipmentStatus(ctx context.Context, id int64, status string) error {
	if !model.IsValidEquipmentStatus(status) {
		return &model.ValidationError{Field: "status", Message: "invalid equipment status"}
	}
	return s.repo.Equipment().UpdateStatus(id, status)
}

// DeleteEquipment removes equipment by ID.
func (s *Service) DeleteEquipment(ctx context.Context, id int64) error {
	return s.repo.Equipment().Delete(id)
}

// CreateMaintenanceTask creates a new maintenance task.
func (s *Service) CreateMaintenanceTask(ctx context.Context, t *model.MaintenanceTask) (*model.MaintenanceTask, error) {
	if t.Status == "" {
		t.Status = model.TaskPending
	}

	if err := model.ValidateMaintenanceTask(t); err != nil {
		return nil, err
	}

	// Verify equipment exists
	eq, err := s.repo.Equipment().GetByID(t.EquipmentID)
	if err != nil {
		return nil, fmt.Errorf("check equipment: %w", err)
	}
	if eq == nil {
		return nil, model.ErrNotFound
	}

	if err := s.repo.Equipment().CreateTask(t); err != nil {
		return nil, fmt.Errorf("create maintenance task: %w", err)
	}
	return t, nil
}

// GetMaintenanceTask retrieves a maintenance task by ID.
func (s *Service) GetMaintenanceTask(ctx context.Context, id int64) (*model.MaintenanceTask, error) {
	task, err := s.repo.Equipment().GetTaskByID(id)
	if err != nil {
		return nil, fmt.Errorf("get maintenance task %d: %w", id, err)
	}
	if task == nil {
		return nil, model.ErrNotFound
	}
	return task, nil
}

// ListMaintenanceTasks returns all maintenance tasks.
func (s *Service) ListMaintenanceTasks(ctx context.Context, limit int) ([]model.MaintenanceTask, error) {
	return s.repo.Equipment().ListTasks(limit)
}

// ListTasksByEquipment returns tasks for a specific equipment.
func (s *Service) ListTasksByEquipment(ctx context.Context, equipID int64) ([]model.MaintenanceTask, error) {
	return s.repo.Equipment().ListTasksByEquipment(equipID)
}

// ListTasksByStatus returns tasks filtered by status.
func (s *Service) ListTasksByStatus(ctx context.Context, status string) ([]model.MaintenanceTask, error) {
	return s.repo.Equipment().ListTasksByStatus(status)
}

// CompleteMaintenanceTask marks a task as completed and updates equipment's last service.
func (s *Service) CompleteMaintenanceTask(ctx context.Context, id int64, technician string) error {
	task, err := s.repo.Equipment().GetTaskByID(id)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}
	if task == nil {
		return model.ErrNotFound
	}
	if task.Status == model.TaskCompleted {
		return &model.ValidationError{Field: "status", Message: "task already completed"}
	}

	// Complete the task
	if err := s.repo.Equipment().CompleteTask(id, technician); err != nil {
		return fmt.Errorf("complete task: %w", err)
	}

	// Update equipment last service time
	if err := s.repo.Equipment().UpdateLastService(task.EquipmentID, time.Now()); err != nil {
		return fmt.Errorf("update equipment service time: %w", err)
	}

	return nil
}

// ListOverdueTasks returns maintenance tasks that are past their scheduled time.
func (s *Service) ListOverdueTasks(ctx context.Context) ([]model.MaintenanceTask, error) {
	return s.repo.Equipment().ListOverdueTasks()
}

// GetEquipmentHealth computes a health score for equipment.
func (s *Service) GetEquipmentHealth(ctx context.Context, id int64) (int, error) {
	eq, err := s.repo.Equipment().GetByID(id)
	if err != nil {
		return 0, fmt.Errorf("get equipment: %w", err)
	}
	if eq == nil {
		return 0, model.ErrNotFound
	}
	tasks, err := s.repo.Equipment().ListTasksByEquipment(id)
	if err != nil && err != model.ErrNotFound {
		return 0, err
	}
	return model.EquipmentHealthScore(eq, tasks), nil
}

// CreateBatch creates a new fish batch.
func (s *Service) CreateBatch(ctx context.Context, b *model.Batch) (*model.Batch, error) {
	if b.ArrivalDate.IsZero() {
		b.ArrivalDate = time.Now()
	}
	b.CurrentCount = b.InitialCount

	if err := model.ValidateBatch(b); err != nil {
		return nil, err
	}

	// Verify tank exists
	tank, err := s.repo.Tanks().GetByID(b.TankID)
	if err != nil {
		return nil, fmt.Errorf("check tank: %w", err)
	}
	if tank == nil {
		return nil, model.ErrNotFound
	}

	if err := s.repo.Batches().Create(b); err != nil {
		return nil, fmt.Errorf("create batch: %w", err)
	}

	// Update tank stock count
	newStock := tank.StockQty + b.InitialCount
	if err := s.repo.Tanks().UpdateStock(b.TankID, newStock); err != nil {
		return nil, fmt.Errorf("update tank stock: %w", err)
	}

	return b, nil
}

// ListBatchesByTank returns batches for a tank.
func (s *Service) ListBatchesByTank(ctx context.Context, tankID int64) ([]model.Batch, error) {
	return s.repo.Batches().ListByTank(tankID)
}

// ListBatches returns all batches.
func (s *Service) ListBatches(ctx context.Context) ([]model.Batch, error) {
	return s.repo.Batches().List()
}

// RecordMortality records a mortality event and updates batch count.
func (s *Service) RecordMortality(ctx context.Context, tankID int64, count int, cause, note string) error {
	if count <= 0 {
		return &model.ValidationError{Field: "count", Message: "count must be positive"}
	}
	if cause == "" {
		return &model.ValidationError{Field: "cause", Message: "cause is required"}
	}

	// Verify tank exists
	tank, err := s.repo.Tanks().GetByID(tankID)
	if err != nil {
		return fmt.Errorf("check tank: %w", err)
	}
	if tank == nil {
		return model.ErrNotFound
	}

	if err := s.repo.Batches().RecordMortality(tankID, count, cause, note); err != nil {
		return fmt.Errorf("record mortality: %w", err)
	}

	// Update tank stock count
	newStock := tank.StockQty - count
	if newStock < 0 {
		newStock = 0
	}
	if err := s.repo.Tanks().UpdateStock(tankID, newStock); err != nil {
		return fmt.Errorf("update tank stock: %w", err)
	}

	return nil
}

// ListMortalityLogs returns mortality logs for a tank.
func (s *Service) ListMortalityLogs(ctx context.Context, tankID int64, limit int) ([]model.MortalityLog, error) {
	return s.repo.Batches().ListMortalityLogs(tankID, limit)
}

// RecordWaterChange logs a water change event.
func (s *Service) RecordWaterChange(ctx context.Context, w *model.WaterChange) (*model.WaterChange, error) {
	if w.Timestamp.IsZero() {
		w.Timestamp = time.Now()
	}

	if err := model.ValidateWaterChange(w); err != nil {
		return nil, err
	}

	// Verify tank exists
	tank, err := s.repo.Tanks().GetByID(w.TankID)
	if err != nil {
		return nil, fmt.Errorf("check tank: %w", err)
	}
	if tank == nil {
		return nil, model.ErrNotFound
	}

	if err := s.repo.Batches().RecordWaterChange(w); err != nil {
		return nil, fmt.Errorf("record water change: %w", err)
	}
	return w, nil
}

// ListWaterChanges returns water change logs for a tank.
func (s *Service) ListWaterChanges(ctx context.Context, tankID int64, limit int) ([]model.WaterChange, error) {
	return s.repo.Batches().ListWaterChanges(tankID, limit)
}

// GetDailySummary returns a daily summary for a tank.
func (s *Service) GetDailySummary(ctx context.Context, tankID int64, date time.Time) (*model.DailySummary, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	readings, err := s.repo.Readings().ListByTankAndTimeRange(tankID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("get readings: %w", err)
	}

	dateStr := startOfDay.Format("2006-01-02")
	feedLogs, err := s.repo.Feeds().ListFeedLogsByDate(tankID, dateStr)
	if err != nil {
		return nil, fmt.Errorf("get feed logs: %w", err)
	}

	mortalityLogs, err := s.repo.Batches().ListMortalityLogs(tankID, 100)
	if err != nil && err != model.ErrNotFound {
		return nil, fmt.Errorf("get mortality logs: %w", err)
	}

	mortality := 0
	for _, m := range mortalityLogs {
		if m.Timestamp.After(startOfDay) && m.Timestamp.Before(endOfDay) {
			mortality += m.Count
		}
	}

	summary := model.ComputeDailyStats(readings, feedLogs, mortality)
	summary.TankID = tankID
	return summary, nil
}

// GetSystemConfig returns a config value.
func (s *Service) GetSystemConfig(ctx context.Context, key string) (string, error) {
	return s.repo.Config().Get(key)
}

// SetSystemConfig sets a config value.
func (s *Service) SetSystemConfig(ctx context.Context, key, value string) error {
	return s.repo.Config().Set(key, value)
}

// ListSystemConfig returns all config entries.
func (s *Service) ListSystemConfig(ctx context.Context) ([]model.SystemConfig, error) {
	return s.repo.Config().List()
}
