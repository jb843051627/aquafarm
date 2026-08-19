package service

import (
	"context"
	"path/filepath"
	"testing"

	"aquafarm/internal/model"
	"aquafarm/internal/store"
)

// TestBug006_ComputeStockDensityNil verifies that ComputeStockDensity does not panic
// when passed a nil tank pointer.
// Bug: model.ComputeStockDensity lacks nil guard, causing nil dereference panic.
// Also: service.ComputeStockDensity lacks nil check after GetByID.
func TestBug006_ComputeStockDensityNil(t *testing.T) {
	// Test 1: model.ComputeStockDensity(nil) must not panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("model.ComputeStockDensity(nil) panicked: %v", r)
		}
	}()
	result := model.ComputeStockDensity(nil)
	if result != 0 {
		t.Errorf("expected 0 for nil tank, got %f", result)
	}

	// Test 2: service.ComputeStockDensity with non-existent tank ID
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
	_, err = svc.ComputeStockDensity(context.Background(), 99999)
	if err == nil {
		t.Fatal("expected error for non-existent tank, got nil")
	}
}