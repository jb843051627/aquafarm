package store

import (
	"database/sql"
	"time"

	"aquafarm/internal/model"
)

// AlertStore manages alert records.
type AlertStore struct {
	db *sql.DB
}

func NewAlertStore(db *sql.DB) *AlertStore {
	return &AlertStore{db: db}
}

// Create inserts a new alert.
func (s *AlertStore) Create(a *model.Alert) error {
	res, err := s.db.Exec(
		`INSERT INTO alerts (tank_id, sensor_type, value, severity, message, resolved, created_at, resolved_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.TankID, a.SensorType, a.Value, a.Severity, a.Message, boolToInt(a.Resolved), a.CreatedAt, a.ResolvedAt,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	a.ID = id
	return nil
}

// GetByID retrieves an alert by ID.
func (s *AlertStore) GetByID(id int64) (*model.Alert, error) {
	row := s.db.QueryRow(
		`SELECT id, tank_id, sensor_type, value, severity, message, resolved, created_at, resolved_at
		 FROM alerts WHERE id = ?`, id,
	)
	a := &model.Alert{}
	var resolved int
	var resolvedAt sql.NullTime
	err := row.Scan(&a.ID, &a.TankID, &a.SensorType, &a.Value, &a.Severity, &a.Message, &resolved, &a.CreatedAt, &resolvedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	a.Resolved = resolved == 1
	if resolvedAt.Valid {
		a.ResolvedAt = &resolvedAt.Time
	}
	return a, nil
}

// List returns all alerts, optionally limited.
func (s *AlertStore) List(limit int) ([]model.Alert, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT id, tank_id, sensor_type, value, severity, message, resolved, created_at, resolved_at
		 FROM alerts ORDER BY created_at DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []model.Alert
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// ListByTank returns alerts for a tank.
func (s *AlertStore) ListByTank(tankID int64, limit int) ([]model.Alert, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT id, tank_id, sensor_type, value, severity, message, resolved, created_at, resolved_at
		 FROM alerts WHERE tank_id = ? ORDER BY created_at DESC LIMIT ?`, tankID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []model.Alert
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// ListUnresolved returns all unresolved alerts.
func (s *AlertStore) ListUnresolved(limit int) ([]model.Alert, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT id, tank_id, sensor_type, value, severity, message, resolved, created_at, resolved_at
		 FROM alerts WHERE resolved = 0 ORDER BY created_at DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []model.Alert
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// Resolve marks an alert as resolved.
func (s *AlertStore) Resolve(id int64) error {
	now := time.Now()
	_, err := s.db.Exec(
		`UPDATE alerts SET resolved = 1, resolved_at = ? WHERE id = ?`, now, id,
	)
	return err
}

// ResolveByTank resolves all unresolved alerts for a tank.
func (s *AlertStore) ResolveByTank(tankID int64) (int64, error) {
	now := time.Now()
	res, err := s.db.Exec(
		`UPDATE alerts SET resolved = 1, resolved_at = ? WHERE tank_id = ? AND resolved = 0`, now, tankID,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ListByTimeRange returns alerts within a time range.
func (s *AlertStore) ListByTimeRange(start, end time.Time, limit int) ([]model.Alert, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT id, tank_id, sensor_type, value, severity, message, resolved, created_at, resolved_at
		 FROM alerts WHERE created_at >= ? AND created_at < ?
		 ORDER BY created_at DESC LIMIT ?`, start, end, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []model.Alert
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// CountUnresolved returns count of unresolved alerts.
func (s *AlertStore) CountUnresolved() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM alerts WHERE resolved = 0`).Scan(&count)
	return count, err
}

// CountUnresolvedByTank returns count of unresolved alerts for a tank.
func (s *AlertStore) CountUnresolvedByTank(tankID int64) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM alerts WHERE resolved = 0 AND tank_id = ?`, tankID).Scan(&count)
	return count, err
}

// CountBySeverity returns alert counts grouped by severity.
func (s *AlertStore) CountBySeverity() (map[string]int, error) {
	rows, err := s.db.Query(
		`SELECT severity, COUNT(*) as cnt FROM alerts WHERE resolved = 0 GROUP BY severity`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var severity string
		var cnt int
		if err := rows.Scan(&severity, &cnt); err != nil {
			return nil, err
		}
		result[severity] = cnt
	}
	return result, rows.Err()
}

// DeleteByTank removes all alerts for a tank.
func (s *AlertStore) DeleteByTank(tankID int64) (int64, error) {
	res, err := s.db.Exec(`DELETE FROM alerts WHERE tank_id = ?`, tankID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func scanAlert(rows *sql.Rows) (model.Alert, error) {
	a := model.Alert{}
	var resolved int
	var resolvedAt sql.NullTime
	err := rows.Scan(&a.ID, &a.TankID, &a.SensorType, &a.Value, &a.Severity, &a.Message, &resolved, &a.CreatedAt, &resolvedAt)
	if err != nil {
		return a, err
	}
	a.Resolved = resolved == 1
	if resolvedAt.Valid {
		a.ResolvedAt = &resolvedAt.Time
	}
	return a, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
