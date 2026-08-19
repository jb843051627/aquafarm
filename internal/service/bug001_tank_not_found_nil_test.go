package service

import (
	"context"
	"path/filepath"
	"testing"

	"aquafarm/internal/store"
)

// TestBug001_GetTankNotFoundNil verifies that GetTank returns ErrNotFound for a non-existent tank ID.
// Bug: GetByID returns nil,nil instead of nil,ErrNotFound, and GetTank lacks nil check.
func TestBug001_GetTankNotFoundNil(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	t.Cleanup(func() {})

	repo, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	defer repo.Close()

	if err := repo.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	svc := New(repo)

	// Query a non-existent tank — should return ErrNotFound
	_, err = svc.GetTank(context.Background(), 99999)
	if err == nil {
		t.Fatal("expected error for non-existent tank, got nil")
	}
}