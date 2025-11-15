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
	Status    map[string]string
}

// Model represents the Bubble Tea UI state.
type Model struct {
	Store          *core.Store
	Table          table.Model
	ActiveProvider string
	Providers      []string
	Refresh        func()
	Statuses       map[string]string
	lastUpdate     time.Time
	headerStyle    lipgloss.Style
	footerStyle    lipgloss.Style
	errorStyle     lipgloss.Style
	successStyle   lipgloss.Style
	failStyle      lipgloss.Style
	runningStyle   lipgloss.Style
}

// NewModel constructs a UI model.
func NewModel(store *core.Store, providers []string, refresh func()) Model {
	columns := []table.Column{
		{Title: "Provider", Width: 10},
		{Title: "Workflow", Width: 20},
		{Title: "Status", Width: 12},
		{Title: "Branch", Width: 15},
	}
	tbl := table.New(table.WithColumns(columns), table.WithRows([]table.Row{}))
	providerList := buildProviderList(providers)
	statusMap := make(map[string]string)
	for _, name := range providerList {
		if name == "all" {
			continue
		}
		statusMap[name] = ""
	}
	m := Model{
		Store:          store,
		Table:          tbl,
		ActiveProvider: providerList[0],
		Providers:      providerList,
		Refresh:        refresh,
		Statuses:       statusMap,
		headerStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
		footerStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		errorStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		successStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true),
		failStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		runningStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
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
		if msg.Status != nil {
			m.Statuses = msg.Status
		}
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
			switch r {
			case "q":
				return m, tea.Quit
			case "r":
				if m.Refresh != nil {
					m.Refresh()
				}
				return m, nil
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
	statusLine := m.renderStatuses()
	footer := lipgloss.JoinVertical(lipgloss.Left,
		m.footerStyle.Render("Tab: cycle providers  |  r: refresh  |  q: quit"),
		m.footerStyle.Render(statusLine),
	)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m Model) renderStatuses() string {
	parts := make([]string, 0, len(m.Statuses))
	for _, name := range m.Providers {
		if name == "all" {
			continue
		}
		label := titleCase(name)
		status := "OK"
		if msg, ok := m.Statuses[name]; ok && msg != "" {
			status = m.errorStyle.Render(msg)
		}
		parts = append(parts, fmt.Sprintf("%s %s", label, status))
	}
	if len(parts) == 0 {
		return "Status: no providers configured"
	}
	return "Status: " + strings.Join(parts, "  |  ")
}

func (m *Model) refreshTable() {
	filter := ""
	if m.ActiveProvider != "all" {
		filter = m.ActiveProvider
	}
	runs := m.Store.ListRuns(filter)
	rows := make([]table.Row, 0, len(runs))
	for _, run := range runs {
		rows = append(rows, table.Row{run.Provider, run.WorkflowName, m.formatStatus(run), run.Branch})
	}
	m.Table.SetRows(rows)
}

func (m *Model) formatStatus(run core.Run) string {
	statusText := strings.ToLower(string(run.Status))
	switch run.Status {
	case core.RunStatusCompleted:
		conclusion := strings.ToLower(run.Conclusion)
		switch conclusion {
		case "success", "succeeded":
			return m.successStyle.Render("success")
		case "failure", "failed", "cancelled", "canceled":
			return m.failStyle.Render(conclusion)
		case "":
			return m.successStyle.Render("done")
		default:
			return m.runningStyle.Render(conclusion)
		}
	case core.RunStatusInProgress:
		return m.runningStyle.Render("running")
	case core.RunStatusQueued:
		return m.runningStyle.Render("queued")
	default:
		return statusText
	}
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
