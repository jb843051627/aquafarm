package monitor

import (
	"context"
	"log"
	"time"

	"aquafarm/internal/service"
)

// Monitor runs periodic checks on the aquaculture system.
type Monitor struct {
	svc      *service.Service
	interval time.Duration
}

// New creates a new monitor.
func New(svc *service.Service) *Monitor {
	return &Monitor{
		svc:      svc,
		interval: 30 * time.Second,
	}
}

// Run starts the monitoring loop.
func (m *Monitor) Run(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("monitor stopped")
			return
		case <-ticker.C:
			m.checkOverdueTasks(ctx)
			m.checkStaleReadings(ctx)
			m.checkEquipmentHealth(ctx)
		}
	}
}

// checkOverdueTasks marks maintenance tasks as overdue if past their scheduled time.
func (m *Monitor) checkOverdueTasks(ctx context.Context) {
	overdue, err := m.svc.ListOverdueTasks(ctx)
	if err != nil {
		log.Printf("monitor: failed to list overdue tasks: %v", err)
		return
	}
	if len(overdue) > 0 {
		log.Printf("monitor: %d overdue maintenance tasks detected", len(overdue))
	}
}

// checkStaleReadings checks if any tanks haven't reported readings recently.
func (m *Monitor) checkStaleReadings(ctx context.Context) {
	tanks, err := m.svc.ListTanks(ctx)
	if err != nil {
		log.Printf("monitor: failed to list tanks: %v", err)
		return
	}

	cutoff := time.Now().Add(-1 * time.Hour)
	for _, tank := range tanks {
		if tank.Status != "active" {
			continue
		}

		// Check temperature sensor as a primary indicator
		readings, err := m.svc.ListReadingsByTankAndType(ctx, tank.ID, "temperature", 1)
		if err != nil {
			continue
		}
		if len(readings) > 0 {
			latest := readings[0]
			if latest.Timestamp.Before(cutoff) {
				log.Printf("monitor: tank %s (ID=%d) has stale readings since %s",
					tank.Name, tank.ID, latest.Timestamp.Format("2006-01-02 15:04"))
			}
		}
	}
}

// checkEquipmentHealth checks for faulted equipment.
func (m *Monitor) checkEquipmentHealth(ctx context.Context) {
	faulted, err := m.svc.ListEquipmentByStatus(ctx, "fault")
	if err != nil {
		log.Printf("monitor: failed to list faulted equipment: %v", err)
		return
	}
	if len(faulted) > 0 {
		for _, eq := range faulted {
			log.Printf("monitor: equipment %s (ID=%d) is in fault state", eq.Name, eq.ID)
		}
	}
}

// CheckAlerts scans for new alert conditions across all tanks.
func (m *Monitor) CheckAlerts(ctx context.Context) error {
	tanks, err := m.svc.ListTanks(ctx)
	if err != nil {
		return err
	}

	for _, tank := range tanks {
		if tank.Status != "active" {
			continue
		}

		// Check if there are unresolved alerts for this tank
		alerts, err := m.svc.ListAlertsByTank(ctx, tank.ID, 10)
		if err != nil {
			continue
		}

		for _, alert := range alerts {
			if !alert.Resolved && time.Since(alert.CreatedAt) > 30*time.Minute {
				log.Printf("monitor: unresolved alert on tank %s (ID=%d): %s [severity=%s]",
					tank.Name, tank.ID, alert.Message, alert.Severity)
			}
		}
	}
	return nil
}

// SetInterval changes the monitoring interval.
func (m *Monitor) SetInterval(d time.Duration) {
	m.interval = d
}

// GetInterval returns the current monitoring interval.
func (m *Monitor) GetInterval() time.Duration {
	return m.interval
}
