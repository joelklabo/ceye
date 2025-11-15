package ui

import (
	"testing"
	"time"

	"github.com/joelklabo/ceye/internal/core"
)

func TestModelRunUpdatedMsgRefreshesTable(t *testing.T) {
	store := core.NewStore()
	store.Merge(core.RunEvent{
		Provider: "github",
		Runs: []core.Run{{
			ID:           "1",
			Provider:     "github",
			WorkflowName: "Build",
			Status:       core.RunStatusInProgress,
			Branch:       "main",
			UpdatedAt:    time.Now(),
		}},
	})
	m := NewModel(store, []string{"github"})

	updatedModel, _ := m.Update(RunUpdatedMsg{})
	actual := updatedModel.(Model)

	rows := actual.Table.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0][0] != "github" || rows[0][1] != "Build" {
		t.Fatalf("unexpected row values: %v", rows[0])
	}
}

func TestModelRunUpdatedMsgRespectsProviderFilter(t *testing.T) {
	store := core.NewStore()
	store.Merge(core.RunEvent{Provider: "github", Runs: []core.Run{{ID: "1", Provider: "github", WorkflowName: "Build", Branch: "main", Status: core.RunStatusInProgress, UpdatedAt: time.Now()}}})
	store.Merge(core.RunEvent{Provider: "azure", Runs: []core.Run{{ID: "2", Provider: "azure", WorkflowName: "Deploy", Branch: "main", Status: core.RunStatusCompleted, UpdatedAt: time.Now()}}})

	m := NewModel(store, []string{"github", "azure"})
	m.ActiveProvider = "github"

	updatedModel, _ := m.Update(RunUpdatedMsg{})
	actual := updatedModel.(Model)

	rows := actual.Table.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row for github filter, got %d", len(rows))
	}
	if rows[0][0] != "github" {
		t.Fatalf("expected github row, got %s", rows[0][0])
	}
}

func TestModelTableUpdatePassthrough(t *testing.T) {
	m := NewModel(core.NewStore(), nil)

	_, cmd := m.Update("noop")
	if cmd != nil {
		t.Fatalf("expected nil cmd for unhandled message, got %v", cmd)
	}
}
