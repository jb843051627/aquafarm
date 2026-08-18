package store

import (
	"database/sql"
	"time"

	"aquafarm/internal/model"
)

// ReadingStore manages sensor readings in the database.
type ReadingStore struct {
	db *sql.DB
}

func NewReadingStore(db *sql.DB) *ReadingStore {
	return &ReadingStore{db: db}
}

// Create inserts a new sensor reading.
func (s *ReadingStore) Create(r *model.SensorReading) error {
	res, err := s.db.Exec(
		`INSERT INTO sensor_readings (tank_id, type, value, unit, timestamp)
		 VALUES (?, ?, ?, ?, ?)`,
		r.TankID, r.Type, r.Value, r.Unit, r.Timestamp,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = id
	return nil
}

// BatchInsert inserts multiple readings in a single transaction.
func (s *ReadingStore) BatchInsert(readings []model.SensorReading) error {
	if len(readings) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(
		`INSERT INTO sensor_readings (tank_id, type, value, unit, timestamp) VALUES (?, ?, ?, ?, ?)`,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for i := range readings {
		_, err := stmt.Exec(readings[i].TankID, readings[i].Type, readings[i].Value, readings[i].Unit, readings[i].Timestamp)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// GetByID retrieves a reading by ID.
func (s *ReadingStore) GetByID(id int64) (*model.SensorReading, error) {
	row := s.db.QueryRow(
		`SELECT id, tank_id, type, value, unit, timestamp FROM sensor_readings WHERE id = ?`, id,
	)
	r := &model.SensorReading{}
	err := row.Scan(&r.ID, &r.TankID, &r.Type, &r.Value, &r.Unit, &r.Timestamp)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ListByTank returns readings for a tank, optionally limited.
func (s *ReadingStore) ListByTank(tankID int64, limit int) ([]model.SensorReading, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT id, tank_id, type, value, unit, timestamp
		 FROM sensor_readings WHERE tank_id = ? ORDER BY timestamp DESC LIMIT ?`,
		tankID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []model.SensorReading
	for rows.Next() {
		var r model.SensorReading
		if err := rows.Scan(&r.ID, &r.TankID, &r.Type, &r.Value, &r.Unit, &r.Timestamp); err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	return readings, rows.Err()
}

// ListByTankAndType returns readings for a specific sensor type on a tank.
func (s *ReadingStore) ListByTankAndType(tankID int64, sensorType string, limit int) ([]model.SensorReading, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT id, tank_id, type, value, unit, timestamp
		 FROM sensor_readings WHERE tank_id = ? AND type = ?
		 ORDER BY timestamp DESC LIMIT ?`,
		tankID, sensorType, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []model.SensorReading
	for rows.Next() {
		var r model.SensorReading
		if err := rows.Scan(&r.ID, &r.TankID, &r.Type, &r.Value, &r.Unit, &r.Timestamp); err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	return readings, rows.Err()
}

// ListByTimeRange returns all readings within a time range.
func (s *ReadingStore) ListByTimeRange(start, end time.Time, limit int) ([]model.SensorReading, error) {
	if limit <= 0 {
		limit = 1000
	}
	rows, err := s.db.Query(
		`SELECT id, tank_id, type, value, unit, timestamp
		 FROM sensor_readings WHERE timestamp >= ? AND timestamp < ?
		 ORDER BY timestamp ASC LIMIT ?`,
		start, end, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []model.SensorReading
	for rows.Next() {
		var r model.SensorReading
		if err := rows.Scan(&r.ID, &r.TankID, &r.Type, &r.Value, &r.Unit, &r.Timestamp); err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	return readings, rows.Err()
}

// GetLatestByTankAndType returns the most recent reading for a tank+sensor combo.
func (s *ReadingStore) GetLatestByTankAndType(tankID int64, sensorType string) (*model.SensorReading, error) {
	row := s.db.QueryRow(
		`SELECT id, tank_id, type, value, unit, timestamp
		 FROM sensor_readings WHERE tank_id = ? AND type = ?
		 ORDER BY timestamp DESC LIMIT 1`,
		tankID, sensorType,
	)
	r := &model.SensorReading{}
	err := row.Scan(&r.ID, &r.TankID, &r.Type, &r.Value, &r.Unit, &r.Timestamp)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return r, nil
}

// GetAvgByTankAndType returns average value for a tank+sensor within a time range.
func (s *ReadingStore) GetAvgByTankAndType(tankID int64, sensorType string, since time.Time) (float64, error) {
	var avg float64
	err := s.db.QueryRow(
		`SELECT COALESCE(AVG(value), 0) FROM sensor_readings
		 WHERE tank_id = ? AND type = ? AND timestamp >= ?`,
		tankID, sensorType, since,
	).Scan(&avg)
	return avg, err
}

// DeleteOlderThan removes readings older than the given time.
func (s *ReadingStore) DeleteOlderThan(t time.Time) (int64, error) {
	res, err := s.db.Exec(`DELETE FROM sensor_readings WHERE timestamp < ?`, t)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// CountByTank returns reading counts grouped by tank.
func (s *ReadingStore) CountByTank() (map[int64]int, error) {
	rows, err := s.db.Query(
		`SELECT tank_id, COUNT(*) as cnt FROM sensor_readings GROUP BY tank_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]int)
	for rows.Next() {
		var tankID int64
		var cnt int
		if err := rows.Scan(&tankID, &cnt); err != nil {
			return nil, err
		}
		result[tankID] = cnt
	}
	return result, rows.Err()
}

// ListByTankAndTimeRange returns readings for a tank within a time range.
func (s *ReadingStore) ListByTankAndTimeRange(tankID int64, start, end time.Time) ([]model.SensorReading, error) {
	rows, err := s.db.Query(
		`SELECT id, tank_id, type, value, unit, timestamp
		 FROM sensor_readings WHERE tank_id = ? AND timestamp >= ? AND timestamp < ?
		 ORDER BY timestamp ASC`,
		tankID, start, end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []model.SensorReading
	for rows.Next() {
		var r model.SensorReading
		if err := rows.Scan(&r.ID, &r.TankID, &r.Type, &r.Value, &r.Unit, &r.Timestamp); err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	return readings, rows.Err()
}

// GetStatsByTank returns min/max/avg for each sensor type on a tank.
func (s *ReadingStore) GetStatsByTank(tankID int64, since time.Time) (map[string]model.DailySummary, error) {
	rows, err := s.db.Query(
		`SELECT type,
		        MIN(value) as min_val,
		        MAX(value) as max_val,
		        AVG(value) as avg_val,
		        COUNT(*) as cnt
		 FROM sensor_readings
		 WHERE tank_id = ? AND timestamp >= ?
		 GROUP BY type`,
		tankID, since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]model.DailySummary)
	for rows.Next() {
		var sensorType string
		var minVal, maxVal, avgVal float64
		var cnt int
		if err := rows.Scan(&sensorType, &minVal, &maxVal, &avgVal, &cnt); err != nil {
			return nil, err
		}
		result[sensorType] = model.DailySummary{
			TankID:     tankID,
			AvgTemp:    avgVal,
			MinOxygen:  minVal,
			MaxAmmonia: maxVal,
		}
	}
	return result, rows.Err()
}

// DeleteByTank removes all readings for a tank.
func (s *ReadingStore) DeleteByTank(tankID int64) (int64, error) {
	res, err := s.db.Exec(`DELETE FROM sensor_readings WHERE tank_id = ?`, tankID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
