package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/joelklabo/ceye/internal/core"
)

// RunUpdatedMsg is emitted when the store receives new data.
type RunUpdatedMsg struct {
	Timestamp time.Time
}

// Model represents the Bubble Tea UI state.
type Model struct {
	Store          *core.Store
	Table          table.Model
	ActiveProvider string
	Providers      []string
	lastUpdate     time.Time
	headerStyle    lipgloss.Style
	footerStyle    lipgloss.Style
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
	providerList := buildProviderList(providers)
	m := Model{
		Store:          store,
		Table:          tbl,
		ActiveProvider: providerList[0],
		Providers:      providerList,
		headerStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
		footerStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
	}
	return m
}

// Init implements tea.Model.Init.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.Update.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case RunUpdatedMsg:
		m.lastUpdate = msg.Timestamp
		m.refreshTable()
		return m, nil
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		case msg.Type == tea.KeyTab:
			m.cycleProvider()
			m.refreshTable()
			return m, nil
		case msg.Type == tea.KeyRunes:
			if len(msg.Runes) == 0 {
				break
			}
			r := strings.ToLower(string(msg.Runes))
			if r == "q" {
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

// View renders the UI.
func (m Model) View() string {
	last := "never"
	if !m.lastUpdate.IsZero() {
		last = m.lastUpdate.Format("15:04:05")
	}
	header := m.headerStyle.Render(fmt.Sprintf("Viewing: %s | Last update: %s", titleCase(m.ActiveProvider), last))
	body := m.Table.View()
	footer := m.footerStyle.Render("Tab: cycle providers  |  q: quit")
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
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

func (m *Model) cycleProvider() {
	if len(m.Providers) == 0 {
		return
	}
	current := m.currentProviderIndex()
	next := (current + 1) % len(m.Providers)
	m.ActiveProvider = m.Providers[next]
}

func buildProviderList(names []string) []string {
	seen := map[string]bool{"all": true}
	list := []string{"all"}
	for _, n := range names {
		if n == "" {
			continue
		}
		n = strings.ToLower(n)
		if seen[n] {
			continue
		}
		list = append(list, n)
		seen[n] = true
	}
	return list
}

func (m *Model) currentProviderIndex() int {
	for i, name := range m.Providers {
		if name == m.ActiveProvider {
			return i
		}
	}
	return 0
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
