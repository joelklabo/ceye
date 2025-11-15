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
	Times     map[string]time.Time
}

// Model represents the Bubble Tea UI state.
type Model struct {
	Store             *core.Store
	Table             table.Model
	ActiveProvider    string
	Providers         []string
	visibleProviders  map[string]bool
	statusFilters     []string
	statusIndex       int
	Refresh           func()
	openURL           func(string)
	helpVisible       bool
	searchActive      bool
	searchQuery       string
	paletteVisible    bool
	paletteCursor     int
	Statuses          map[string]string
	ProviderTimes     map[string]time.Time
	visibleRuns       []core.Run
	runTotals         map[string]int
	lastUpdate        time.Time
	headerStyle       lipgloss.Style
	footerStyle       lipgloss.Style
	errorStyle        lipgloss.Style
	successStyle      lipgloss.Style
	failStyle         lipgloss.Style
	runningStyle      lipgloss.Style
}

// NewModel constructs a UI model.
func NewModel(store *core.Store, providers []string, refresh func(), openURL func(string)) Model {
	columns := []table.Column{
		{Title: "Provider", Width: 10},
		{Title: "Workflow", Width: 20},
		{Title: "Status", Width: 12},
		{Title: "Branch", Width: 15},
	}
	tbl := table.New(table.WithColumns(columns), table.WithRows([]table.Row{}))
	tbl.Focus()
	providerList := buildProviderList(providers)
	statusMap := make(map[string]string)
	visibleProviders := make(map[string]bool)
	for _, name := range providerList {
		if name == "all" {
			continue
		}
		statusMap[name] = ""
		visibleProviders[name] = true
	}
	m := Model{
		Store:             store,
		Table:             tbl,
		ActiveProvider:    providerList[0],
		Providers:         providerList,
		visibleProviders:  visibleProviders,
		statusFilters:     []string{"all", "running", "queued", "failed", "success"},
		statusIndex:       0,
		Refresh:           refresh,
		openURL:           openURL,
		helpVisible:       false,
		searchActive:      false,
		searchQuery:       "",
		paletteVisible:    false,
		paletteCursor:     0,
		Statuses:          statusMap,
		ProviderTimes:     make(map[string]time.Time),
		runTotals:         make(map[string]int),
		headerStyle:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
		footerStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		errorStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		successStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true),
		failStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		runningStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
	}
	return m
}

// Init implements tea.Model.Init.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.Update.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.paletteVisible {
		if key, ok := msg.(tea.KeyMsg); ok {
			if m.handlePaletteInput(key) {
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case RunUpdatedMsg:
		m.lastUpdate = msg.Timestamp
		if msg.Status != nil {
			m.Statuses = msg.Status
		}
		if msg.Times != nil {
			m.ProviderTimes = msg.Times
		}
		m.refreshTable()
		return m, nil
	case tea.KeyMsg:
		if m.handleSearchInput(msg) {
			return m, nil
		}
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
			case "f":
				m.cycleStatus()
				m.refreshTable()
				return m, nil
			case "p":
				m.togglePalette()
				return m, nil
			case "/":
				m.startSearch()
				return m, nil
			case "?":
				m.helpVisible = !m.helpVisible
				return m, nil
			case "o":
				m.openSelectedURL()
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
	detail := m.renderDetails()
	statusLine := m.renderStatuses()
	footer := lipgloss.JoinVertical(lipgloss.Left,
		m.footerStyle.Render(detail),
		m.footerStyle.Render(statusLine),
		m.renderHelp(),
	)
	view := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	if m.paletteVisible {
		view = lipgloss.JoinVertical(lipgloss.Left, view, m.renderPalette())
	}
	return view
}

func (m Model) renderStatuses() string {
	parts := []string{fmt.Sprintf("Totals: running %d | queued %d | failed %d | success %d",
		m.runTotals["running"], m.runTotals["queued"], m.runTotals["failed"], m.runTotals["success"],
	)}
	for _, name := range m.Providers {
		if name == "all" {
			continue
		}
		label := titleCase(name)
		status := "OK"
		if msg, ok := m.Statuses[name]; ok && msg != "" {
			status = m.errorStyle.Render(msg)
		}
		last := "never"
		if ts, ok := m.ProviderTimes[name]; ok && !ts.IsZero() {
			last = ts.Format("15:04:05")
		}
		parts = append(parts, fmt.Sprintf("%s %s (updated %s)", label, status, last))
	}
	if len(parts) == 0 {
		return "Status: no providers configured"
	}
	return "Status: " + strings.Join(parts, "  |  ")
}

func (m *Model) refreshTable() {
	providerFilter := ""
	if m.ActiveProvider != "all" {
		providerFilter = m.ActiveProvider
	}
	runs := m.Store.ListRuns(providerFilter)
	filtered := make([]core.Run, 0, len(runs))
	rows := make([]table.Row, 0, len(runs))
	statusFilter := m.statusFilters[m.statusIndex]
	m.runTotals = countStatuses(runs)
	for _, run := range runs {
		if providerFilter == "" {
			providerName := strings.ToLower(run.Provider)
			if allowed, ok := m.visibleProviders[providerName]; ok {
				if !allowed {
					continue
				}
			} else {
				// default visible providers added at init, but ensure new providers remain visible
				m.visibleProviders[providerName] = true
			}
		}
		if providerFilter == "" && !m.visibleProviders[strings.ToLower(run.Provider)] {
			continue
		}
		if !matchesStatusFilter(run, statusFilter) {
			continue
		}
		if !matchesSearch(run, m.searchQuery) {
			continue
		}
		rows = append(rows, table.Row{run.Provider, run.WorkflowName, m.formatStatus(run), run.Branch})
		filtered = append(filtered, run)
	}
	m.visibleRuns = filtered
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

func (m *Model) cycleStatus() {
	if len(m.statusFilters) == 0 {
		return
	}
	m.statusIndex = (m.statusIndex + 1) % len(m.statusFilters)
}

func (m *Model) startSearch() {
	m.searchActive = true
	m.searchQuery = ""
}

func (m Model) renderDetails() string {
	filterLabel := fmt.Sprintf("Status filter: %s", titleCase(m.statusFilters[m.statusIndex]))
	searchLabel := ""
	if m.searchQuery != "" {
		searchLabel = fmt.Sprintf(" | Search: %s", m.searchQuery)
	}
	if len(m.visibleRuns) == 0 {
		return filterLabel + searchLabel + " | Details: no runs"
	}
	idx := m.Table.Cursor()
	if idx < 0 || idx >= len(m.visibleRuns) {
		return filterLabel + searchLabel + " | Details: select a run"
	}
	run := m.visibleRuns[idx]
	return fmt.Sprintf(
		"%s%s | Details: %s | %s | SHA %s | %s",
		filterLabel,
		searchLabel,
		run.Repo,
		run.Branch,
		shortSHA(run.CommitSHA),
		run.URL,
	)
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func (m Model) openSelectedURL() {
	if m.openURL == nil || len(m.visibleRuns) == 0 {
		return
	}
	idx := m.Table.Cursor()
	if idx < 0 || idx >= len(m.visibleRuns) {
		return
	}
	run := m.visibleRuns[idx]
	if run.URL == "" {
		return
	}
	m.openURL(run.URL)
}

func (m Model) renderHelp() string {
	if m.helpVisible {
		return lipgloss.JoinVertical(lipgloss.Left,
			m.footerStyle.Render("Help: Tab provider, f status, / search, p providers, r refresh, o open, q quit"),
			m.footerStyle.Render("Use ↑/↓ or j/k to move, pgup/pgdn to page, Enter to select, ? hide help"),
		)
	}
	return m.footerStyle.Render("Press ? for help | Tab provider, f status, / search, p providers, r refresh, o open")
}

func (m *Model) handlePaletteInput(msg tea.KeyMsg) bool {
	if !m.paletteVisible {
		return false
	}
	switch msg.String() {
	case "up", "k":
		m.movePaletteCursor(-1)
	case "down", "j":
		m.movePaletteCursor(1)
	case " ":
		m.togglePaletteSelection()
	case "enter":
		m.paletteVisible = false
		m.refreshTable()
	case "esc":
		m.paletteVisible = false
	default:
		return true
	}
	return true
}

func (m *Model) movePaletteCursor(delta int) {
	items := m.paletteItems()
	if len(items) == 0 {
		return
	}
	m.paletteCursor += delta
	if m.paletteCursor < 0 {
		m.paletteCursor = len(items) - 1
	}
	if m.paletteCursor >= len(items) {
		m.paletteCursor = 0
	}
}

func (m *Model) togglePaletteSelection() {
	items := m.paletteItems()
	if len(items) == 0 {
		return
	}
	idx := m.paletteCursor
	if idx < 0 || idx >= len(items) {
		return
	}
	name := items[idx]
	m.visibleProviders[name] = !m.visibleProviders[name]
	m.refreshTable()
}

func (m *Model) togglePalette() {
	m.paletteVisible = !m.paletteVisible
	if m.paletteVisible {
		m.paletteCursor = 0
	}
}

func (m Model) paletteItems() []string {
	items := make([]string, 0, len(m.Providers))
	for _, name := range m.Providers {
		if name == "all" {
			continue
		}
		items = append(items, name)
	}
	return items
}

func (m Model) renderPalette() string {
	items := m.paletteItems()
	if len(items) == 0 {
		return ""
	}
	lines := []string{m.footerStyle.Render("Providers: space toggles, enter closes, esc cancels")}
	for i, name := range items {
		line := fmt.Sprintf("[%s] %s", checkbox(m.visibleProviders[name]), titleCase(name))
		if i == m.paletteCursor {
			lines = append(lines, m.headerStyle.Render(line))
		} else {
			lines = append(lines, m.footerStyle.Render(line))
		}
	}
	return strings.Join(lines, "\n")
}

func checkbox(on bool) string {
	if on {
		return "x"
	}
	return " "
}

func (m *Model) handleSearchInput(msg tea.KeyMsg) bool {
	if !m.searchActive {
		return false
	}
	switch msg.Type {
	case tea.KeyEscape:
		m.searchActive = false
		m.searchQuery = ""
		m.refreshTable()
	case tea.KeyEnter:
		m.searchActive = false
		m.refreshTable()
	case tea.KeyBackspace, tea.KeyCtrlH:
		if len(m.searchQuery) > 0 {
			runes := []rune(m.searchQuery)
			m.searchQuery = string(runes[:len(runes)-1])
			m.refreshTable()
		}
	default:
		if len(msg.Runes) > 0 {
			m.searchQuery += string(msg.Runes)
			m.refreshTable()
		}
	}
	return true
}

func matchesSearch(run core.Run, query string) bool {
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return true
	}
	fields := []string{
		strings.ToLower(run.Provider),
		strings.ToLower(run.Repo),
		strings.ToLower(run.WorkflowName),
		strings.ToLower(run.Branch),
		strings.ToLower(string(run.Status)),
		strings.ToLower(run.Conclusion),
		strings.ToLower(run.URL),
	}
	for _, f := range fields {
		if strings.Contains(f, q) {
			return true
		}
	}
	return false
}

func countStatuses(runs []core.Run) map[string]int {
	counts := map[string]int{
		"running": 0,
		"queued":  0,
		"failed":  0,
		"success": 0,
	}
	for _, run := range runs {
		switch run.Status {
		case core.RunStatusInProgress:
			counts["running"]++
		case core.RunStatusQueued:
			counts["queued"]++
		case core.RunStatusFailed, core.RunStatusCancelled:
			counts["failed"]++
		case core.RunStatusCompleted:
			if strings.EqualFold(run.Conclusion, "success") || strings.EqualFold(run.Conclusion, "succeeded") || run.Conclusion == "" {
				counts["success"]++
			} else {
				counts["failed"]++
			}
		}
	}
	return counts
}

func matchesStatusFilter(run core.Run, filter string) bool {
	switch filter {
	case "all":
		return true
	case "running":
		return run.Status == core.RunStatusInProgress
	case "queued":
		return run.Status == core.RunStatusQueued
	case "failed":
		if run.Status == core.RunStatusFailed || run.Status == core.RunStatusCancelled {
			return true
		}
		if run.Status == core.RunStatusCompleted {
			c := strings.ToLower(run.Conclusion)
			return c == "failure" || c == "failed" || c == "cancelled" || c == "canceled"
		}
		return false
	case "success":
		return run.Status == core.RunStatusCompleted && (run.Conclusion == "" || strings.EqualFold(run.Conclusion, "success") || strings.EqualFold(run.Conclusion, "succeeded"))
	default:
		return true
	}
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
