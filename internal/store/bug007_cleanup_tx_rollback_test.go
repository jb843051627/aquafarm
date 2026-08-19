package store

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

// TestBug007_CleanupTransactionRollback verifies that the Cleanup method does not use
// defer tx.Rollback() inside a loop body, which would delay the rollback until
// function return instead of executing it immediately.
//
// Bug: defer tx.Rollback() inside the loop body delays rollback.
// Fix: use tx.Rollback() directly (immediate).
//
// This test uses static analysis (AST inspection) to detect the defer-in-loop
// pattern in queries.go.
func TestBug007_CleanupTransactionRollback(t *testing.T) {
	srcFile := filepath.Join("queries.go")

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, srcFile, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("failed to parse queries.go: %v", err)
	}

	// Walk the AST to find defer statements inside for/range loops
	var foundDeferInLoop bool
	ast.Inspect(node, func(n ast.Node) bool {
		// Check both ForStmt and RangeStmt
		var loopBody *ast.BlockStmt
		switch stmt := n.(type) {
		case *ast.ForStmt:
			loopBody = stmt.Body
		case *ast.RangeStmt:
			loopBody = stmt.Body
		}
		if loopBody == nil {
			return true
		}

		// Check if there's a defer inside the loop body
		ast.Inspect(loopBody, func(inner ast.Node) bool {
			defStmt, ok := inner.(*ast.DeferStmt)
			if !ok {
				return true
			}
			// Check if it's a defer of tx.Rollback()
			if call, ok := defStmt.Call.Fun.(*ast.SelectorExpr); ok {
				if call.Sel.Name == "Rollback" {
					foundDeferInLoop = true
				}
			}
			return true
		})
		return true
	})

	if foundDeferInLoop {
		t.Fatal("found defer tx.Rollback() inside for/range loop — should use tx.Rollback() directly for immediate rollback")
	}
}