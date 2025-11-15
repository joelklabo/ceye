package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
	m := NewModel(store, []string{"github"}, nil, nil)

	updatedModel, _ := m.Update(RunUpdatedMsg{Timestamp: time.Now()})
	actual := updatedModel.(Model)

	rows := actual.Table.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0][0] != "github" || rows[0][2] != "Build" {
		t.Fatalf("unexpected row values: %v", rows[0])
	}
}

func TestModelRunUpdatedMsgRespectsProviderFilter(t *testing.T) {
	store := core.NewStore()
	store.Merge(core.RunEvent{Provider: "github", Runs: []core.Run{{ID: "1", Provider: "github", WorkflowName: "Build", Branch: "main", Status: core.RunStatusInProgress, UpdatedAt: time.Now()}}})
	store.Merge(core.RunEvent{Provider: "azure", Runs: []core.Run{{ID: "2", Provider: "azure", WorkflowName: "Deploy", Branch: "main", Status: core.RunStatusCompleted, UpdatedAt: time.Now()}}})

	m := NewModel(store, []string{"github", "azure"}, nil, nil)
	m.ActiveProvider = "github"

	updatedModel, _ := m.Update(RunUpdatedMsg{Timestamp: time.Now()})
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
	m := NewModel(core.NewStore(), nil, nil, nil)

	_, cmd := m.Update("noop")
	if cmd != nil {
		t.Fatalf("expected nil cmd for unhandled message, got %v", cmd)
	}
}

func TestModelQuitKeys(t *testing.T) {
	m := NewModel(core.NewStore(), nil, nil, nil)

	tests := []tea.KeyMsg{
		{Type: tea.KeyCtrlC},
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
	}

	for _, key := range tests {
		_, cmd := m.Update(key)
		if cmd == nil {
			t.Fatalf("expected quit command for key %v", key)
		}
		msg := cmd()
		if msg == nil {
			t.Fatalf("expected quit msg from command")
		}
		if _, ok := msg.(tea.QuitMsg); !ok {
			t.Fatalf("expected tea.QuitMsg, got %T", msg)
		}
	}
}

func TestModelProviderCycleOnTab(t *testing.T) {
	store := core.NewStore()
	m := NewModel(store, []string{"github", "azure"}, nil, nil)
	m.ActiveProvider = "github"

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	actual := next.(Model)
	if actual.ActiveProvider != "azure" {
		t.Fatalf("expected active provider to cycle to azure, got %s", actual.ActiveProvider)
	}

	// Cycle again should wrap back to all
	next, _ = actual.Update(tea.KeyMsg{Type: tea.KeyTab})
	actual = next.(Model)
	if actual.ActiveProvider != "all" {
		t.Fatalf("expected active provider to wrap to all, got %s", actual.ActiveProvider)
	}
}

func TestModelRefreshKeyInvokesCallback(t *testing.T) {
	called := 0
	refresh := func() { called++ }
	m := NewModel(core.NewStore(), nil, refresh, nil)

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if called != 1 {
		t.Fatalf("expected refresh callback once, got %d", called)
	}
}
