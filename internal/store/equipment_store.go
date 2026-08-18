package store

import (
	"database/sql"
	"time"

	"aquafarm/internal/model"
)

// EquipmentStore manages equipment and maintenance task records.
type EquipmentStore struct {
	db *sql.DB
}

func NewEquipmentStore(db *sql.DB) *EquipmentStore {
	return &EquipmentStore{db: db}
}

// Create inserts a new equipment record.
func (s *EquipmentStore) Create(e *model.Equipment) error {
	res, err := s.db.Exec(
		`INSERT INTO equipment (tank_id, name, type, status, power_rating, installed_at, last_service)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.TankID, e.Name, e.Type, e.Status, e.PowerRating, e.InstalledAt, e.LastService,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

// GetByID retrieves equipment by ID.
func (s *EquipmentStore) GetByID(id int64) (*model.Equipment, error) {
	row := s.db.QueryRow(
		`SELECT id, tank_id, name, type, status, power_rating, installed_at, last_service
		 FROM equipment WHERE id = ?`, id,
	)
	e := &model.Equipment{}
	err := row.Scan(&e.ID, &e.TankID, &e.Name, &e.Type, &e.Status, &e.PowerRating, &e.InstalledAt, &e.LastService)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return e, nil
}

// ListByTank returns equipment for a tank.
func (s *EquipmentStore) ListByTank(tankID int64) ([]model.Equipment, error) {
	rows, err := s.db.Query(
		`SELECT id, tank_id, name, type, status, power_rating, installed_at, last_service
		 FROM equipment WHERE tank_id = ? ORDER BY id`, tankID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Equipment
	for rows.Next() {
		var e model.Equipment
		if err := rows.Scan(&e.ID, &e.TankID, &e.Name, &e.Type, &e.Status, &e.PowerRating, &e.InstalledAt, &e.LastService); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

// List returns all equipment.
func (s *EquipmentStore) List() ([]model.Equipment, error) {
	rows, err := s.db.Query(
		`SELECT id, tank_id, name, type, status, power_rating, installed_at, last_service
		 FROM equipment ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Equipment
	for rows.Next() {
		var e model.Equipment
		if err := rows.Scan(&e.ID, &e.TankID, &e.Name, &e.Type, &e.Status, &e.PowerRating, &e.InstalledAt, &e.LastService); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

// ListByStatus returns equipment filtered by status.
func (s *EquipmentStore) ListByStatus(status string) ([]model.Equipment, error) {
	rows, err := s.db.Query(
		`SELECT id, tank_id, name, type, status, power_rating, installed_at, last_service
		 FROM equipment WHERE status = ? ORDER BY id`, status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Equipment
	for rows.Next() {
		var e model.Equipment
		if err := rows.Scan(&e.ID, &e.TankID, &e.Name, &e.Type, &e.Status, &e.PowerRating, &e.InstalledAt, &e.LastService); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

// Update updates equipment fields.
func (s *EquipmentStore) Update(e *model.Equipment) error {
	_, err := s.db.Exec(
		`UPDATE equipment SET tank_id=?, name=?, type=?, status=?, power_rating=?, last_service=?
		 WHERE id=?`,
		e.TankID, e.Name, e.Type, e.Status, e.PowerRating, e.LastService, e.ID,
	)
	return err
}

// UpdateStatus updates only the status.
func (s *EquipmentStore) UpdateStatus(id int64, status string) error {
	_, err := s.db.Exec(`UPDATE equipment SET status = ? WHERE id = ?`, status, id)
	return err
}

// UpdateLastService updates the last service timestamp.
func (s *EquipmentStore) UpdateLastService(id int64, t time.Time) error {
	_, err := s.db.Exec(`UPDATE equipment SET last_service = ? WHERE id = ?`, t, id)
	return err
}

// Delete removes equipment by ID.
func (s *EquipmentStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM equipment WHERE id = ?`, id)
	return err
}

// CreateTask inserts a maintenance task.
func (s *EquipmentStore) CreateTask(t *model.MaintenanceTask) error {
	res, err := s.db.Exec(
		`INSERT INTO maintenance_tasks (equipment_id, type, description, scheduled_at, completed_at, status, technician)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.EquipmentID, t.Type, t.Description, t.ScheduledAt, t.CompletedAt, t.Status, t.Technician,
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

// GetTaskByID retrieves a maintenance task by ID.
func (s *EquipmentStore) GetTaskByID(id int64) (*model.MaintenanceTask, error) {
	row := s.db.QueryRow(
		`SELECT id, equipment_id, type, description, scheduled_at, completed_at, status, technician
		 FROM maintenance_tasks WHERE id = ?`, id,
	)
	t := &model.MaintenanceTask{}
	var completedAt sql.NullTime
	err := row.Scan(&t.ID, &t.EquipmentID, &t.Type, &t.Description, &t.ScheduledAt, &completedAt, &t.Status, &t.Technician)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	return t, nil
}

// ListTasks returns all maintenance tasks.
func (s *EquipmentStore) ListTasks(limit int) ([]model.MaintenanceTask, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT id, equipment_id, type, description, scheduled_at, completed_at, status, technician
		 FROM maintenance_tasks ORDER BY scheduled_at DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.MaintenanceTask
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ListTasksByEquipment returns tasks for a specific equipment.
func (s *EquipmentStore) ListTasksByEquipment(equipID int64) ([]model.MaintenanceTask, error) {
	rows, err := s.db.Query(
		`SELECT id, equipment_id, type, description, scheduled_at, completed_at, status, technician
		 FROM maintenance_tasks WHERE equipment_id = ? ORDER BY scheduled_at DESC`, equipID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.MaintenanceTask
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ListTasksByStatus returns tasks filtered by status.
func (s *EquipmentStore) ListTasksByStatus(status string) ([]model.MaintenanceTask, error) {
	rows, err := s.db.Query(
		`SELECT id, equipment_id, type, description, scheduled_at, completed_at, status, technician
		 FROM maintenance_tasks WHERE status = ? ORDER BY scheduled_at ASC`, status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.MaintenanceTask
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// CompleteTask marks a task as completed.
func (s *EquipmentStore) CompleteTask(id int64, technician string) error {
	now := time.Now()
	_, err := s.db.Exec(
		`UPDATE maintenance_tasks SET status = ?, completed_at = ?, technician = ? WHERE id = ?`,
		model.TaskCompleted, now, technician, id,
	)
	return err
}

// ListOverdueTasks returns tasks that are past their scheduled time and not completed.
func (s *EquipmentStore) ListOverdueTasks() ([]model.MaintenanceTask, error) {
	now := time.Now()
	rows, err := s.db.Query(
		`SELECT id, equipment_id, type, description, scheduled_at, completed_at, status, technician
		 FROM maintenance_tasks WHERE scheduled_at < ? AND status != ? AND status != ?
		 ORDER BY scheduled_at ASC`, now, model.TaskCompleted, model.TaskOverdue,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.MaintenanceTask
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func scanTask(rows *sql.Rows) (model.MaintenanceTask, error) {
	t := model.MaintenanceTask{}
	var completedAt sql.NullTime
	err := rows.Scan(&t.ID, &t.EquipmentID, &t.Type, &t.Description, &t.ScheduledAt, &completedAt, &t.Status, &t.Technician)
	if err != nil {
		return t, err
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	return t, nil
}
