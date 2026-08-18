package store

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

// Repo wraps the database connection and provides access to all stores.
type Repo struct {
	db       *sql.DB
	tanks    *TankStore
	readings *ReadingStore
	alerts   *AlertStore
	feeds    *FeedStore
	equip    *EquipmentStore
	config   *ConfigStore
	batch    *BatchStore
	mu       sync.RWMutex
}

// Open opens a SQLite database at the given path (file mode, NOT :memory:).
func Open(path string) (*Repo, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	dsn := fmt.Sprintf("file:%s?_journalmode=WAL&_busytimeout=5000", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	// Verify connectivity
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	r := &Repo{db: db}
	r.tanks = NewTankStore(db)
	r.readings = NewReadingStore(db)
	r.alerts = NewAlertStore(db)
	r.feeds = NewFeedStore(db)
	r.equip = NewEquipmentStore(db)
	r.config = NewConfigStore(db)
	r.batch = NewBatchStore(db)
	return r, nil
}

// Close closes the database connection.
func (r *Repo) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// Migrate creates all tables if they don't exist.
func (r *Repo) Migrate() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	migrations := []string{
		migrateTanks,
		migrateReadings,
		migrateThresholds,
		migrateAlerts,
		migrateFeedPlans,
		migrateFeedLogs,
		migrateEquipment,
		migrateMaintenanceTasks,
		migrateWaterChanges,
		migrateMortalityLogs,
		migrateBatches,
		migrateSystemConfig,
	}

	for _, m := range migrations {
		if _, err := r.db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}

// DB exposes the underlying *sql.DB (for testing only).
func (r *Repo) DB() *sql.DB {
	return r.db
}

// Tanks returns the tank store.
func (r *Repo) Tanks() *TankStore { return r.tanks }

// Readings returns the reading store.
func (r *Repo) Readings() *ReadingStore { return r.readings }

// Alerts returns the alert store.
func (r *Repo) Alerts() *AlertStore { return r.alerts }

// Feeds returns the feed store.
func (r *Repo) Feeds() *FeedStore { return r.feeds }

// Equipment returns the equipment store.
func (r *Repo) Equipment() *EquipmentStore { return r.equip }

// Config returns the system config store.
func (r *Repo) Config() *ConfigStore { return r.config }

// Batches returns the batch store.
func (r *Repo) Batches() *BatchStore { return r.batch }

const migrateTanks = `
CREATE TABLE IF NOT EXISTS tanks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    species TEXT NOT NULL,
    capacity REAL NOT NULL,
    stock_qty INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

const migrateReadings = `
CREATE TABLE IF NOT EXISTS sensor_readings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tank_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    value REAL NOT NULL,
    unit TEXT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tank_id) REFERENCES tanks(id)
);
CREATE INDEX IF NOT EXISTS idx_readings_tank_type ON sensor_readings(tank_id, type);
CREATE INDEX IF NOT EXISTS idx_readings_timestamp ON sensor_readings(timestamp);`

const migrateThresholds = `
CREATE TABLE IF NOT EXISTS threshold_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tank_id INTEGER NOT NULL,
    sensor_type TEXT NOT NULL,
    min_value REAL NOT NULL,
    max_value REAL NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (tank_id) REFERENCES tanks(id)
);`

const migrateAlerts = `
CREATE TABLE IF NOT EXISTS alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tank_id INTEGER NOT NULL,
    sensor_type TEXT NOT NULL,
    value REAL NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    resolved INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    FOREIGN KEY (tank_id) REFERENCES tanks(id)
);
CREATE INDEX IF NOT EXISTS idx_alerts_tank ON alerts(tank_id);
CREATE INDEX IF NOT EXISTS idx_alerts_unresolved ON alerts(resolved);`

const migrateFeedPlans = `
CREATE TABLE IF NOT EXISTS feed_plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tank_id INTEGER NOT NULL,
    feed_type TEXT NOT NULL,
    amount REAL NOT NULL,
    schedule TEXT NOT NULL,
    active INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tank_id) REFERENCES tanks(id)
);`

const migrateFeedLogs = `
CREATE TABLE IF NOT EXISTS feed_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tank_id INTEGER NOT NULL,
    plan_id INTEGER,
    feed_type TEXT NOT NULL,
    amount REAL NOT NULL,
    source TEXT NOT NULL DEFAULT 'manual',
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tank_id) REFERENCES tanks(id)
);
CREATE INDEX IF NOT EXISTS idx_feed_logs_tank ON feed_logs(tank_id);`

const migrateEquipment = `
CREATE TABLE IF NOT EXISTS equipment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tank_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    power_rating REAL NOT NULL DEFAULT 0,
    installed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_service DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tank_id) REFERENCES tanks(id)
);`

const migrateMaintenanceTasks = `
CREATE TABLE IF NOT EXISTS maintenance_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    equipment_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    description TEXT NOT NULL,
    scheduled_at DATETIME NOT NULL,
    completed_at DATETIME,
    status TEXT NOT NULL DEFAULT 'pending',
    technician TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (equipment_id) REFERENCES equipment(id)
);`

const migrateWaterChanges = `
CREATE TABLE IF NOT EXISTS water_changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tank_id INTEGER NOT NULL,
    volume REAL NOT NULL,
    reason TEXT NOT NULL,
    operator TEXT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tank_id) REFERENCES tanks(id)
);`

const migrateMortalityLogs = `
CREATE TABLE IF NOT EXISTS mortality_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tank_id INTEGER NOT NULL,
    count INTEGER NOT NULL,
    cause TEXT NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tank_id) REFERENCES tanks(id)
);`

const migrateBatches = `
CREATE TABLE IF NOT EXISTS batches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tank_id INTEGER NOT NULL,
    species TEXT NOT NULL,
    initial_count INTEGER NOT NULL,
    current_count INTEGER NOT NULL,
    arrival_date DATETIME NOT NULL,
    source TEXT NOT NULL,
    FOREIGN KEY (tank_id) REFERENCES tanks(id)
);`

const migrateSystemConfig = `
CREATE TABLE IF NOT EXISTS system_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);`
