package store

import (
	"database/sql"
	"fmt"

	"aquafarm/internal/model"
)

// Cleanup removes all data for a tank (cascading delete).
func (r *Repo) Cleanup(tankID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	tables := []string{
		`DELETE FROM sensor_readings WHERE tank_id = ?`,
		`DELETE FROM alerts WHERE tank_id = ?`,
		`DELETE FROM feed_plans WHERE tank_id = ?`,
		`DELETE FROM feed_logs WHERE tank_id = ?`,
		`DELETE FROM equipment WHERE tank_id = ?`,
		`DELETE FROM water_changes WHERE tank_id = ?`,
		`DELETE FROM mortality_logs WHERE tank_id = ?`,
		`DELETE FROM threshold_configs WHERE tank_id = ?`,
		`DELETE FROM batches WHERE tank_id = ?`,
		`DELETE FROM maintenance_tasks WHERE equipment_id IN (SELECT id FROM equipment WHERE tank_id = ?)`,
	}

	for _, q := range tables {
		if _, err := tx.Exec(q, tankID); err != nil {
			defer tx.Rollback()
			return fmt.Errorf("cleanup failed for tank %d: %w", tankID, err)
		}
	}

	if _, err := tx.Exec(`DELETE FROM tanks WHERE id = ?`, tankID); err != nil {
		defer tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetTankSummary returns aggregated stats for a tank.
type TankSummary struct {
	Tank          model.Tank
	ReadingCount  int
	AlertCount    int
	FeedTotal     float64
	EquipmentCount int
	ActiveBatch   *model.Batch
}

// GetTankSummary returns a comprehensive summary of a tank's state.
func (r *Repo) GetTankSummary(tankID int64) (*TankSummary, error) {
	tank, err := r.tanks.GetByID(tankID)
	if err != nil {
		return nil, err
	}
	if tank == nil {
		return nil, model.ErrNotFound
	}

	summary := &TankSummary{Tank: *tank}

	// Reading count
	readings, err := r.readings.ListByTank(tankID, 1)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}
	if readings != nil {
		// Get actual count
		var cnt int
		err = r.db.QueryRow(`SELECT COUNT(*) FROM sensor_readings WHERE tank_id = ?`, tankID).Scan(&cnt)
		if err != nil {
			return nil, err
		}
		summary.ReadingCount = cnt
	}

	// Unresolved alert count
	alertCount, err := r.alerts.CountUnresolved()
	if err != nil {
		return nil, err
	}
	summary.AlertCount = alertCount

	// Feed total
	feedTotal, err := r.feeds.GetFeedTotalByTank(tankID)
	if err != nil {
		return nil, err
	}
	summary.FeedTotal = feedTotal

	// Equipment count
	eqList, err := r.equip.ListByTank(tankID)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}
	summary.EquipmentCount = len(eqList)

	// Active batch
	batches, err := r.batch.ListByTank(tankID)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}
	for _, b := range batches {
		if b.CurrentCount > 0 {
			summary.ActiveBatch = &b
			break
		}
	}

	return summary, nil
}

// GetSystemOverview returns system-wide statistics.
type SystemOverview struct {
	TotalTanks      int
	TotalStock      int
	TotalCapacity   float64
	ActiveAlerts    int
	UnresolvedAlerts int
	TotalFeedToday  float64
	EquipmentRunning int
	EquipmentFault   int
	OverdueTasks     int
}

// GetSystemOverview aggregates system-wide metrics.
func (r *Repo) GetSystemOverview() (*SystemOverview, error) {
	overview := &SystemOverview{}

	tankCount, err := r.tanks.Count()
	if err != nil {
		return nil, err
	}
	overview.TotalTanks = tankCount

	stock, err := r.tanks.TotalStock()
	if err != nil {
		return nil, err
	}
	overview.TotalStock = stock

	capacity, err := r.tanks.TotalCapacity()
	if err != nil {
		return nil, err
	}
	overview.TotalCapacity = capacity

	unresolved, err := r.alerts.CountUnresolved()
	if err != nil {
		return nil, err
	}
	overview.UnresolvedAlerts = unresolved

	running, err := r.equip.ListByStatus(model.EquipmentRunning)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	overview.EquipmentRunning = len(running)

	faults, err := r.equip.ListByStatus(model.EquipmentFault)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	overview.EquipmentFault = len(faults)

	overdue, err := r.equip.ListOverdueTasks()
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	overview.OverdueTasks = len(overdue)

	return overview, nil
}

// SearchTanks finds tanks by name or species substring.
func (r *Repo) SearchTanks(query string) ([]model.Tank, error) {
	if query == "" {
		return r.tanks.List()
	}
	pattern := "%" + query + "%"
	rows, err := r.db.Query(
		`SELECT id, name, species, capacity, stock_qty, status, created_at, updated_at
		 FROM tanks WHERE name LIKE ? OR species LIKE ? ORDER BY id`,
		pattern, pattern,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tanks []model.Tank
	for rows.Next() {
		var t model.Tank
		if err := rows.Scan(&t.ID, &t.Name, &t.Species, &t.Capacity, &t.StockQty, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tanks = append(tanks, t)
	}
	return tanks, rows.Err()
}

// GetTanksNeedingAttention returns tanks with unresolved alerts or overdue equipment.
func (r *Repo) GetTanksNeedingAttention() ([]int64, error) {
	rows, err := r.db.Query(
		`SELECT DISTINCT a.tank_id FROM alerts a WHERE a.resolved = 0
		 UNION
		 SELECT DISTINCT e.tank_id FROM equipment e WHERE e.status = 'fault'`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tankIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		tankIDs = append(tankIDs, id)
	}
	return tankIDs, rows.Err()
}
