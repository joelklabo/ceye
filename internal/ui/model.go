package ui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/joelklabo/ceye/internal/core"
)

// RunUpdatedMsg is emitted when the store receives new data.
type RunUpdatedMsg struct{}

// Model represents the Bubble Tea UI state.
type Model struct {
	Store          *core.Store
	Table          table.Model
	ActiveProvider string
	Providers      []string
}

// NewModel constructs a UI model.
func NewModel(store *core.Store, providers []string) Model {
	columns := []table.Column{
		{Title: "Provider", Width: 10},
		{Title: "Workflow", Width: 20},
		{Title: "Status", Width: 12},
		{Title: "Branch", Width: 15},
	}
	tbl := table.New(table.WithColumns(columns), table.WithRows([]table.Row{}))
	return Model{Store: store, Table: tbl, ActiveProvider: "all", Providers: providers}
}

// Init implements tea.Model.Init.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.Update.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case RunUpdatedMsg:
		m.refreshTable()
	}

	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

// View renders the UI.
func (m Model) View() string {
	return m.Table.View()
}

func (m *Model) refreshTable() {
	filter := ""
	if m.ActiveProvider != "all" {
		filter = m.ActiveProvider
	}
	runs := m.Store.ListRuns(filter)
	rows := make([]table.Row, 0, len(runs))
	for _, run := range runs {
		rows = append(rows, table.Row{run.Provider, run.WorkflowName, string(run.Status), run.Branch})
	}
	m.Table.SetRows(rows)
}
