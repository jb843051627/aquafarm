package store

import (
	"database/sql"
	"time"

	"aquafarm/internal/model"
)

// ConfigStore manages system configuration key-value pairs.
type ConfigStore struct {
	db *sql.DB
}

func NewConfigStore(db *sql.DB) *ConfigStore {
	return &ConfigStore{db: db}
}

// Get retrieves a config value by key.
func (s *ConfigStore) Get(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM system_config WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", model.ErrNotFound
	}
	return value, err
}

// Set inserts or updates a config value.
func (s *ConfigStore) Set(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO system_config (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = ?`,
		key, value, value,
	)
	return err
}

// List returns all config entries.
func (s *ConfigStore) List() ([]model.SystemConfig, error) {
	rows, err := s.db.Query(`SELECT key, value FROM system_config ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.SystemConfig
	for rows.Next() {
		var c model.SystemConfig
		if err := rows.Scan(&c.Key, &c.Value); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// Delete removes a config entry.
func (s *ConfigStore) Delete(key string) error {
	_, err := s.db.Exec(`DELETE FROM system_config WHERE key = ?`, key)
	return err
}

// BatchStore manages fish batch records.
type BatchStore struct {
	db *sql.DB
}

func NewBatchStore(db *sql.DB) *BatchStore {
	return &BatchStore{db: db}
}

// Create inserts a new batch.
func (s *BatchStore) Create(b *model.Batch) error {
	res, err := s.db.Exec(
		`INSERT INTO batches (tank_id, species, initial_count, current_count, arrival_date, source)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		b.TankID, b.Species, b.InitialCount, b.CurrentCount, b.ArrivalDate, b.Source,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	b.ID = id
	return nil
}

// GetByID retrieves a batch by ID.
func (s *BatchStore) GetByID(id int64) (*model.Batch, error) {
	row := s.db.QueryRow(
		`SELECT id, tank_id, species, initial_count, current_count, arrival_date, source
		 FROM batches WHERE id = ?`, id,
	)
	b := &model.Batch{}
	err := row.Scan(&b.ID, &b.TankID, &b.Species, &b.InitialCount, &b.CurrentCount, &b.ArrivalDate, &b.Source)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

// ListByTank returns batches for a tank.
func (s *BatchStore) ListByTank(tankID int64) ([]model.Batch, error) {
	rows, err := s.db.Query(
		`SELECT id, tank_id, species, initial_count, current_count, arrival_date, source
		 FROM batches WHERE tank_id = ? ORDER BY arrival_date DESC`, tankID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []model.Batch
	for rows.Next() {
		var b model.Batch
		if err := rows.Scan(&b.ID, &b.TankID, &b.Species, &b.InitialCount, &b.CurrentCount, &b.ArrivalDate, &b.Source); err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}
	return batches, rows.Err()
}

// List returns all batches.
func (s *BatchStore) List() ([]model.Batch, error) {
	rows, err := s.db.Query(
		`SELECT id, tank_id, species, initial_count, current_count, arrival_date, source
		 FROM batches ORDER BY arrival_date DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []model.Batch
	for rows.Next() {
		var b model.Batch
		if err := rows.Scan(&b.ID, &b.TankID, &b.Species, &b.InitialCount, &b.CurrentCount, &b.ArrivalDate, &b.Source); err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}
	return batches, rows.Err()
}

// UpdateCount updates the current count of a batch.
func (s *BatchStore) UpdateCount(id int64, count int) error {
	_, err := s.db.Exec(`UPDATE batches SET current_count = ? WHERE id = ?`, count, id)
	return err
}

// RecordMortality decreases the current count and logs the mortality event.
func (s *BatchStore) RecordMortality(tankID int64, count int, cause, note string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Insert mortality log
	_, err = tx.Exec(
		`INSERT INTO mortality_logs (tank_id, count, cause, note, timestamp)
		 VALUES (?, ?, ?, ?, ?)`, tankID, count, cause, note, time.Now(),
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Update batch current count
	_, err = tx.Exec(
		`UPDATE batches SET current_count = current_count - ? WHERE tank_id = ?`, count, tankID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ListMortalityLogs returns mortality logs for a tank.
func (s *BatchStore) ListMortalityLogs(tankID int64, limit int) ([]model.MortalityLog, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT id, tank_id, count, cause, note, timestamp
		 FROM mortality_logs WHERE tank_id = ? ORDER BY timestamp DESC LIMIT ?`, tankID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.MortalityLog
	for rows.Next() {
		var l model.MortalityLog
		if err := rows.Scan(&l.ID, &l.TankID, &l.Count, &l.Cause, &l.Note, &l.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// RecordWaterChange logs a water change event.
func (s *BatchStore) RecordWaterChange(w *model.WaterChange) error {
	res, err := s.db.Exec(
		`INSERT INTO water_changes (tank_id, volume, reason, operator, timestamp)
		 VALUES (?, ?, ?, ?, ?)`, w.TankID, w.Volume, w.Reason, w.Operator, w.Timestamp,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	w.ID = id
	return nil
}

// ListWaterChanges returns water change logs for a tank.
func (s *BatchStore) ListWaterChanges(tankID int64, limit int) ([]model.WaterChange, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT id, tank_id, volume, reason, operator, timestamp
		 FROM water_changes WHERE tank_id = ? ORDER BY timestamp DESC LIMIT ?`, tankID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []model.WaterChange
	for rows.Next() {
		var w model.WaterChange
		if err := rows.Scan(&w.ID, &w.TankID, &w.Volume, &w.Reason, &w.Operator, &w.Timestamp); err != nil {
			return nil, err
		}
		changes = append(changes, w)
	}
	return changes, rows.Err()
}

// SaveThreshold inserts or updates a threshold config.
func (s *ConfigStore) SaveThreshold(t *model.ThresholdConfig) error {
	res, err := s.db.Exec(
		`INSERT INTO threshold_configs (tank_id, sensor_type, min_value, max_value, enabled)
		 VALUES (?, ?, ?, ?, ?)`,
		t.TankID, t.SensorType, t.MinValue, t.MaxValue, boolToInt(t.Enabled),
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = id
	return nil
}

// GetThresholdsByTank returns threshold configs for a tank.
func (s *ConfigStore) GetThresholdsByTank(tankID int64) ([]model.ThresholdConfig, error) {
	rows, err := s.db.Query(
		`SELECT id, tank_id, sensor_type, min_value, max_value, enabled
		 FROM threshold_configs WHERE tank_id = ?`, tankID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.ThresholdConfig
	for rows.Next() {
		var t model.ThresholdConfig
		var enabledInt int
		if err := rows.Scan(&t.ID, &t.TankID, &t.SensorType, &t.MinValue, &t.MaxValue, &enabledInt); err != nil {
			return nil, err
		}
		t.Enabled = enabledInt == 1
		items = append(items, t)
	}
	return items, rows.Err()
}

// DeleteThreshold removes a threshold config.
func (s *ConfigStore) DeleteThreshold(id int64) error {
	_, err := s.db.Exec(`DELETE FROM threshold_configs WHERE id = ?`, id)
	return err
}
