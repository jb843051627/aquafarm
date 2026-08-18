package model

import (
	"fmt"
	"time"
)

// Tank represents a fish tank in the recirculating aquaculture system.
type Tank struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Species   string    `json:"species"`
	Capacity  float64   `json:"capacity"`  // liters
	StockQty  int       `json:"stock_qty"`  // number of fish
	Status    string    `json:"status"`     // active, maintenance, idle
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SensorReading represents a water quality measurement from a sensor.
type SensorReading struct {
	ID        int64     `json:"id"`
	TankID    int64     `json:"tank_id"`
	Type      string    `json:"type"`  // temperature, ph, oxygen, ammonia, nitrite
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

// ThresholdConfig defines alert thresholds for a tank's sensor type.
type ThresholdConfig struct {
	ID        int64  `json:"id"`
	TankID    int64  `json:"tank_id"`
	SensorType string `json:"sensor_type"`
	MinValue  float64 `json:"min_value"`
	MaxValue  float64 `json:"max_value"`
	Enabled   bool   `json:"enabled"`
}

// Alert represents a threshold violation event.
type Alert struct {
	ID        int64     `json:"id"`
	TankID    int64     `json:"tank_id"`
	SensorType string   `json:"sensor_type"`
	Value     float64   `json:"value"`
	Severity  string    `json:"severity"`  // warning, critical
	Message   string    `json:"message"`
	Resolved  bool      `json:"resolved"`
	CreatedAt time.Time `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// FeedPlan defines a feeding schedule for a tank.
type FeedPlan struct {
	ID        int64     `json:"id"`
	TankID    int64     `json:"tank_id"`
	FeedType  string    `json:"feed_type"`
	Amount    float64   `json:"amount"`   // grams per feeding
	Schedule  string    `json:"schedule"` // cron-like: "08:00,16:00"
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FeedLog records each actual feeding event.
type FeedLog struct {
	ID        int64     `json:"id"`
	TankID    int64     `json:"tank_id"`
	PlanID    int64     `json:"plan_id"`
	FeedType  string    `json:"feed_type"`
	Amount    float64   `json:"amount"`
	Source    string    `json:"source"` // auto, manual
	Timestamp time.Time `json:"timestamp"`
}

// Equipment represents a piece of equipment in the system.
type Equipment struct {
	ID          int64     `json:"id"`
	TankID      int64     `json:"tank_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // pump, filter, aerator, heater
	Status      string    `json:"status"` // running, stopped, fault
	PowerRating float64   `json:"power_rating"` // watts
	InstalledAt time.Time `json:"installed_at"`
	LastService time.Time `json:"last_service"`
}

// MaintenanceTask represents a scheduled maintenance job.
type MaintenanceTask struct {
	ID          int64     `json:"id"`
	EquipmentID int64     `json:"equipment_id"`
	Type        string    `json:"type"` // cleaning, inspection, replacement
	Description string    `json:"description"`
	ScheduledAt time.Time `json:"scheduled_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Status      string    `json:"status"` // pending, in_progress, completed, overdue
	Technician  string    `json:"technician"`
}

// WaterChange records a water exchange event.
type WaterChange struct {
	ID         int64     `json:"id"`
	TankID     int64     `json:"tank_id"`
	Volume     float64   `json:"volume"`      // liters changed
	Reason     string    `json:"reason"`
	Operator   string    `json:"operator"`
	Timestamp  time.Time `json:"timestamp"`
}

// MortalityLog records fish mortality events.
type MortalityLog struct {
	ID        int64     `json:"id"`
	TankID    int64     `json:"tank_id"`
	Count     int       `json:"count"`
	Cause     string    `json:"cause"`
	Note      string    `json:"note"`
	Timestamp time.Time `json:"timestamp"`
}

// Batch represents a group of fish introduced together.
type Batch struct {
	ID           int64     `json:"id"`
	TankID       int64     `json:"tank_id"`
	Species      string    `json:"species"`
	InitialCount int       `json:"initial_count"`
	CurrentCount int       `json:"current_count"`
	ArrivalDate  time.Time `json:"arrival_date"`
	Source       string    `json:"source"` // hatchery name
}

// SystemConfig stores key-value system configuration.
type SystemConfig struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DailySummary is a computed aggregation for a tank's daily stats.
type DailySummary struct {
	TankID    int64   `json:"tank_id"`
	Date      string  `json:"date"`
	AvgTemp   float64 `json:"avg_temp"`
	AvgPH     float64 `json:"avg_ph"`
	MinOxygen float64 `json:"min_oxygen"`
	MaxAmmonia float64 `json:"max_ammonia"`
	FeedTotal  float64 `json:"feed_total"`
	Mortality  int     `json:"mortality"`
}

// Alert severity constants
const (
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)

// Tank status constants
const (
	TankStatusActive     = "active"
	TankStatusMaintenance = "maintenance"
	TankStatusIdle       = "idle"
)

// Equipment status constants
const (
	EquipmentRunning = "running"
	EquipmentStopped = "stopped"
	EquipmentFault   = "fault"
)

// Task status constants
const (
	TaskPending    = "pending"
	TaskInProgress = "in_progress"
	TaskCompleted  = "completed"
	TaskOverdue    = "overdue"
)

// Feed source constants
const (
	FeedSourceAuto   = "auto"
	FeedSourceManual = "manual"
)

// Sensor type constants
const (
	SensorTemperature = "temperature"
	SensorPH          = "ph"
	SensorOxygen      = "oxygen"
	SensorAmmonia     = "ammonia"
	SensorNitrite     = "nitrite"
)

// ErrNotFound is a sentinel error for entity-not-found.
var ErrNotFound = fmt.Errorf("record not found")

// Validation helpers
func IsValidSensorType(t string) bool {
	switch t {
	case SensorTemperature, SensorPH, SensorOxygen, SensorAmmonia, SensorNitrite:
		return true
	}
	return false
}

func IsValidTankStatus(s string) bool {
	switch s {
	case TankStatusActive, TankStatusMaintenance, TankStatusIdle:
		return true
	}
	return false
}

func IsValidEquipmentStatus(s string) bool {
	switch s {
	case EquipmentRunning, EquipmentStopped, EquipmentFault:
		return true
	}
	return false
}
