package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"aquafarm/internal/model"
	"aquafarm/internal/store"
)

// TestBug004_CompleteTaskErrorChain verifies that completing an already-completed task
// returns a *ValidationError that can be unwrapped via errors.As.
// Bug: returns fmt.Errorf instead of *ValidationError.
func TestBug004_CompleteTaskErrorChain(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	repo, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	defer repo.Close()

	if err := repo.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	svc := New(repo)

	// Create a tank
	tank := &model.Tank{
		Name:     "TestTank",
		Species:  "Salmon",
		Capacity: 1000,
		StockQty: 50,
		Status:   model.TankStatusActive,
	}
	if err := repo.Tanks().Create(tank); err != nil {
		t.Fatalf("create tank: %v", err)
	}

	// Create equipment
	eq := &model.Equipment{
		TankID:      tank.ID,
		Name:        "Pump1",
		Type:        "pump",
		Status:      model.EquipmentRunning,
		PowerRating: 100,
		InstalledAt: time.Now(),
		LastService: time.Now(),
	}
	if err := repo.Equipment().Create(eq); err != nil {
		t.Fatalf("create equipment: %v", err)
	}

	// Create a maintenance task
	task := &model.MaintenanceTask{
		EquipmentID: eq.ID,
		Type:        "cleaning",
		Description: "Regular cleaning",
		ScheduledAt: time.Now().Add(24 * time.Hour),
		Status:      model.TaskPending,
	}
	if err := repo.Equipment().CreateTask(task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	// Complete the task — should succeed
	if err := svc.CompleteMaintenanceTask(context.Background(), task.ID, "tech1"); err != nil {
		t.Fatalf("first complete failed: %v", err)
	}

	// Complete again — should return *ValidationError
	err = svc.CompleteMaintenanceTask(context.Background(), task.ID, "tech1")
	if err == nil {
		t.Fatal("expected error for already-completed task, got nil")
	}

	var ve *model.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}