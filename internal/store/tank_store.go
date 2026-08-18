package store

import (
	"database/sql"
	"time"

	"aquafarm/internal/model"
)

// TankStore manages tank records in the database.
type TankStore struct {
	db *sql.DB
}

func NewTankStore(db *sql.DB) *TankStore {
	return &TankStore{db: db}
}

// Create inserts a new tank and returns it with the assigned ID.
func (s *TankStore) Create(t *model.Tank) error {
	res, err := s.db.Exec(
		`INSERT INTO tanks (name, species, capacity, stock_qty, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.Name, t.Species, t.Capacity, t.StockQty, t.Status, t.CreatedAt, t.UpdatedAt,
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

// GetByID retrieves a tank by ID. Returns nil, nil if not found (bug: should return error).
func (s *TankStore) GetByID(id int64) (*model.Tank, error) {
	row := s.db.QueryRow(
		`SELECT id, name, species, capacity, stock_qty, status, created_at, updated_at
		 FROM tanks WHERE id = ?`, id,
	)
	t := &model.Tank{}
	err := row.Scan(&t.ID, &t.Name, &t.Species, &t.Capacity, &t.StockQty, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // BUG-001: returns nil,nil instead of ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

// GetByName retrieves a tank by name.
func (s *TankStore) GetByName(name string) (*model.Tank, error) {
	row := s.db.QueryRow(
		`SELECT id, name, species, capacity, stock_qty, status, created_at, updated_at
		 FROM tanks WHERE name = ?`, name,
	)
	t := &model.Tank{}
	err := row.Scan(&t.ID, &t.Name, &t.Species, &t.Capacity, &t.StockQty, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

// List returns all tanks ordered by ID.
func (s *TankStore) List() ([]model.Tank, error) {
	rows, err := s.db.Query(
		`SELECT id, name, species, capacity, stock_qty, status, created_at, updated_at
		 FROM tanks ORDER BY id`,
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

// ListByStatus returns tanks filtered by status.
func (s *TankStore) ListByStatus(status string) ([]model.Tank, error) {
	rows, err := s.db.Query(
		`SELECT id, name, species, capacity, stock_qty, status, created_at, updated_at
		 FROM tanks WHERE status = ? ORDER BY id`, status,
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

// Update updates a tank's fields.
func (s *TankStore) Update(t *model.Tank) error {
	t.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE tanks SET name=?, species=?, capacity=?, stock_qty=?, status=?, updated_at=?
		 WHERE id=?`,
		t.Name, t.Species, t.Capacity, t.StockQty, t.Status, t.UpdatedAt, t.ID,
	)
	return err
}

// UpdateStock updates only the stock quantity.
func (s *TankStore) UpdateStock(id int64, stock int) error {
	_, err := s.db.Exec(
		`UPDATE tanks SET stock_qty=?, updated_at=? WHERE id=?`,
		stock, time.Now(), id,
	)
	return err
}

// Delete removes a tank by ID.
func (s *TankStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM tanks WHERE id=?`, id)
	return err
}

// Count returns total tank count.
func (s *TankStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM tanks`).Scan(&count)
	return count, err
}

// CountBySpecies returns the count of tanks for each species.
func (s *TankStore) CountBySpecies() (map[string]int, error) {
	rows, err := s.db.Query(
		`SELECT species, COUNT(*) as cnt FROM tanks GROUP BY species ORDER BY cnt DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var species string
		var cnt int
		if err := rows.Scan(&species, &cnt); err != nil {
			return nil, err
		}
		result[species] = cnt
	}
	return result, rows.Err()
}

// TotalStock returns the sum of all tank stock quantities.
func (s *TankStore) TotalStock() (int, error) {
	var total int
	err := s.db.QueryRow(`SELECT COALESCE(SUM(stock_qty), 0) FROM tanks`).Scan(&total)
	return total, err
}

// TotalCapacity returns the sum of all tank capacities.
func (s *TankStore) TotalCapacity() (float64, error) {
	var total float64
	err := s.db.QueryRow(`SELECT COALESCE(SUM(capacity), 0) FROM tanks`).Scan(&total)
	return total, err
}
