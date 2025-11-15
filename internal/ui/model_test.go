package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/joelklabo/ceye/internal/core"
	"github.com/joelklabo/ceye/internal/providers"
	"github.com/joelklabo/ceye/internal/providers/manager"
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
	m := NewModel(store, []string{"github"}, nil, nil, nil)

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

	m := NewModel(store, []string{"github", "azure"}, nil, nil, nil)
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
	m := NewModel(core.NewStore(), nil, nil, nil, nil)

	_, cmd := m.Update("noop")
	if cmd != nil {
		t.Fatalf("expected nil cmd for unhandled message, got %v", cmd)
	}
}

func TestModelQuitKeys(t *testing.T) {
	m := NewModel(core.NewStore(), nil, nil, nil, nil)

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
	m := NewModel(store, []string{"github", "azure"}, nil, nil, nil)
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

func TestModelProviderStoreToggle(t *testing.T) {
	m := NewModel(core.NewStore(), nil, nil, nil, nil)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	actual := next.(Model)
	if !actual.providerStoreVisible {
		t.Fatalf("expected provider store overlay visible after 'P'")
	}

	next, _ = actual.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	actual = next.(Model)
	if actual.providerStoreVisible {
		t.Fatalf("expected provider store overlay hidden after 'P' again")
	}
}

func TestModelProviderStoreEntriesUpdate(t *testing.T) {
	m := NewModel(core.NewStore(), nil, nil, nil, nil)
	storeRecords := []manager.ProviderRecord{
		{
			ID:      "abc123",
			Enabled: true,
			Config: providers.ProviderConfig{
				Type:        "demo",
				DisplayName: "demo store entry",
			},
		},
	}
	next, _ := m.Update(RunUpdatedMsg{Timestamp: time.Now(), Store: storeRecords})
	actual := next.(Model)
	if len(actual.providerStoreEntries) != 1 {
		t.Fatalf("expected provider store entries to update, got %d", len(actual.providerStoreEntries))
	}
	if actual.providerStoreEntries[0].Config.Type != "demo" {
		t.Fatalf("expected demo provider saved, got %s", actual.providerStoreEntries[0].Config.Type)
	}
}

func TestModelProviderStoreActionSpace(t *testing.T) {
	m := NewModel(core.NewStore(), nil, nil, nil, nil)
	m.providerStoreVisible = true
	m.providerStoreEntries = []manager.ProviderRecord{
		{ID: "abc", Enabled: true, Config: providers.ProviderConfig{Type: "demo"}},
	}
	called := false
	var last manager.ProviderRecord
	var enabled bool
	m.SetProviderStoreAction(func(entry manager.ProviderRecord, action ProviderStoreActionType) {
		called = true
		last = entry
		if action != ProviderStoreActionToggle {
			t.Fatalf("expected toggle action, got %v", action)
		}
		enabled = entry.Enabled
	})

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	if !called {
		t.Fatalf("expected action called")
	}
	if last.ID != "abc" {
		t.Fatalf("expected entry abc, got %s", last.ID)
	}
	if enabled {
		t.Fatalf("expected disable toggle, got enabled=%t", enabled)
	}
	if m.providerStoreEntries[0].Enabled {
		t.Fatalf("expected local entry updated to disabled")
	}
}

func TestModelProviderStoreRemoveKey(t *testing.T) {
	m := NewModel(core.NewStore(), nil, nil, nil, nil)
	m.providerStoreVisible = true
	m.providerStoreEntries = []manager.ProviderRecord{
		{ID: "abc", Enabled: true, Config: providers.ProviderConfig{Type: "demo"}},
	}
	called := false
	m.SetProviderStoreAction(func(entry manager.ProviderRecord, action ProviderStoreActionType) {
		called = true
		if action != ProviderStoreActionRemove {
			t.Fatalf("expected remove action, got %v", action)
		}
		if entry.ID != "abc" {
			t.Fatalf("expected entry abc, got %s", entry.ID)
		}
	})

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	if !called {
		t.Fatalf("expected remove action called")
	}
}

func TestModelProviderStoreDuplicateKey(t *testing.T) {
	m := NewModel(core.NewStore(), nil, nil, nil, nil)
	m.providerStoreVisible = true
	m.providerStoreEntries = []manager.ProviderRecord{
		{ID: "abc", Enabled: true, Config: providers.ProviderConfig{Type: "demo"}},
	}
	called := false
	m.SetProviderStoreAction(func(entry manager.ProviderRecord, action ProviderStoreActionType) {
		called = true
		if action != ProviderStoreActionDuplicate {
			t.Fatalf("expected duplicate action, got %v", action)
		}
		if entry.ID != "abc" {
			t.Fatalf("expected entry abc, got %s", entry.ID)
		}
	})

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	if !called {
		t.Fatalf("expected duplicate action called")
	}
}

func TestModelSetProviderListUpdatesVisibility(t *testing.T) {
	m := NewModel(core.NewStore(), []string{"github"}, nil, nil, nil)
	m.SetProviderList([]string{"github", "azure"})
	if m.ActiveProvider != "all" {
		t.Fatalf("expected active provider reset to all, got %s", m.ActiveProvider)
	}
	if len(m.Providers) != 3 {
		t.Fatalf("expected providers list to contain all, github, azure; got %v", m.Providers)
	}
	if !m.visibleProviders["github"] || !m.visibleProviders["azure"] {
		t.Fatalf("expected visibility map to include github and azure, got %v", m.visibleProviders)
	}
}

func TestModelRefreshKeyInvokesCallback(t *testing.T) {
	called := 0
	refresh := func() { called++ }
	m := NewModel(core.NewStore(), nil, refresh, nil, nil)

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if called != 1 {
		t.Fatalf("expected refresh callback once, got %d", called)
	}
}

func TestModelHelpOverlayToggle(t *testing.T) {
	m := NewModel(core.NewStore(), nil, nil, nil, nil)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	actual := next.(Model)
	if !actual.helpVisible {
		t.Fatalf("expected help overlay visible after '?'")
	}

	next, _ = actual.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	actual = next.(Model)
	if actual.helpVisible {
		t.Fatalf("expected help overlay hidden after '?' again")
	}
}

func TestModelFocusModeToggle(t *testing.T) {
	m := NewModel(core.NewStore(), nil, nil, nil, nil)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	actual := next.(Model)
	if !actual.focusMode {
		t.Fatalf("expected focus mode enabled after 'v'")
	}

	next, _ = actual.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	actual = next.(Model)
	if actual.focusMode {
		t.Fatalf("expected focus mode disabled after second 'v'")
	}
}

func TestModelCycleSort(t *testing.T) {
	m := NewModel(core.NewStore(), nil, nil, nil, nil)
	initial := m.sortIndex
	m.cycleSort()
	if m.sortIndex == initial {
		t.Fatalf("expected sort index to advance")
	}
}

func TestModelSortByUpdated(t *testing.T) {
	store := core.NewStore()
	old := time.Now().Add(-1 * time.Hour)
	newer := time.Now()
	store.Merge(core.RunEvent{Provider: "github", Runs: []core.Run{{ID: "1", Provider: "github", WorkflowName: "Build", Branch: "main", Status: core.RunStatusInProgress, UpdatedAt: old}}})
	store.Merge(core.RunEvent{Provider: "github", Runs: []core.Run{{ID: "2", Provider: "github", WorkflowName: "Test", Branch: "dev", Status: core.RunStatusCompleted, Conclusion: "success", UpdatedAt: newer}}})
	m := NewModel(store, []string{"github"}, nil, nil, nil)
	m.sortIndex = 1 // updated
	m.refreshTable()
	rows := m.Table.Rows()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0][2] != "Test" {
		t.Fatalf("expected newest run first, got %v", rows[0])
	}
}

func TestModelCopyURLKey(t *testing.T) {
	store := core.NewStore()
	url := "https://example.com/run/1"
	store.Merge(core.RunEvent{Provider: "github", Runs: []core.Run{{ID: "1", Provider: "github", WorkflowName: "Build", Branch: "main", Status: core.RunStatusInProgress, UpdatedAt: time.Now(), URL: url}}})
	copied := ""
	copyFn := func(text string) { copied = text }
	m := NewModel(store, []string{"github"}, nil, nil, copyFn)
	next, _ := m.Update(RunUpdatedMsg{Timestamp: time.Now()})
	m = next.(Model)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if copied != url {
		t.Fatalf("expected URL copied, got %q", copied)
	}
}

func TestModelEnterOpensURL(t *testing.T) {
	store := core.NewStore()
	url := "https://example.com/run/enter"
	store.Merge(core.RunEvent{Provider: "github", Runs: []core.Run{{ID: "1", Provider: "github", WorkflowName: "Deploy", Branch: "main", Status: core.RunStatusCompleted, Conclusion: "success", UpdatedAt: time.Now(), URL: url}}})
	opened := ""
	openFn := func(text string) { opened = text }
	m := NewModel(store, []string{"github"}, nil, openFn, nil)
	next, _ := m.Update(RunUpdatedMsg{Timestamp: time.Now()})
	m = next.(Model)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if opened != url {
		t.Fatalf("expected enter to open URL, got %q", opened)
	}
}

func TestModelCopySummaryKey(t *testing.T) {
	store := core.NewStore()
	url := "https://example.com/run/summary"
	store.Merge(core.RunEvent{Provider: "github", Runs: []core.Run{{
		ID:           "1",
		Provider:     "github",
		WorkflowName: "Build",
		Branch:       "feat",
		Repo:         "org/repo",
		Status:       core.RunStatusInProgress,
		UpdatedAt:    time.Now(),
		URL:          url,
	}}})
	copied := ""
	copyFn := func(text string) { copied = text }
	m := NewModel(store, []string{"github"}, nil, nil, copyFn)
	next, _ := m.Update(RunUpdatedMsg{Timestamp: time.Now()})
	m = next.(Model)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if !strings.Contains(copied, "org/repo") || !strings.Contains(copied, url) {
		t.Fatalf("expected summary to include repo and url, got %q", copied)
	}
}
