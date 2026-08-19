package service

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
	"time"

	"aquafarm/internal/model"
	"aquafarm/internal/store"
)

// TestBug009_CreateFeedPlanErrorShadowing verifies two aspects:
// 1. Functional: CreateFeedPlan returns an error for a non-existent tank.
// 2. Static: CreateFeedPlan does not shadow the err variable with := inside
//    the CreatePlan if-statement (after GetByID already set err via =).
//
// Bug: err is shadowed by := in `if err := s.repo.Feeds().CreatePlan(f)`,
// breaking the error chain. The fix uses `err = s.repo.Feeds().CreatePlan(f)`.
func TestBug009_CreateFeedPlanErrorShadowing(t *testing.T) {
	// === Functional test ===
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

	plan := &model.FeedPlan{
		TankID:    99999,
		FeedType:  "pellets",
		Amount:    50,
		Schedule:  "08:00",
		Active:    true,
		CreatedAt: time.Now(),
	}

	_, err = svc.CreateFeedPlan(context.Background(), plan)
	if err == nil {
		t.Fatal("expected error for non-existent tank, got nil (err shadowing bug)")
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}

	// === Static analysis: check for err shadowing in CreatePlan call ===
	srcFile := filepath.Join("tank_feed_service.go")
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, srcFile, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("failed to parse tank_feed_service.go: %v", err)
	}

	// Find the CreateFeedPlan function and look for the specific pattern:
	// `if err := s.repo.Feeds().CreatePlan(f); err != nil`
	var foundShadowedErr bool
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "CreateFeedPlan" {
			return true
		}

		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			ifStmt, ok := inner.(*ast.IfStmt)
			if !ok {
				return true
			}
			// Check if the if has an init statement using :=
			init, ok := ifStmt.Init.(*ast.AssignStmt)
			if !ok || init.Tok != token.DEFINE {
				return true
			}
			// Check if any of the LHS variables is "err"
			hasErr := false
			for _, expr := range init.Lhs {
				if ident, ok := expr.(*ast.Ident); ok && ident.Name == "err" {
					hasErr = true
				}
			}
			if !hasErr {
				return true
			}
			// Check if the RHS calls CreatePlan
			if call, ok := init.Rhs[0].(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if sel.Sel.Name == "CreatePlan" {
						foundShadowedErr = true
					}
				}
			}
			return true
		})
		return true
	})

	if foundShadowedErr {
		t.Fatal("found err shadowed with := in `if err := ...CreatePlan(f)` — should use `err = ...` to reuse existing err")
	}
}