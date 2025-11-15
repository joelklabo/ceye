package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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
	Message   string
	Level     string
}

type keyMap struct {
	Provider key.Binding
	Status   key.Binding
	Search   key.Binding
	Palette  key.Binding
	Refresh  key.Binding
	Open     key.Binding
	Focus    key.Binding
	Sort     key.Binding
	Copy     key.Binding
	Help     key.Binding
	Quit     key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Provider: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "cycle providers")),
		Status:   key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "cycle status")),
		Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Palette:  key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "provider palette")),
		Refresh:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Open:     key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open run")),
		Focus:    key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "toggle focus view")),
		Sort:     key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "cycle sort")),
		Copy:     key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy URL")),
		Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
		Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Provider, k.Status, k.Sort, k.Copy, k.Focus, k.Refresh, k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Provider, k.Status, k.Sort, k.Search},
		{k.Palette, k.Focus, k.Refresh, k.Open},
		{k.Copy, k.Help, k.Quit},
	}
}

// Model represents the Bubble Tea UI state.
type Model struct {
	Store            *core.Store
	Table            table.Model
	ActiveProvider   string
	Providers        []string
	visibleProviders map[string]bool
	statusFilters    []string
	statusIndex      int
	Refresh          func()
	openURL          func(string)
	copyText         func(string)
	helpModel        help.Model
	keys             keyMap
	searchActive     bool
	searchQuery      string
	paletteVisible   bool
	paletteCursor    int
	helpVisible      bool
	focusMode        bool
	sortModes        []string
	sortIndex        int
	Statuses         map[string]string
	ProviderTimes    map[string]time.Time
	visibleRuns      []core.Run
	runTotals        map[string]int
	logEntries       []logEntry
	lastUpdate       time.Time
	width            int
	height           int
	headerStyle      lipgloss.Style
	footerStyle      lipgloss.Style
	bodyBoxStyle     lipgloss.Style
	panelStyle       lipgloss.Style
	tagStyle         lipgloss.Style
	tagWarningStyle  lipgloss.Style
	tagErrorStyle    lipgloss.Style
	errorStyle       lipgloss.Style
	successStyle     lipgloss.Style
	failStyle        lipgloss.Style
	runningStyle     lipgloss.Style
}

// NewModel constructs a UI model.
func NewModel(store *core.Store, providers []string, refresh func(), openURL func(string), copyText func(string)) Model {
	columns := []table.Column{
		{Title: "Provider", Width: 10},
		{Title: "Repository", Width: 24},
		{Title: "Workflow", Width: 24},
		{Title: "Status", Width: 12},
		{Title: "Branch", Width: 16},
		{Title: "Updated", Width: 10},
		{Title: "Duration", Width: 9},
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
	keys := newKeyMap()
	helpModel := help.New()
	helpModel.ShowAll = false
	m := Model{
		Store:            store,
		Table:            tbl,
		ActiveProvider:   providerList[0],
		Providers:        providerList,
		visibleProviders: visibleProviders,
		statusFilters:    []string{"all", "running", "queued", "failed", "success"},
		statusIndex:      0,
		sortModes:        []string{"status", "updated", "duration"},
		sortIndex:        0,
		Refresh:          refresh,
		openURL:          openURL,
		copyText:         copyText,
		helpModel:        helpModel,
		keys:             keys,
		searchActive:     false,
		searchQuery:      "",
		paletteVisible:   false,
		paletteCursor:    0,
		helpVisible:      false,
		Statuses:         statusMap,
		ProviderTimes:    make(map[string]time.Time),
		runTotals:        make(map[string]int),
		logEntries:       make([]logEntry, 0),
		focusMode:        false,
		headerStyle:      headerStyle,
		footerStyle:      footerStyle,
		bodyBoxStyle:     bodyBox,
		panelStyle:       panel,
		tagStyle:         tag,
		tagWarningStyle:  tagWarn,
		tagErrorStyle:    tagErr,
		errorStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		successStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true),
		failStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		runningStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
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
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if m.handlePaletteInput(keyMsg) {
				return m, nil
			}
			return m, nil
		}
	}
	if m.helpVisible {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if m.handleHelpOverlayInput(keyMsg) {
				return m, nil
			}
			return m, nil
		}
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
		if msg.Message != "" {
			entry := logEntry{text: msg.Message, timestamp: msg.Timestamp, level: msg.Level}
			m.logEntries = append([]logEntry{entry}, m.logEntries...)
			if len(m.logEntries) > 8 {
				m.logEntries = m.logEntries[:8]
			}
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
			case "o":
				m.openSelectedURL()
				return m, nil
			case "y":
				m.copySelectedURL()
				return m, nil
			case "t":
				m.cycleSort()
				m.refreshTable()
				return m, nil
			case "v":
				m.toggleFocusMode()
				return m, nil
			case "?":
				m.toggleHelpOverlay()
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableSize()
		return m, nil
	}

	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

// View renders the UI.
func (m Model) View() string {
	header := m.renderHeader()
	stats := m.renderStatsBar()
	sections := []string{header}
	if stats != "" {
		sections = append(sections, stats)
	}
	if m.focusMode {
		sections = append(sections, m.renderFocusBody())
	} else if m.compactLayout() {
		sections = append(sections, m.renderCompactBody())
	} else {
		sections = append(sections, m.renderStatuses(), m.renderDashboardBody())
	}
	sections = append(sections, m.renderHelp())
	view := lipgloss.JoinVertical(lipgloss.Left, sections...)
	if m.paletteVisible {
		view = lipgloss.JoinHorizontal(lipgloss.Top, view, m.renderPalette())
	}
	if m.helpVisible {
		view = lipgloss.JoinHorizontal(lipgloss.Top, view, m.renderHelpOverlay())
	}
	return view
}

func (m Model) renderHeader() string {
	last := "never"
	if !m.lastUpdate.IsZero() {
		last = m.lastUpdate.Format("15:04:05")
	}
	totals := fmt.Sprintf("Runs: %d (running %d | queued %d | failed %d | success %d)",
		len(m.visibleRuns),
		m.runTotals["running"],
		m.runTotals["queued"],
		m.runTotals["failed"],
		m.runTotals["success"],
	)
	sortLabel := titleCase(m.sortModes[m.sortIndex])
	filters := fmt.Sprintf("Provider: %s | Status: %s | Sort: %s | Search: %s",
		titleCase(m.ActiveProvider),
		titleCase(m.statusFilters[m.statusIndex]),
		sortLabel,
		func() string {
			if m.searchQuery == "" {
				return "none"
			}
			return m.searchQuery
		}(),
	)
	content := lipgloss.JoinVertical(lipgloss.Left,
		m.headerStyle.Render(fmt.Sprintf("CI Status Dashboard  •  Last update %s", last)),
		m.footerStyle.Render(totals),
		m.footerStyle.Render(filters),
	)
	return content
}

func (m Model) renderRunsTable() string {
	title := sectionTitleStyle.Render(fmt.Sprintf("Runs (%d showing)", len(m.visibleRuns)))
	return lipgloss.JoinVertical(lipgloss.Left, title, m.bodyBoxStyle.Render(m.Table.View()))
}

func (m Model) renderDashboardBody() string {
	table := m.renderRunsTable()
	sidebar := lipgloss.JoinVertical(lipgloss.Left, m.renderDetails(), m.renderLogs())
	return lipgloss.JoinHorizontal(lipgloss.Top, table, sidebar)
}

func (m Model) renderCompactBody() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderStatuses(),
		m.renderRunsTable(),
		m.renderDetails(),
		m.renderLogs(),
	)
}

func (m Model) renderFocusBody() string {
	banner := m.panelStyle.Render(bodyTextStyle.Render("Focus mode: table maximized (press 'v' to return to dashboard view)"))
	lower := lipgloss.JoinHorizontal(lipgloss.Top, m.renderDetails(), m.renderLogs())
	body := []string{banner, m.renderRunsTable()}
	if m.compactLayout() {
		body = append(body, m.renderDetails(), m.renderLogs(), m.renderStatuses())
	} else {
		body = append(body, lower, m.renderStatuses())
	}
	return lipgloss.JoinVertical(lipgloss.Left, body...)
}

func (m Model) renderStatsBar() string {
	if len(m.visibleRuns) == 0 && len(m.runTotals) == 0 {
		return ""
	}
	boxes := []string{
		statBoxStyle.Render(fmt.Sprintf("Total\n%d", len(m.visibleRuns))),
		statRunningStyle.Render(fmt.Sprintf("Running\n%d", m.runTotals["running"])),
		statQueuedStyle.Render(fmt.Sprintf("Queued\n%d", m.runTotals["queued"])),
		statFailedStyle.Render(fmt.Sprintf("Failed\n%d", m.runTotals["failed"])),
		statSuccessStyle.Render(fmt.Sprintf("Success\n%d", m.runTotals["success"])),
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, boxes...)
}

func (m Model) renderStatuses() string {
	items := m.renderProviderBadges()
	body := "Providers: none configured"
	if len(items) > 0 {
		body = lipgloss.JoinHorizontal(lipgloss.Top, items...)
	}
	content := lipgloss.JoinVertical(lipgloss.Left, sectionTitleStyle.Render("Providers"), bodyTextStyle.Render(body))
	return m.panelStyle.Render(content)
}

func (m Model) renderProviderBadges() []string {
	parts := make([]string, 0)
	for _, name := range m.Providers {
		if name == "all" {
			continue
		}
		label := titleCase(name)
		style := m.tagStyle
		if msg, ok := m.Statuses[name]; ok && msg != "" {
			label = fmt.Sprintf("%s ! %s", label, msg)
			style = m.tagErrorStyle
		} else if ts, ok := m.ProviderTimes[name]; ok && !ts.IsZero() {
			label = fmt.Sprintf("%s %s", label, ts.Format("15:04:05"))
		} else {
			label = fmt.Sprintf("%s waiting", label)
			style = m.tagWarningStyle
		}
		parts = append(parts, style.Render(label))
	}
	return parts
}

func (m Model) renderLogs() string {
	if len(m.logEntries) == 0 {
		return m.panelStyle.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			sectionTitleStyle.Render("Activity"),
			bodyTextStyle.Render("none yet"),
		))
	}
	lines := make([]string, 0, len(m.logEntries)+1)
	lines = append(lines, sectionTitleStyle.Render("Activity"))
	for _, entry := range m.logEntries {
		line := fmt.Sprintf("%s — %s", entry.timestamp.Format("15:04:05"), entry.text)
		lines = append(lines, renderLogLine(entry.level, line))
	}
	return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func renderLogLine(level, text string) string {
	switch level {
	case "error":
		return logErrorStyle.Render(text)
	case "warn":
		return logWarnStyle.Render(text)
	default:
		return logInfoStyle.Render(text)
	}
}

func (m *Model) refreshTable() {
	providerFilter := ""
	if m.ActiveProvider != "all" {
		providerFilter = m.ActiveProvider
	}
	runs := m.Store.ListRuns(providerFilter)
	filtered := make([]core.Run, 0, len(runs))
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
		filtered = append(filtered, run)
	}
	sorted := m.sortRuns(filtered)
	rows := make([]table.Row, 0, len(sorted))
	for _, run := range sorted {
		rows = append(rows, table.Row{
			strings.ToLower(run.Provider),
			run.Repo,
			run.WorkflowName,
			m.formatStatus(run),
			run.Branch,
			formatRelativeTime(run.UpdatedAt),
			formatDuration(run.Duration, run.StartedAt, run.UpdatedAt),
		})
	}
	m.visibleRuns = sorted
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

func (m *Model) sortRuns(runs []core.Run) []core.Run {
	if len(runs) <= 1 {
		return runs
	}
	sorted := make([]core.Run, len(runs))
	copy(sorted, runs)
	mode := "status"
	if len(m.sortModes) > 0 {
		mode = m.sortModes[m.sortIndex%len(m.sortModes)]
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		switch mode {
		case "updated":
			return sorted[i].UpdatedAt.After(sorted[j].UpdatedAt)
		case "duration":
			return durationValue(sorted[i]).Nanoseconds() > durationValue(sorted[j]).Nanoseconds()
		default:
			wi := statusWeight(sorted[i])
			wj := statusWeight(sorted[j])
			if wi == wj {
				return sorted[i].UpdatedAt.After(sorted[j].UpdatedAt)
			}
			return wi < wj
		}
	})
	return sorted
}

func statusWeight(run core.Run) int {
	switch run.Status {
	case core.RunStatusInProgress:
		return 0
	case core.RunStatusQueued:
		return 1
	case core.RunStatusFailed, core.RunStatusCancelled:
		return 2
	case core.RunStatusCompleted:
		if strings.EqualFold(run.Conclusion, "success") || strings.EqualFold(run.Conclusion, "succeeded") || run.Conclusion == "" {
			return 3
		}
		return 2
	default:
		return 4
	}
}

func durationValue(run core.Run) time.Duration {
	d := run.Duration
	if d <= 0 && !run.StartedAt.IsZero() {
		d = time.Since(run.StartedAt)
		if !run.UpdatedAt.IsZero() {
			d = run.UpdatedAt.Sub(run.StartedAt)
		}
	}
	return d
}

func formatStatusText(run core.Run) string {
	switch run.Status {
	case core.RunStatusCompleted:
		if run.Conclusion == "" {
			return "completed"
		}
		return strings.ToLower(run.Conclusion)
	case core.RunStatusInProgress:
		return "running"
	case core.RunStatusQueued:
		return "queued"
	default:
		return strings.ToLower(string(run.Status))
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

func (m *Model) cycleSort() {
	if len(m.sortModes) == 0 {
		return
	}
	m.sortIndex = (m.sortIndex + 1) % len(m.sortModes)
}

func (m *Model) startSearch() {
	m.searchActive = true
	m.searchQuery = ""
}

func (m *Model) toggleHelpOverlay() {
	m.helpVisible = !m.helpVisible
	m.helpModel.ShowAll = m.helpVisible
}

func (m *Model) toggleFocusMode() {
	m.focusMode = !m.focusMode
	m.updateTableSize()
}

func (m Model) renderDetails() string {
	lines := []string{
		sectionTitleStyle.Render("Selection"),
		bodyTextStyle.Render(fmt.Sprintf("Provider filter: %s", titleCase(m.ActiveProvider))),
		bodyTextStyle.Render(fmt.Sprintf("Status filter: %s", titleCase(m.statusFilters[m.statusIndex]))),
	}
	if m.searchQuery != "" {
		lines = append(lines, bodyTextStyle.Render(fmt.Sprintf("Search: %s", m.searchQuery)))
	}
	if len(m.visibleRuns) == 0 {
		lines = append(lines, bodyTextStyle.Render("No runs to display"))
		return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}
	idx := m.Table.Cursor()
	if idx < 0 || idx >= len(m.visibleRuns) {
		lines = append(lines, bodyTextStyle.Render("Select a run to see details"))
		return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}
	run := m.visibleRuns[idx]
	lines = append(lines,
		bodyTextStyle.Render(fmt.Sprintf("Repo: %s", run.Repo)),
		bodyTextStyle.Render(fmt.Sprintf("Workflow: %s", run.WorkflowName)),
		bodyTextStyle.Render(fmt.Sprintf("Branch: %s", run.Branch)),
		bodyTextStyle.Render(fmt.Sprintf("Commit: %s", shortSHA(run.CommitSHA))),
		bodyTextStyle.Render(fmt.Sprintf("Status: %s", formatStatusText(run))),
		bodyTextStyle.Render(fmt.Sprintf("Updated: %s", formatTimestamp(run.UpdatedAt))),
	)
	if run.URL != "" {
		lines = append(lines, bodyTextStyle.Render(fmt.Sprintf("URL: %s", run.URL)))
	}
	return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func formatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "n/a"
	}
	return t.Format("2006-01-02 15:04:05")
}

func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	diff := time.Since(t)
	if diff < time.Minute {
		secs := int(diff.Seconds())
		if secs <= 0 {
			return "now"
		}
		return fmt.Sprintf("%ds", secs)
	}
	if diff < time.Hour {
		return fmt.Sprintf("%dm", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh", int(diff.Hours()))
	}
	return fmt.Sprintf("%dd", int(diff.Hours()/24))
}

func formatDuration(duration time.Duration, start, updated time.Time) string {
	if duration <= 0 {
		if !start.IsZero() && !updated.IsZero() {
			duration = updated.Sub(start)
		} else if !start.IsZero() {
			duration = time.Since(start)
		}
	}
	if duration <= 0 {
		return "-"
	}
	return humanDuration(duration)
}

func humanDuration(d time.Duration) string {
	if d >= time.Hour {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	if d >= time.Minute {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
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

func (m Model) copySelectedURL() {
	if m.copyText == nil || len(m.visibleRuns) == 0 {
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
	m.copyText(run.URL)
}

func (m Model) renderHelp() string {
	return m.footerStyle.Render(m.helpModel.View(m.keys))
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
		return false
	}
	return true
}

func (m *Model) handleHelpOverlayInput(msg tea.KeyMsg) bool {
	if !m.helpVisible {
		return false
	}
	switch msg.String() {
	case "?", "esc":
		m.helpVisible = false
		m.helpModel.ShowAll = false
	default:
		// swallow other keys while overlay is visible
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

func (m *Model) updateTableSize() {
	if m.width <= 0 {
		return
	}
	margin := 32
	if m.focusMode || m.compactLayout() {
		margin = 8
	}
	width := m.width - margin
	if width < 40 {
		width = 40
	}
	m.Table.SetWidth(width)
}

func (m *Model) togglePalette() {
	m.paletteVisible = !m.paletteVisible
	if m.paletteVisible {
		m.paletteCursor = 0
	}
}

func (m Model) paletteItems() []string {
	items := make([]string, 0, len(m.visibleProviders))
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
		return paletteBox.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			sectionTitleStyle.Render("Provider palette"),
			bodyTextStyle.Render("(no providers)"),
		))
	}
	lines := []string{sectionTitleStyle.Render("Provider palette (space toggles, enter closes, esc cancels)")}
	for i, name := range items {
		line := fmt.Sprintf("[%s] %s", checkbox(m.visibleProviders[name]), titleCase(name))
		style := m.footerStyle
		if i == m.paletteCursor {
			style = m.headerStyle.Copy().Background(accentDarkColor).Foreground(baseTextColor)
		}
		lines = append(lines, style.Render(line))
	}
	return paletteBox.Render(strings.Join(lines, "\n"))
}

func (m Model) renderHelpOverlay() string {
	sections := []string{
		sectionTitleStyle.Render("Help & Shortcuts (press ? or esc to close)"),
		renderHelpSection("Navigation", [][2]string{
			{"↑/↓ / j/k", "Move selection"},
			{"Tab", "Cycle providers"},
			{"Enter / o", "Open selected run URL"},
		}),
		renderHelpSection("Filtering", [][2]string{
			{"f", "Cycle status filter"},
			{"t", "Cycle sort mode"},
			{"/", "Start text search"},
			{"p", "Toggle provider palette"},
		}),
		renderHelpSection("Actions", [][2]string{
			{"r", "Force refresh providers"},
			{"y", "Copy run URL"},
			{"v", "Toggle focus/dashboard view"},
			{"?", "Toggle help overlay"},
			{"q / ctrl+c", "Quit"},
		}),
	}
	return helpBox.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

func (m Model) compactLayout() bool {
	return m.width > 0 && m.width < 120
}

func renderHelpSection(title string, entries [][2]string) string {
	lines := []string{sectionTitleStyle.Render(title)}
	for _, entry := range entries {
		lines = append(lines, formatHelpEntry(entry[0], entry[1]))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func formatHelpEntry(key, desc string) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		helpKeyStyle.Render(fmt.Sprintf("%-12s", key)),
		helpDescStyle.Render(desc),
	)
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

var (
	accentColor       = lipgloss.Color("#b392f0")
	accentDarkColor   = lipgloss.Color("#342e5c")
	subtleColor       = lipgloss.Color("#8b8fb8")
	borderColor       = lipgloss.Color("#3f3c69")
	successColor      = lipgloss.Color("#43bf6d")
	warningColor      = lipgloss.Color("#f4c069")
	errorColor        = lipgloss.Color("#ff6b81")
	baseTextColor     = lipgloss.Color("#e4e5f1")
	headerStyle       = lipgloss.NewStyle().Background(accentColor).Foreground(lipgloss.Color("#0e0d19")).Bold(true).Padding(0, 2)
	footerStyle       = lipgloss.NewStyle().Foreground(subtleColor).Padding(0, 2)
	bodyBox           = lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderForeground(borderColor)
	panel             = lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderForeground(borderColor)
	tag               = lipgloss.NewStyle().Padding(0, 1).MarginRight(1).Background(accentDarkColor).Foreground(baseTextColor).Bold(true)
	tagWarn           = tag.Copy().Background(lipgloss.Color("#4f3a10"))
	tagErr            = tag.Copy().Background(lipgloss.Color("#4f1424"))
	sectionTitleStyle = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	bodyTextStyle     = lipgloss.NewStyle().Foreground(baseTextColor)
	statBoxStyle      = lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderForeground(borderColor).MarginRight(1)
	statRunningStyle  = statBoxStyle.Copy().Foreground(warningColor)
	statQueuedStyle   = statBoxStyle.Copy().Foreground(subtleColor)
	statFailedStyle   = statBoxStyle.Copy().Foreground(errorColor)
	statSuccessStyle  = statBoxStyle.Copy().Foreground(successColor)
	paletteBox        = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(borderColor).Padding(0, 1).MarginLeft(2)
	helpBox           = lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(accentColor).Padding(1, 2).MarginLeft(2)
	helpKeyStyle      = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	helpDescStyle     = lipgloss.NewStyle().Foreground(baseTextColor)
	logInfoStyle      = lipgloss.NewStyle().Foreground(baseTextColor)
	logWarnStyle      = lipgloss.NewStyle().Foreground(warningColor)
	logErrorStyle     = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
)

type logEntry struct {
	text      string
	timestamp time.Time
	level     string
}
