package store

import (
	"database/sql"

	"aquafarm/internal/model"
)

// FeedStore manages feed plans and feed logs.
type FeedStore struct {
	db *sql.DB
}

func NewFeedStore(db *sql.DB) *FeedStore {
	return &FeedStore{db: db}
}

// CreatePlan inserts a new feed plan.
func (s *FeedStore) CreatePlan(f *model.FeedPlan) error {
	res, err := s.db.Exec(
		`INSERT INTO feed_plans (tank_id, feed_type, amount, schedule, active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		f.TankID, f.FeedType, f.Amount, f.Schedule, boolToInt(f.Active), f.CreatedAt, f.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	f.ID = id
	return nil
}

// GetPlanByID retrieves a feed plan by ID.
func (s *FeedStore) GetPlanByID(id int64) (*model.FeedPlan, error) {
	row := s.db.QueryRow(
		`SELECT id, tank_id, feed_type, amount, schedule, active, created_at, updated_at
		 FROM feed_plans WHERE id = ?`, id,
	)
	f := &model.FeedPlan{}
	var activeInt int
	err := row.Scan(&f.ID, &f.TankID, &f.FeedType, &f.Amount, &f.Schedule, &activeInt, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	f.Active = activeInt == 1
	return f, nil
}

// ListPlans returns all feed plans.
func (s *FeedStore) ListPlans() ([]model.FeedPlan, error) {
	rows, err := s.db.Query(
		`SELECT id, tank_id, feed_type, amount, schedule, active, created_at, updated_at
		 FROM feed_plans ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []model.FeedPlan
	for rows.Next() {
		var f model.FeedPlan
		var activeInt int
		if err := rows.Scan(&f.ID, &f.TankID, &f.FeedType, &f.Amount, &f.Schedule, &activeInt, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.Active = activeInt == 1
		plans = append(plans, f)
	}
	return plans, rows.Err()
}

// ListPlansByTank returns feed plans for a tank.
func (s *FeedStore) ListPlansByTank(tankID int64) ([]model.FeedPlan, error) {
	rows, err := s.db.Query(
		`SELECT id, tank_id, feed_type, amount, schedule, active, created_at, updated_at
		 FROM feed_plans WHERE tank_id = ? ORDER BY id`, tankID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []model.FeedPlan
	for rows.Next() {
		var f model.FeedPlan
		var activeInt int
		if err := rows.Scan(&f.ID, &f.TankID, &f.FeedType, &f.Amount, &f.Schedule, &activeInt, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.Active = activeInt == 1
		plans = append(plans, f)
	}
	return plans, rows.Err()
}

// ListActivePlans returns active feed plans.
func (s *FeedStore) ListActivePlans() ([]model.FeedPlan, error) {
	rows, err := s.db.Query(
		`SELECT id, tank_id, feed_type, amount, schedule, active, created_at, updated_at
		 FROM feed_plans WHERE active = 1 ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []model.FeedPlan
	for rows.Next() {
		var f model.FeedPlan
		var activeInt int
		if err := rows.Scan(&f.ID, &f.TankID, &f.FeedType, &f.Amount, &f.Schedule, &activeInt, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.Active = activeInt == 1
		plans = append(plans, f)
	}
	return plans, rows.Err()
}

// UpdatePlan updates a feed plan.
func (s *FeedStore) UpdatePlan(f *model.FeedPlan) error {
	_, err := s.db.Exec(
		`UPDATE feed_plans SET tank_id=?, feed_type=?, amount=?, schedule=?, active=?, updated_at=?
		 WHERE id=?`,
		f.TankID, f.FeedType, f.Amount, f.Schedule, boolToInt(f.Active), f.UpdatedAt, f.ID,
	)
	return err
}

// DeactivatePlan deactivates a feed plan.
func (s *FeedStore) DeactivatePlan(id int64) error {
	_, err := s.db.Exec(`UPDATE feed_plans SET active = 0 WHERE id = ?`, id)
	return err
}

// DeletePlan removes a feed plan.
func (s *FeedStore) DeletePlan(id int64) error {
	_, err := s.db.Exec(`DELETE FROM feed_plans WHERE id = ?`, id)
	return err
}

// LogFeed inserts a feed log entry.
func (s *FeedStore) LogFeed(fl *model.FeedLog) error {
	res, err := s.db.Exec(
		`INSERT INTO feed_logs (tank_id, plan_id, feed_type, amount, source, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		fl.TankID, fl.PlanID, fl.FeedType, fl.Amount, fl.Source, fl.Timestamp,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	fl.ID = id
	return nil
}

// ListFeedLogs returns feed logs for a tank.
func (s *FeedStore) ListFeedLogs(tankID int64, limit int) ([]model.FeedLog, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT id, tank_id, plan_id, feed_type, amount, source, timestamp
		 FROM feed_logs WHERE tank_id = ? ORDER BY timestamp DESC LIMIT ?`, tankID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.FeedLog
	for rows.Next() {
		var l model.FeedLog
		if err := rows.Scan(&l.ID, &l.TankID, &l.PlanID, &l.FeedType, &l.Amount, &l.Source, &l.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// GetFeedTotalByTank returns total feed amount for a tank.
func (s *FeedStore) GetFeedTotalByTank(tankID int64) (float64, error) {
	var total float64
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(amount), 0) FROM feed_logs WHERE tank_id = ?`, tankID,
	).Scan(&total)
	return total, err
}

// GetFeedTotalByDate returns total feed amount for a tank on a specific date.
func (s *FeedStore) GetFeedTotalByDate(tankID int64, date string) (float64, error) {
	var total float64
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(amount), 0) FROM feed_logs
		 WHERE tank_id = ? AND substr(timestamp, 1, 10) = ?`, tankID, date,
	).Scan(&total)
	return total, err
}

// ListFeedLogsByDate returns feed logs for a tank on a specific date.
func (s *FeedStore) ListFeedLogsByDate(tankID int64, date string) ([]model.FeedLog, error) {
	rows, err := s.db.Query(
		`SELECT id, tank_id, plan_id, feed_type, amount, source, timestamp
		 FROM feed_logs WHERE tank_id = ? AND substr(timestamp, 1, 10) = ?
		 ORDER BY timestamp ASC`, tankID, date,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.FeedLog
	for rows.Next() {
		var l model.FeedLog
		if err := rows.Scan(&l.ID, &l.TankID, &l.PlanID, &l.FeedType, &l.Amount, &l.Source, &l.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// DeleteFeedLogsByTank removes all feed logs for a tank.
func (s *FeedStore) DeleteFeedLogsByTank(tankID int64) (int64, error) {
	res, err := s.db.Exec(`DELETE FROM feed_logs WHERE tank_id = ?`, tankID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
