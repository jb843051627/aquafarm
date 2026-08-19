package model

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
	"time"
)

// TestBug010_ParseScheduleTimezone verifies that ParseSchedule parses time slots
// in the local timezone, not UTC.
//
// Bug: ParseSchedule uses time.ParseInLocation("15:04", p, time.UTC),
// which parses times as UTC rather than local. While t.Hour() returns the
// same value regardless of timezone (since "15:04" only has hour:minute),
// the parsed time's Location is UTC, which can cause issues if the time
// value is used for timezone-aware comparisons.
//
// This test checks:
// 1. Functional: parsed slot's Hour() matches the input hour
// 2. Static: ParseSchedule does not use time.UTC as the parse location
func TestBug010_ParseScheduleTimezone(t *testing.T) {
	// === Functional test ===
	slots, err := ParseSchedule("08:00")
	if err != nil {
		t.Fatalf("ParseSchedule failed: %v", err)
	}
	if len(slots) != 1 {
		t.Fatalf("expected 1 slot, got %d", len(slots))
	}
	if slots[0].Hour() != 8 {
		t.Fatalf("expected Hour()=8, got %d", slots[0].Hour())
	}

	// === Static analysis: check that time.UTC is not used in ParseSchedule ===
	srcFile := filepath.Join("validation.go")
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, srcFile, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("failed to parse validation.go: %v", err)
	}

	// Find ParseSchedule and check for time.UTC usage
	var foundUTCInParse bool
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "ParseSchedule" {
			return true
		}

		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			// Look for time.UTC selector expression
			selExpr, ok := inner.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "time" && selExpr.Sel.Name == "UTC" {
				foundUTCInParse = true
			}
			return true
		})
		return true
	})

	if foundUTCInParse {
		t.Fatal("ParseSchedule uses time.UTC — should use local timezone for time parsing")
	}

	// Verify the slot is in local time, not UTC
	if slots[0].Location() != time.Local {
		t.Fatalf("expected slot Location to be time.Local, got %v", slots[0].Location())
	}
}