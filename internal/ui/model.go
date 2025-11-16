package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/joelklabo/ceye/internal/core"
	"github.com/joelklabo/ceye/internal/providers"
	githubprovider "github.com/joelklabo/ceye/internal/providers/github"
	"github.com/joelklabo/ceye/internal/providers/manager"
)

const (
	statusIconSuccess = "✓"
	statusIconFailed  = "✗"
	statusIconRunning = "▸"
	statusIconQueued  = "…"
	auditPanelLimit   = 5
)

// RunUpdatedMsg is emitted when the store receives new data.
type RunUpdatedMsg struct {
	Timestamp    time.Time
	Status       map[string]string
	Times        map[string]time.Time
	Message      string
	Level        string
	Health       map[string]core.ProviderHealth
	Lag          map[string]time.Duration
	History      map[string][]string
	Store        []manager.ProviderRecord
	Audit        []manager.StoreAuditEntry
	MissingRepos []string
}

type flashExpiredMsg struct{}

// ProviderStoreActionType describes the type of action requested in the store overlay.
type ProviderStoreActionType int

const (
	ProviderStoreActionToggle ProviderStoreActionType = iota
	ProviderStoreActionRemove
	ProviderStoreActionDuplicate
	ProviderStoreActionEdit
)

type keyMap struct {
	Provider      key.Binding
	Status        key.Binding
	Search        key.Binding
	Palette       key.Binding
	ProviderStore key.Binding
	Refresh       key.Binding
	Open          key.Binding
	Focus         key.Binding
	Sort          key.Binding
	CopyURL       key.Binding
	CopyInfo      key.Binding
	Contrast      key.Binding
	Detail        key.Binding
	Help          key.Binding
	Quit          key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Provider:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "cycle providers")),
		Status:        key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "cycle status")),
		Search:        key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Palette:       key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "provider palette")),
		ProviderStore: key.NewBinding(key.WithKeys("P"), key.WithHelp("P", "show provider store")),
		Refresh:       key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Open:          key.NewBinding(key.WithKeys("enter", "o"), key.WithHelp("enter/o", "open run")),
		Focus:         key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "toggle focus view")),
		Sort:          key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "cycle sort")),
		CopyURL:       key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy URL")),
		CopyInfo:      key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy summary")),
		Contrast:      key.NewBinding(key.WithKeys("H"), key.WithHelp("H", "toggle high contrast")),
		Detail:        key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "toggle detail view")),
		Help:          key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
		Quit:          key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Provider, k.Status, k.Sort, k.CopyURL, k.CopyInfo, k.Focus, k.Contrast, k.Refresh, k.Help, k.ProviderStore}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Provider, k.Status, k.Sort, k.Search},
		{k.Palette, k.ProviderStore, k.Focus, k.Refresh},
		{k.Open, k.CopyURL, k.CopyInfo, k.Contrast},
		{k.Help},
		{k.Quit},
	}
}

// Model represents the Bubble Tea UI state.
type Model struct {
	Store                     *core.Store
	Table                     table.Model
	ActiveProvider            string
	Providers                 []string
	visibleProviders          map[string]bool
	statusFilters             []string
	statusIndex               int
	Refresh                   func()
	openURL                   func(string)
	copyText                  func(string)
	providerStoreAction       func(manager.ProviderRecord, ProviderStoreActionType)
	missingRepoAction         func()
	helpModel                 help.Model
	keys                      keyMap
	searchActive              bool
	searchQuery               string
	paletteVisible            bool
	paletteCursor             int
	helpVisible               bool
	focusMode                 bool
	sortModes                 []string
	sortIndex                 int
	darkBackground            bool
	contrastMode              bool
	Statuses                  map[string]string
	ProviderTimes             map[string]time.Time
	ProviderLag               map[string]time.Duration
	ProviderHealth            map[string]core.ProviderHealth
	ProviderHistory           map[string][]string
	storeAudit                []manager.StoreAuditEntry
	buildInfo                 string
	detailVisible             bool
	alertLog                  []string
	visibleRuns               []core.Run
	runTotals                 map[string]int
	logEntries                []logEntry
	lastUpdate                time.Time
	flashMessage              string
	width                     int
	height                    int
	alertMessage              string
	headerStyle               lipgloss.Style
	footerStyle               lipgloss.Style
	bodyBoxStyle              lipgloss.Style
	panelStyle                lipgloss.Style
	tagStyle                  lipgloss.Style
	tagWarningStyle           lipgloss.Style
	tagErrorStyle             lipgloss.Style
	errorStyle                lipgloss.Style
	successStyle              lipgloss.Style
	failStyle                 lipgloss.Style
	runningStyle              lipgloss.Style
	providerStoreVisible      bool
	providerStoreCursor       int
	providerStoreEntries      []manager.ProviderRecord
	MissingRepos              []string
	missingIndex              int
	providerStoreEditing      bool
	providerStoreEditEntry    manager.ProviderRecord
	providerStoreTextInput    textinput.Model
	providerStoreEditField    string
	durationHistory           map[string][]time.Duration // key: "repo/workflow"
	providerStoreInstructions string
	commitCache               map[string]commitInfo // key: SHA
}

type commitInfo struct {
	SHA       string
	Message   string
	Author    string
	Timestamp time.Time
}

// NewModel constructs a UI model using default build info (empty).
func NewModel(store *core.Store, providers []string, refresh func(), openURL func(string), copyText func(string)) Model {
	return NewModelWithBuildInfo(store, providers, refresh, openURL, copyText, "")
}

// NewModelWithBuildInfo constructs a UI model and displays the provided build info in the header.
func NewModelWithBuildInfo(store *core.Store, providers []string, refresh func(), openURL func(string), copyText func(string), buildInfo string) Model {
	columns := []table.Column{
		{Title: "Provider", Width: 8},
		{Title: "Repository", Width: 18},
		{Title: "Workflow", Width: 16},
		{Title: "Status", Width: 12},
		{Title: "Branch", Width: 12},
		{Title: "Updated", Width: 8},
		{Title: "Duration", Width: 8},
	}
	tbl := table.New(table.WithColumns(columns), table.WithRows([]table.Row{}))
	tbl.Focus()
	tbl.SetStyles(tableStyles())
	providerList := buildProviderList(providers)
	statusMap := make(map[string]string)
	for _, name := range providerList {
		if name == "all" {
			continue
		}
		statusMap[name] = ""
	}
	keys := newKeyMap()
	helpModel := help.New()
	helpModel.ShowAll = false
	ti := textinput.New()
	ti.Placeholder = "display name"
	ti.CharLimit = 64
	ti.Prompt = ""
	ti.PromptStyle = lipgloss.NewStyle().Foreground(accentColor)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(accentColor)
	ti.TextStyle = lipgloss.NewStyle().Foreground(baseTextColor)
	ti.Blur()
	m := Model{
		Store:                  store,
		Table:                  tbl,
		ActiveProvider:         providerList[0],
		Providers:              providerList,
		visibleProviders:       make(map[string]bool),
		statusFilters:          []string{"all", "running", "queued", "failed", "success"},
		statusIndex:            0,
		sortModes:              []string{"status", "updated", "duration"},
		sortIndex:              0,
		darkBackground:         currentDark,
		contrastMode:           currentHigh,
		Refresh:                refresh,
		openURL:                openURL,
		copyText:               copyText,
		helpModel:              helpModel,
		keys:                   keys,
		searchActive:           false,
		searchQuery:            "",
		paletteVisible:         false,
		paletteCursor:          0,
		helpVisible:            false,
		Statuses:               statusMap,
		ProviderTimes:          make(map[string]time.Time),
		ProviderLag:            make(map[string]time.Duration),
		ProviderHealth:         make(map[string]core.ProviderHealth),
		ProviderHistory:        make(map[string][]string),
		storeAudit:             make([]manager.StoreAuditEntry, 0),
		runTotals:              make(map[string]int),
		logEntries:             make([]logEntry, 0),
		focusMode:              false,
		providerStoreVisible:   false,
		providerStoreCursor:    0,
		providerStoreEntries:   make([]manager.ProviderRecord, 0),
		MissingRepos:           []string{},
		missingIndex:           0,
		providerStoreTextInput: ti,
		buildInfo:              buildInfo,
		durationHistory:        make(map[string][]time.Duration),
		commitCache:            make(map[string]commitInfo),
		headerStyle:            headerStyle,
		footerStyle:            footerStyle,
		bodyBoxStyle:           bodyBox,
		panelStyle:             panel,
		tagStyle:               tag,
		tagWarningStyle:        tagWarn,
		tagErrorStyle:          tagErr,
		errorStyle:             lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		successStyle:           lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true),
		failStyle:              lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		runningStyle:           lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
		detailVisible:          false,
	}
	m.refreshStyles()
	m.SetProviderList(providers)
	return m
}

// SetProviderList updates the provider filter list and refreshes the table.
func (m *Model) SetProviderList(names []string) {
	providerList := buildProviderList(names)
	m.Providers = providerList
	m.visibleProviders = make(map[string]bool)
	for _, name := range providerList {
		if name == "all" {
			continue
		}
		m.visibleProviders[name] = true
	}
	if len(providerList) > 0 {
		m.ActiveProvider = providerList[0]
	}
	m.refreshTable()
}

// SetProviderStoreAction wires the action invoked when the overlay manipulates entries.
func (m *Model) SetProviderStoreAction(action func(manager.ProviderRecord, ProviderStoreActionType)) {
	m.providerStoreAction = action
}

// SetMissingRepoAction sets the callback when a missing-repo config is created.
func (m *Model) SetMissingRepoAction(action func()) {
	m.missingRepoAction = action
}

// Init implements tea.Model.Init.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.Update.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.providerStoreEditing {
		var cmd tea.Cmd
		m.providerStoreTextInput, cmd = m.providerStoreTextInput.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.Type {
			case tea.KeyEnter:
				m.finishProviderStoreEdit()
			case tea.KeyEsc:
				m.providerStoreEditing = false
				m.providerStoreTextInput.Blur()
			}
		}
		return m, cmd
	}
	if m.providerStoreVisible {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if m.handleProviderStoreInput(keyMsg) {
				return m, nil
			}
		}
	}
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
		cmd := tea.Cmd(nil)
		m.lastUpdate = msg.Timestamp
		if msg.Status != nil {
			m.Statuses = msg.Status
		}
		if msg.Times != nil {
			m.ProviderTimes = msg.Times
		}
		if msg.Health != nil {
			m.ProviderHealth = msg.Health
		}
		if msg.Lag != nil {
			m.ProviderLag = msg.Lag
		}
		if msg.History != nil {
			m.ProviderHistory = msg.History
		}
		if msg.MissingRepos != nil {
			m.MissingRepos = msg.MissingRepos
			if m.missingIndex >= len(m.MissingRepos) {
				m.missingIndex = 0
			}
		}
		if msg.Store != nil {
			m.providerStoreEntries = msg.Store
			if m.providerStoreCursor >= len(m.providerStoreEntries) {
				m.providerStoreCursor = len(m.providerStoreEntries) - 1
			}
			if m.providerStoreCursor < 0 {
				m.providerStoreCursor = 0
			}
		}
		if msg.Audit != nil {
			m.storeAudit = msg.Audit
		}
		if msg.Message != "" {
			entry := logEntry{text: msg.Message, timestamp: msg.Timestamp, level: msg.Level}
			m.logEntries = append([]logEntry{entry}, m.logEntries...)
			if len(m.logEntries) > 8 {
				m.logEntries = m.logEntries[:8]
			}
		}
		if msg.Level == "error" && msg.Message != "" {
			cmd = m.setAlert(msg.Message)
		}
		m.refreshTable()
		return m, cmd
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
		case msg.Type == tea.KeyEnter:
			m.openSelectedURL()
			return m, nil
		case msg.Type == tea.KeyRunes:
			if len(msg.Runes) == 0 {
				break
			}
			if string(msg.Runes) == "P" {
				m.toggleProviderStore()
				return m, nil
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
			case "n":
				m.cycleMissingRepo()
				return m, nil
			case "a":
				m.createMissingConfig()
				return m, nil
			case "/":
				m.startSearch()
				return m, nil
			case "o":
				m.openSelectedURL()
				return m, nil
			case "y":
				cmd := m.copySelectedURL()
				return m, cmd
			case "c":
				cmd := m.copyRunSummary()
				return m, cmd
			case "H":
				m.toggleContrast()
				return m, nil
			case "D":
				m.detailVisible = !m.detailVisible
				return m, nil
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				if idx, err := strconv.Atoi(r); err == nil && idx < len(m.Providers) {
					m.ActiveProvider = m.Providers[idx]
					m.refreshTable()
				}
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
	case alertExpiredMsg:
		m.alertMessage = ""
		return m, nil
	case flashExpiredMsg:
		m.flashMessage = ""
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
	if m.providerStoreVisible {
		view = lipgloss.JoinHorizontal(lipgloss.Top, view, m.renderProviderStore())
	}
	if m.helpVisible {
		view = lipgloss.JoinHorizontal(lipgloss.Top, view, m.renderHelpOverlay())
	}
	return appStyle.Render(view)
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
	title := "CI Status Dashboard"
	if m.buildInfo != "" {
		title = fmt.Sprintf("%s (%s)", title, m.buildInfo)
	}
	lines := []string{
		m.headerStyle.Render(fmt.Sprintf("%s  •  Last update %s", title, last)),
		m.footerStyle.Render(totals),
		m.footerStyle.Render(filters),
		m.renderProviderTabs(),
		m.renderStatusTabs(),
	}
	if m.flashMessage != "" {
		lines = append(lines, m.footerStyle.Render(m.flashMessage))
	}
	if m.alertMessage != "" {
		lines = append(lines, alertStyle.Render(m.alertMessage))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderRunsTable() string {
	title := sectionTitleStyle.Render(fmt.Sprintf("Runs (%d showing)", len(m.visibleRuns)))
	return lipgloss.JoinVertical(lipgloss.Left, title, m.bodyBoxStyle.Render(m.Table.View()))
}

func (m Model) renderDashboardBody() string {
	table := m.renderRunsTable()
	sidebarParts := []string{
		m.renderDetails(),
		m.renderDetailView(),
	}
	
	// Add commit details panel (for selected run)
	if commitDetails := m.renderCommitDetailsPanel(); commitDetails != "" {
		sidebarParts = append(sidebarParts, commitDetails)
	}
	
	// Add active runs panel
	if activeRuns := m.renderActiveRunsPanel(); activeRuns != "" {
		sidebarParts = append(sidebarParts, activeRuns)
	}
	
	// Add provider health panel
	if providerHealth := m.renderProviderHealthPanel(); providerHealth != "" {
		sidebarParts = append(sidebarParts, providerHealth)
	}
	
	// Add failure rate panel
	if failureRate := m.renderFailureRatePanel(); failureRate != "" {
		sidebarParts = append(sidebarParts, failureRate)
	}
	
	// Add duration trends panel
	if durationTrends := m.renderDurationTrendsPanel(); durationTrends != "" {
		sidebarParts = append(sidebarParts, durationTrends)
	}
	
	sidebarParts = append(sidebarParts, m.renderLogs())
	
	if audit := m.renderAuditPanel(); audit != "" {
		sidebarParts = append(sidebarParts, audit)
	}
	sidebarParts = append(sidebarParts, m.renderHistoryPanel(), m.renderAlertLog())
	sidebar := lipgloss.JoinVertical(lipgloss.Left, sidebarParts...)
	if missing := m.renderMissingPanel(); missing != "" {
		sidebar = lipgloss.JoinVertical(lipgloss.Left, sidebar, missing)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, table, sidebar)
}

func (m Model) renderCompactBody() string {
	history := m.renderHistoryPanel()
	alertLog := m.renderAlertLog()
	sidebar := []string{m.renderDetails(), m.renderDetailView(), m.renderLogs()}
	if audit := m.renderAuditPanel(); audit != "" {
		sidebar = append(sidebar, audit)
	}
	if history != "" {
		sidebar = append(sidebar, history)
	}
	if alertLog != "" {
		sidebar = append(sidebar, alertLog)
	}
	if missing := m.renderMissingPanel(); missing != "" {
		sidebar = append(sidebar, missing)
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderStatuses(),
		m.renderRunsTable(),
		lipgloss.JoinVertical(lipgloss.Left, sidebar...),
	)
}

func (m Model) renderFocusBody() string {
	banner := m.panelStyle.Render(bodyTextStyle.Render("Focus mode: table maximized (press 'v' to return to dashboard view)"))
	lower := lipgloss.JoinHorizontal(lipgloss.Top, m.renderDetails(), m.renderLogs())
	body := []string{banner, m.renderRunsTable()}
	auditPanel := m.renderAuditPanel()
	if m.compactLayout() {
		body = append(body, m.renderDetails(), m.renderLogs())
		if auditPanel != "" {
			body = append(body, auditPanel)
		}
		body = append(body, m.renderStatuses())
	} else {
		body = append(body, lower)
		if auditPanel != "" {
			body = append(body, auditPanel)
		}
		body = append(body, m.renderStatuses())
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

func (m Model) renderMissingPanel() string {
	if len(m.MissingRepos) == 0 {
		return ""
	}
	lines := []string{
		sectionTitleStyle.Render("Missing configs"),
		bodyTextStyle.Render("Press n to cycle, a to create config"),
	}
	for i, repo := range m.MissingRepos {
		prefix := " "
		if i == m.missingIndex {
			prefix = ">"
		}
		lines = append(lines, bodyTextStyle.Render(fmt.Sprintf("%s %s", prefix, filepath.Base(repo))))
	}
	return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderActiveRunsPanel() string {
	activeRuns := []core.Run{}
	for _, run := range m.visibleRuns {
		if run.Status == core.RunStatusInProgress || run.Status == core.RunStatusQueued {
			activeRuns = append(activeRuns, run)
		}
	}
	
	if len(activeRuns) == 0 {
		return ""
	}
	
	// Limit to 5 most recent
	if len(activeRuns) > 5 {
		activeRuns = activeRuns[:5]
	}
	
	lines := []string{
		sectionTitleStyle.Render(fmt.Sprintf("Running Now (%d)", len(activeRuns))),
	}
	
	for _, run := range activeRuns {
		elapsed := time.Since(run.StartedAt)
		if run.StartedAt.IsZero() {
			elapsed = time.Since(run.UpdatedAt)
		}
		
		icon := statusIconRunning
		if run.Status == core.RunStatusQueued {
			icon = statusIconQueued
		}
		
		repoShort := truncate(run.Repo, 15)
		workflowShort := truncate(run.WorkflowName, 15)
		
		line := fmt.Sprintf("%s %s/%s %s",
			icon,
			repoShort,
			workflowShort,
			formatDuration(elapsed, run.StartedAt, time.Now()),
		)
		lines = append(lines, bodyTextStyle.Render(line))
	}
	
	return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderCommitDetailsPanel() string {
	idx := m.Table.Cursor()
	if idx < 0 || idx >= len(m.visibleRuns) {
		return ""
	}
	
	run := m.visibleRuns[idx]
	if run.CommitSHA == "" {
		return ""
	}
	
	lines := []string{
		sectionTitleStyle.Render("Commit Details"),
	}
	
	// Check cache for commit info
	commit, hasCached := m.commitCache[run.CommitSHA]
	
	if hasCached {
		// Show cached commit details
		shaShort := shortSHA(commit.SHA)
		author := truncate(commit.Author, 15)
		timeAgo := formatRelativeTime(commit.Timestamp)
		
		lines = append(lines,
			bodyTextStyle.Render(fmt.Sprintf("%s - %s (%s)", shaShort, author, timeAgo)),
		)
		
		// Truncate commit message to fit
		message := truncate(commit.Message, 28)
		lines = append(lines, bodyTextStyle.Render(fmt.Sprintf("\"%s\"", message)))
	} else {
		// Show basic info from run
		shaShort := shortSHA(run.CommitSHA)
		timeAgo := formatRelativeTime(run.UpdatedAt)
		
		lines = append(lines,
			bodyTextStyle.Render(fmt.Sprintf("%s (%s)", shaShort, timeAgo)),
		)
		lines = append(lines, bodyTextStyle.Render("(commit details not fetched)"))
	}
	
	return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderProviderHealthPanel() string {
	if len(m.Providers) == 0 {
		return ""
	}
	
	lines := []string{
		sectionTitleStyle.Render("Provider Health"),
	}
	
	for _, name := range m.Providers {
		health, hasHealth := m.ProviderHealth[name]
		lag, hasLag := m.ProviderLag[name]
		
		// Determine health status
		statusIcon := "✓"
		statusColor := m.successStyle
		statusText := "healthy"
		
		if hasHealth && health.ErrorCount > 0 {
			statusIcon = "✗"
			statusColor = m.errorStyle
			statusText = "errors"
		} else if hasLag && lag > 2*time.Second {
			statusIcon = "⚠"
			statusColor = m.tagWarningStyle
			statusText = "slow"
		}
		
		// Provider name line with status
		providerLine := fmt.Sprintf("%s %s: %s",
			statusIcon,
			truncate(name, 12),
			statusText,
		)
		lines = append(lines, statusColor.Render(providerLine))
		
		// Details line
		var details []string
		if hasLag {
			lagStr := "Lag: " + lag.Round(time.Millisecond).String()
			details = append(details, lagStr)
		}
		if hasHealth && health.ErrorCount > 0 {
			errStr := fmt.Sprintf("Errors: %d", health.ErrorCount)
			if !health.LastError.IsZero() {
				errStr += fmt.Sprintf(" (%s ago)", time.Since(health.LastError).Round(time.Second))
			}
			details = append(details, errStr)
		}
		
		if len(details) > 0 {
			detailLine := "  " + strings.Join(details, "  ")
			lines = append(lines, bodyTextStyle.Render(truncate(detailLine, 30)))
		}
	}
	
	return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderFailureRatePanel() string {
	// Aggregate runs by repo
	type repoStats struct {
		repo         string
		successCount int
		failureCount int
		totalCount   int
	}
	
	statsMap := make(map[string]*repoStats)
	
	// Count last 20 runs per repo (or all visible runs if less)
	repoRunCounts := make(map[string]int)
	for _, run := range m.visibleRuns {
		// Only count completed runs
		if run.Status == core.RunStatusCompleted {
			repoRunCounts[run.Repo]++
			if repoRunCounts[run.Repo] > 20 {
				continue // Only consider last 20 runs per repo
			}
			
			stats, exists := statsMap[run.Repo]
			if !exists {
				stats = &repoStats{repo: run.Repo}
				statsMap[run.Repo] = stats
			}
			
			stats.totalCount++
			// Check if success
			if strings.EqualFold(run.Conclusion, "success") || strings.EqualFold(run.Conclusion, "succeeded") || run.Conclusion == "" {
				stats.successCount++
			} else {
				stats.failureCount++
			}
		}
	}
	
	if len(statsMap) == 0 {
		return ""
	}
	
	// Convert to slice and sort by failure rate (worst first)
	statsList := make([]repoStats, 0, len(statsMap))
	for _, stats := range statsMap {
		statsList = append(statsList, *stats)
	}
	sort.Slice(statsList, func(i, j int) bool {
		rateI := float64(statsList[i].successCount) / float64(statsList[i].totalCount)
		rateJ := float64(statsList[j].successCount) / float64(statsList[j].totalCount)
		return rateI < rateJ // Lower success rate (higher failure rate) first
	})
	
	// Limit to top 5 repos
	if len(statsList) > 5 {
		statsList = statsList[:5]
	}
	
	lines := []string{
		sectionTitleStyle.Render("Success Rates"),
	}
	
	for _, stats := range statsList {
		successRate := float64(stats.successCount) / float64(stats.totalCount) * 100
		icon := statusIconSuccess
		statusColor := m.successStyle
		
		if successRate < 50 {
			icon = statusIconFailed
			statusColor = m.errorStyle
		} else if successRate < 80 {
			icon = "⚠"
			statusColor = m.tagWarningStyle
		}
		
		repoShort := truncate(stats.repo, 12)
		line := fmt.Sprintf("%s: %3.0f%% %s (%d/%d)",
			repoShort,
			successRate,
			icon,
			stats.successCount,
			stats.totalCount,
		)
		lines = append(lines, statusColor.Render(line))
	}
	
	return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderDurationTrendsPanel() string {
	if len(m.durationHistory) == 0 {
		return ""
	}
	
	type workflowTrend struct {
		key        string
		avgCurrent time.Duration
		avgPrev    time.Duration
		trend      string
		pctChange  float64
	}
	
	trends := []workflowTrend{}
	
	for key, history := range m.durationHistory {
		if len(history) < 2 {
			continue // Need at least 2 data points
		}
		
		// Calculate current average (last 5 runs)
		currentCount := 5
		if len(history) < 5 {
			currentCount = len(history)
		}
		currentRuns := history[len(history)-currentCount:]
		var currentSum time.Duration
		for _, d := range currentRuns {
			currentSum += d
		}
		avgCurrent := currentSum / time.Duration(len(currentRuns))
		
		// Calculate previous average (runs before the last 5)
		prevStart := 0
		prevEnd := len(history) - currentCount
		if prevEnd <= 0 {
			continue // Not enough data for trend
		}
		prevRuns := history[prevStart:prevEnd]
		var prevSum time.Duration
		for _, d := range prevRuns {
			prevSum += d
		}
		avgPrev := prevSum / time.Duration(len(prevRuns))
		
		// Calculate trend
		pctChange := float64(avgCurrent-avgPrev) / float64(avgPrev) * 100
		trendIcon := "→"
		if pctChange > 5 {
			trendIcon = "↑"
		} else if pctChange < -5 {
			trendIcon = "↓"
		}
		
		trends = append(trends, workflowTrend{
			key:        key,
			avgCurrent: avgCurrent,
			avgPrev:    avgPrev,
			trend:      trendIcon,
			pctChange:  pctChange,
		})
	}
	
	if len(trends) == 0 {
		return ""
	}
	
	// Sort by run count (most active first)
	sort.Slice(trends, func(i, j int) bool {
		return len(m.durationHistory[trends[i].key]) > len(m.durationHistory[trends[j].key])
	})
	
	// Limit to top 5
	if len(trends) > 5 {
		trends = trends[:5]
	}
	
	lines := []string{
		sectionTitleStyle.Render("Duration Trends"),
	}
	
	for _, trend := range trends {
		// Extract workflow name from key
		parts := strings.Split(trend.key, "/")
		workflowName := trend.key
		if len(parts) >= 2 {
			workflowName = parts[len(parts)-1]
		}
		workflowShort := truncate(workflowName, 10)
		
		avgStr := formatDuration(trend.avgCurrent, time.Time{}, time.Time{})
		changeStr := fmt.Sprintf("%s %.0f%%", trend.trend, trend.pctChange)
		if trend.pctChange > 0 {
			changeStr = fmt.Sprintf("%s +%.0f%%", trend.trend, trend.pctChange)
		}
		
		line := fmt.Sprintf("%s: avg %s %s",
			workflowShort,
			avgStr,
			changeStr,
		)
		
		// Color code based on trend
		if trend.trend == "↑" {
			line = m.tagWarningStyle.Render(line)
		} else if trend.trend == "↓" {
			line = m.successStyle.Render(line)
		} else {
			line = bodyTextStyle.Render(line)
		}
		
		lines = append(lines, line)
	}
	
	return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
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
		if health, ok := m.ProviderHealth[name]; ok && health.ErrorCount > 0 {
			label = fmt.Sprintf("%s (%d errs)", label, health.ErrorCount)
			style = m.tagErrorStyle
		} else if health.LastSuccess.After(time.Time{}) {
			label = fmt.Sprintf("%s [%s]", label, health.LastSuccess.Format("15:04:05"))
		}
		if lag, ok := m.ProviderLag[name]; ok && lag > 0 {
			label = fmt.Sprintf("%s • lag %s", label, formatLag(lag))
			if lag > 20*time.Second {
				style = m.tagWarningStyle
			}
		}
		parts = append(parts, style.Render(label))
	}
	return parts
}

func (m Model) renderProviderTabs() string {
	tabs := make([]string, 0, len(m.Providers))
	for i, name := range m.Providers {
		if name == "all" {
			continue
		}
		style := m.tagStyle
		if name == m.ActiveProvider {
			style = m.headerStyle.Copy().Foreground(baseTextColor)
			tabs = append(tabs, style.Render(fmt.Sprintf("[%d] %s", i, titleCase(name))))
		} else {
			tabs = append(tabs, style.Render(fmt.Sprintf("[%d] %s", i, titleCase(name))))
		}
	}
	if len(tabs) == 0 {
		return ""
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func (m Model) renderStatusTabs() string {
	tabs := make([]string, 0, len(m.statusFilters))
	for i, status := range m.statusFilters {
		label := titleCase(status)
		style := m.tagStyle
		if i == m.statusIndex {
			style = m.tagErrorStyle
		}
		tabs = append(tabs, style.Render(label))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func formatLag(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
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

func (m Model) renderAlertLog() string {
	if len(m.alertLog) == 0 {
		return ""
	}
	lines := []string{sectionTitleStyle.Render("Alert log")}
	for _, entry := range m.alertLog {
		lines = append(lines, logErrorStyle.Render(entry))
	}
	return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderAuditPanel() string {
	if len(m.storeAudit) == 0 {
		return ""
	}
	lines := []string{sectionTitleStyle.Render("Store audit")}
	for i, entry := range m.storeAudit {
		if i >= auditPanelLimit {
			break
		}
		label := entry.DisplayName
		if label == "" {
			label = entry.ProviderType
		}
		if label == "" {
			label = shortID(entry.ID)
		}
		actor := entry.Actor
		if actor == "" {
			actor = "system"
		}
		action := titleCase(entry.Action)
		if action == "" {
			action = "action"
		}
		line := fmt.Sprintf("[%s] %s %s %s", entry.Timestamp.Format("15:04:05"), actor, action, label)
		if entry.Details != "" {
			line = fmt.Sprintf("%s • %s", line, entry.Details)
		}
		lines = append(lines, bodyTextStyle.Render(line))
	}
	return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderHistoryPanel() string {
	lines := []string{sectionTitleStyle.Render("History")}
	added := false
	for _, name := range m.Providers {
		if name == "all" {
			continue
		}
		runs := m.ProviderHistory[name]
		if len(runs) == 0 {
			continue
		}
		added = true
		lines = append(lines, m.footerStyle.Render(titleCase(name)))
		for _, entry := range runs {
			lines = append(lines, bodyTextStyle.Render(entry))
		}
	}
	if !added {
		return ""
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
			truncate(strings.ToLower(run.Provider), 8),
			truncate(run.Repo, 18),
			truncate(run.WorkflowName, 16),
			m.formatStatus(run),
			truncate(formatBranchName(run.Branch), 12),
			formatRelativeTime(run.UpdatedAt),
			formatDuration(run.Duration, run.StartedAt, run.UpdatedAt),
		})
	}
	m.visibleRuns = sorted
	
	// Track duration history for completed runs
	for _, run := range sorted {
		if run.Status == core.RunStatusCompleted && run.Duration > 0 {
			key := run.Repo + "/" + run.WorkflowName
			history := m.durationHistory[key]
			
			// Check if this run is already tracked (by checking if latest duration matches)
			alreadyTracked := false
			if len(history) > 0 && history[len(history)-1] == run.Duration {
				alreadyTracked = true
			}
			
			if !alreadyTracked {
				history = append(history, run.Duration)
				// Keep only last 10 runs for each workflow
				if len(history) > 10 {
					history = history[len(history)-10:]
				}
				m.durationHistory[key] = history
			}
		}
	}
	
	m.Table.SetRows(rows)
}

func (m *Model) formatStatus(run core.Run) string {
	var icon, text string
	
	switch run.Status {
	case core.RunStatusCompleted:
		conclusion := strings.ToLower(run.Conclusion)
		switch conclusion {
		case "", "success", "succeeded":
			icon, text = statusIconSuccess, "OK"
		case "cancelled", "canceled":
			icon, text = statusIconFailed, "cancel"
		default:
			icon, text = statusIconFailed, conclusion
		}
	case core.RunStatusInProgress:
		icon, text = statusIconRunning, "run"
	case core.RunStatusQueued:
		icon, text = statusIconQueued, "queue"
	case core.RunStatusFailed, core.RunStatusCancelled:
		icon, text = statusIconFailed, "fail"
	default:
		icon, text = statusIconRunning, strings.ToLower(string(run.Status))
	}
	
	// Truncate long status text
	if len(text) > 8 {
		text = text[:7] + "…"
	}
	
	return fmt.Sprintf("%s %s", icon, text)
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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-1] + "…"
}

func formatBranchName(branch string) string {
	if branch == "" {
		return branch
	}
	blower := strings.ToLower(branch)
	switch {
	case blower == "main" || blower == "master":
		return branchMainStyle.Render(branch)
	case blower == "develop" || blower == "dev":
		return branchDevStyle.Render(branch)
	case strings.HasPrefix(blower, "release"):
		return branchReleaseStyle.Render(branch)
	default:
		return branchDefaultStyle.Render(branch)
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

func (m *Model) setFlash(message string) tea.Cmd {
	m.flashMessage = message
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return flashExpiredMsg{}
	})
}

func (m *Model) setAlert(message string) tea.Cmd {
	m.alertMessage = message
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return alertExpiredMsg{}
	})
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

func (m Model) renderDetailView() string {
	if !m.detailVisible {
		return ""
	}
	idx := m.Table.Cursor()
	if idx < 0 || idx >= len(m.visibleRuns) {
		return m.panelStyle.Render(bodyTextStyle.Render("Select a run to see details"))
	}
	run := m.visibleRuns[idx]
	lines := []string{
		sectionTitleStyle.Render("Detail View"),
		bodyTextStyle.Render(fmt.Sprintf("Status: %s", formatStatusText(run))),
		bodyTextStyle.Render(fmt.Sprintf("Duration: %s", formatDuration(run.Duration, run.StartedAt, run.UpdatedAt))),
		bodyTextStyle.Render(fmt.Sprintf("Updated: %s", formatTimestamp(run.UpdatedAt))),
		bodyTextStyle.Render(fmt.Sprintf("Workflow: %s", run.WorkflowName)),
		bodyTextStyle.Render(fmt.Sprintf("Branch: %s", run.Branch)),
		bodyTextStyle.Render(fmt.Sprintf("Commit: %s", shortSHA(run.CommitSHA))),
		bodyTextStyle.Render(fmt.Sprintf("URL: %s", run.URL)),
	}
	return m.panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
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

func (m *Model) copySelectedURL() tea.Cmd {
	if m.copyText == nil || len(m.visibleRuns) == 0 {
		return nil
	}
	idx := m.Table.Cursor()
	if idx < 0 || idx >= len(m.visibleRuns) {
		return nil
	}
	run := m.visibleRuns[idx]
	if run.URL == "" {
		return nil
	}
	m.copyText(run.URL)
	return m.setFlash("Copied run URL")
}

func (m *Model) copyRunSummary() tea.Cmd {
	if m.copyText == nil || len(m.visibleRuns) == 0 {
		return nil
	}
	idx := m.Table.Cursor()
	if idx < 0 || idx >= len(m.visibleRuns) {
		return nil
	}
	run := m.visibleRuns[idx]
	summary := fmt.Sprintf("%s • %s • %s • %s • %s",
		strings.ToUpper(run.Provider),
		run.Repo,
		run.Branch,
		fmt.Sprintf("%s (%s)", run.WorkflowName, formatStatusText(run)),
		run.URL,
	)
	m.copyText(summary)
	return m.setFlash("Copied run summary")
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

func (m *Model) handleProviderStoreInput(msg tea.KeyMsg) bool {
	if !m.providerStoreVisible {
		return false
	}
	switch msg.String() {
	case "up", "k":
		m.moveProviderStoreCursor(-1)
	case "down", "j":
		m.moveProviderStoreCursor(1)
	case " ", "enter":
		if len(m.providerStoreEntries) == 0 || m.providerStoreAction == nil {
			return true
		}
		index := m.providerStoreCursor
		if index < 0 || index >= len(m.providerStoreEntries) {
			return true
		}
		entry := m.providerStoreEntries[index]
		target := !entry.Enabled
		m.providerStoreEntries[index].Enabled = target
		entry.Enabled = target
		m.providerStoreAction(entry, ProviderStoreActionToggle)
	case "E":
		if len(m.providerStoreEntries) == 0 {
			return true
		}
		index := m.providerStoreCursor
		if index < 0 || index >= len(m.providerStoreEntries) {
			return true
		}
		entry := m.providerStoreEntries[index]
		m.startProviderStoreEdit(entry)
		return true
	case "e":
		if len(m.providerStoreEntries) == 0 || m.providerStoreAction == nil {
			return true
		}
		index := m.providerStoreCursor
		if index < 0 || index >= len(m.providerStoreEntries) {
			return true
		}
		entry := m.providerStoreEntries[index]
		m.providerStoreAction(entry, ProviderStoreActionDuplicate)
	case "d":
		if len(m.providerStoreEntries) == 0 || m.providerStoreAction == nil {
			return true
		}
		index := m.providerStoreCursor
		if index < 0 || index >= len(m.providerStoreEntries) {
			return true
		}
		entry := m.providerStoreEntries[index]
		m.providerStoreAction(entry, ProviderStoreActionRemove)
	case "P", "esc":
		m.providerStoreVisible = false
	default:
		return false
	}
	return true
}

func (m *Model) cycleMissingRepo() {
	if len(m.MissingRepos) == 0 {
		return
	}
	m.missingIndex = (m.missingIndex + 1) % len(m.MissingRepos)
}

func (m *Model) createMissingConfig() {
	if len(m.MissingRepos) == 0 {
		m.flashMessage = "no missing configs to create"
		return
	}
	index := m.missingIndex
	repoPath := m.MissingRepos[index]
	cfgPath := filepath.Join(repoPath, "ceye.yaml")
	if _, err := os.Stat(cfgPath); err == nil {
		m.flashMessage = fmt.Sprintf("config already exists at %s", cfgPath)
		return
	}
	owner, repo := guessOwnerRepo(repoPath)
	if repo == "" {
		repo = filepath.Base(repoPath)
	}
	if owner == "" {
		owner = "OWNER"
	}
	content := fmt.Sprintf("providers:\n  - type: github\n    repos:\n      - owner: \"%s\"\n        repo: \"%s\"\n", owner, repo)
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		m.flashMessage = fmt.Sprintf("write config: %v", err)
		return
	}
	m.flashMessage = fmt.Sprintf("created %s", cfgPath)
	m.MissingRepos = append(m.MissingRepos[:index], m.MissingRepos[index+1:]...)
	if m.missingIndex >= len(m.MissingRepos) {
		m.missingIndex = 0
	}
	if m.missingRepoAction != nil {
		m.missingRepoAction()
	}
}

func guessOwnerRepo(path string) (string, string) {
	cmd := exec.Command("git", "-C", path, "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return "", filepath.Base(path)
	}
	url := strings.TrimSpace(string(out))
	url = strings.TrimSuffix(url, ".git")
	if idx := strings.Index(url, ":"); idx != -1 && strings.Contains(url[:idx], "@") {
		url = url[idx+1:]
	}
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	segments := strings.Split(url, "/")
	if len(segments) >= 2 {
		return segments[len(segments)-2], segments[len(segments)-1]
	}
	return "", filepath.Base(path)
}

func (m *Model) startProviderStoreEdit(entry manager.ProviderRecord) {
	m.providerStoreEditing = true
	m.providerStoreEditEntry = entry
	value := ""
	m.providerStoreInstructions = ""
	switch entry.Config.Type {
	case "github":
		m.providerStoreEditField = "owner/repo"
		m.providerStoreInstructions = "Enter owner/repo (e.g. octocat/hello-world)"
		if len(entry.Config.Repos) > 0 {
			value = fmt.Sprintf("%s/%s", entry.Config.Repos[0].Owner, entry.Config.Repos[0].Repo)
		}
	case "azure":
		m.providerStoreEditField = "org/project[:pipelines]"
		m.providerStoreInstructions = "Format org/project[:comma-separated pipelines]"
		if entry.Config.Org != "" || entry.Config.Project != "" {
			value = fmt.Sprintf("%s/%s", entry.Config.Org, entry.Config.Project)
			if len(entry.Config.Pipelines) > 0 {
				parts := make([]string, len(entry.Config.Pipelines))
				for i, id := range entry.Config.Pipelines {
					parts[i] = fmt.Sprintf("%d", id)
				}
				value = fmt.Sprintf("%s:%s", value, strings.Join(parts, ","))
			}
		}
	case "gitlab":
		m.providerStoreEditField = "gitlab_project"
		m.providerStoreInstructions = "Enter the GitLab project path (e.g. org/repo)"
		value = entry.Config.GitLabProject
	default:
		m.providerStoreEditField = "display name"
		m.providerStoreInstructions = "Enter a friendly display name"
		if entry.Config.DisplayName != "" {
			value = entry.Config.DisplayName
		} else {
			value = entry.Config.Type + " provider"
		}
	}
	m.providerStoreTextInput.Placeholder = m.providerStoreEditField
	if value == "" {
		m.providerStoreTextInput.SetValue("")
		m.providerStoreTextInput.Blur()
	} else {
		m.providerStoreTextInput.SetValue(value)
		m.providerStoreTextInput.CursorEnd()
	}
	m.providerStoreTextInput.Focus()
}

func (m *Model) finishProviderStoreEdit() {
	if !m.providerStoreEditing {
		return
	}
	entry := m.providerStoreEditEntry
	value := strings.TrimSpace(m.providerStoreTextInput.Value())
	if value == "" {
		m.flashMessage = "Value cannot be empty"
		return
	}
	switch entry.Config.Type {
	case "github":
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			m.flashMessage = "GitHub value must be owner/repo"
			return
		}
		entry.Config.Repos = []githubprovider.RepoConfig{{Owner: parts[0], Repo: parts[1]}}
	case "azure":
		parts := strings.SplitN(value, ":", 2)
		project := strings.SplitN(parts[0], "/", 2)
		if len(project) != 2 {
			m.flashMessage = "Azure value must be org/project"
			return
		}
		entry.Config.Org = project[0]
		entry.Config.Project = project[1]
		entry.Config.Pipelines = nil
		if len(parts) == 2 && parts[1] != "" {
			keys := strings.Split(parts[1], ",")
			for _, key := range keys {
				key = strings.TrimSpace(key)
				if key == "" {
					continue
				}
				var id int
				fmt.Sscanf(key, "%d", &id)
				if id > 0 {
					entry.Config.Pipelines = append(entry.Config.Pipelines, id)
				}
			}
		}
	case "gitlab":
		entry.Config.GitLabProject = value
	default:
		entry.Config.DisplayName = value
	}
	m.providerStoreEditing = false
	m.providerStoreTextInput.Blur()
	m.providerStoreInstructions = ""
	if m.providerStoreAction != nil {
		m.providerStoreAction(entry, ProviderStoreActionEdit)
	}
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

func (m *Model) moveProviderStoreCursor(delta int) {
	count := len(m.providerStoreEntries)
	if count == 0 {
		return
	}
	m.providerStoreCursor += delta
	if m.providerStoreCursor < 0 {
		m.providerStoreCursor = count - 1
	}
	if m.providerStoreCursor >= count {
		m.providerStoreCursor = 0
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
	
	// Reserve space for sidebar in normal dashboard mode
	sidebarWidth := 45
	margin := 8
	
	width := m.width - sidebarWidth - margin
	
	if m.focusMode {
		// In focus mode, use more width since sidebar is below
		width = m.width - margin
	} else if m.compactLayout() {
		// Compact mode also uses full width
		width = m.width - margin
	}
	
	if width < 40 {
		width = 40
	}
	m.Table.SetWidth(width)
	
	// Set table height based on available space
	headerHeight := 15 // header + stats + filters
	footerHeight := 2  // help line
	if m.height > 0 {
		availableHeight := m.height - headerHeight - footerHeight
		if m.focusMode {
			availableHeight = m.height - 8
		} else if !m.compactLayout() {
			availableHeight = m.height - 18
		}
		if availableHeight < 5 {
			availableHeight = 5
		}
		m.Table.SetHeight(availableHeight)
	}
}

func (m *Model) togglePalette() {
	m.paletteVisible = !m.paletteVisible
	if m.paletteVisible {
		m.paletteCursor = 0
	}
}

func (m *Model) toggleProviderStore() {
	m.providerStoreVisible = !m.providerStoreVisible
	if m.providerStoreVisible {
		m.providerStoreCursor = 0
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

func (m Model) renderProviderStore() string {
	lines := []string{sectionTitleStyle.Render("Provider store (press P to close)")}
	if m.providerStoreEditing {
		header := fmt.Sprintf("Editing %s (enter to save, esc to cancel)", shortID(m.providerStoreEditEntry.ID))
		lines = append(lines, bodyTextStyle.Render(header))
		if m.providerStoreInstructions != "" {
			lines = append(lines, bodyTextStyle.Render(m.providerStoreInstructions))
		}
		lines = append(lines, storeEntrySelected.Render(m.providerStoreTextInput.View()))
		return storeBox.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}
	if len(m.providerStoreEntries) == 0 {
		lines = append(lines, bodyTextStyle.Render("No stored providers"))
		return storeBox.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}
	for i, entry := range m.providerStoreEntries {
		name := entry.Config.DisplayName
		if name == "" {
			name = fmt.Sprintf("%s provider", entry.Config.Type)
		}
		if name == "" {
			name = entry.Config.Type
		}
		status := "disabled"
		if entry.Enabled {
			status = "enabled"
		}
		line := fmt.Sprintf("%s • %s • %s", shortID(entry.ID), name, status)
		style := storeEntryStyle
		if i == m.providerStoreCursor {
			style = storeEntrySelected
		}
		lines = append(lines, style.Render(line))
		if detail := providers.StoreDetail(entry.Config); detail != "" {
			lines = append(lines, bodyTextStyle.Foreground(subtleColor).Render(detail))
		}
	}
	return storeBox.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
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
			{"P", "View stored providers"},
			{"n", "Cycle missing repo entries"},
			{"a", "Create config for highlighted repo"},
			{"E", "Edit stored provider display name"},
		}),
		renderHelpSection("Actions", [][2]string{
			{"r", "Force refresh providers"},
			{"y", "Copy run URL"},
			{"c", "Copy run summary"},
			{"H", "Toggle high contrast"},
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
	accentColor        lipgloss.Color
	accentDarkColor    lipgloss.Color
	appBackground      lipgloss.Color
	subtleColor        lipgloss.Color
	borderColor        lipgloss.Color
	successColor       lipgloss.Color
	warningColor       lipgloss.Color
	errorColor         lipgloss.Color
	baseTextColor      lipgloss.Color
	headerStyle        lipgloss.Style
	footerStyle        lipgloss.Style
	bodyBox            lipgloss.Style
	panel              lipgloss.Style
	tag                lipgloss.Style
	tagWarn            lipgloss.Style
	tagErr             lipgloss.Style
	sectionTitleStyle  lipgloss.Style
	bodyTextStyle      lipgloss.Style
	statBoxStyle       lipgloss.Style
	statRunningStyle   lipgloss.Style
	statQueuedStyle    lipgloss.Style
	statFailedStyle    lipgloss.Style
	statSuccessStyle   lipgloss.Style
	paletteBox         lipgloss.Style
	helpBox            lipgloss.Style
	helpKeyStyle       lipgloss.Style
	helpDescStyle      lipgloss.Style
	logInfoStyle       lipgloss.Style
	logWarnStyle       lipgloss.Style
	logErrorStyle      lipgloss.Style
	branchMainStyle    lipgloss.Style
	branchDevStyle     lipgloss.Style
	branchReleaseStyle lipgloss.Style
	branchDefaultStyle lipgloss.Style
	rowHighlightBg     lipgloss.Color
	appStyle           lipgloss.Style
	currentDark        bool
	currentHigh        bool
	alertStyle         lipgloss.Style
	storeBox           lipgloss.Style
	storeEntryStyle    lipgloss.Style
	storeEntrySelected lipgloss.Style
)

type logEntry struct {
	text      string
	timestamp time.Time
	level     string
}

type alertExpiredMsg struct{}

func init() {
	setTheme(lipgloss.HasDarkBackground(), false)
}

func setTheme(dark, high bool) {
	currentDark = dark
	currentHigh = high
	applyTheme(dark, high)
}

func applyTheme(dark, high bool) {
	if dark {
		accentColor = lipgloss.Color("#b392f0")
		accentDarkColor = lipgloss.Color("#342e5c")
		appBackground = lipgloss.Color("#141225")
		subtleColor = lipgloss.Color("#8b8fb8")
		borderColor = lipgloss.Color("#3f3c69")
		successColor = lipgloss.Color("#43bf6d")
		warningColor = lipgloss.Color("#f4c069")
		errorColor = lipgloss.Color("#ff6b81")
		baseTextColor = lipgloss.Color("#e4e5f1")
		rowHighlightBg = lipgloss.Color("#373257")
	} else {
		accentColor = lipgloss.Color("#36236b")
		accentDarkColor = lipgloss.Color("#cec6f2")
		appBackground = lipgloss.Color("#0e0d1a")
		subtleColor = lipgloss.Color("#4a4370")
		borderColor = lipgloss.Color("#3b3660")
		successColor = lipgloss.Color("#065229")
		warningColor = lipgloss.Color("#8a5b00")
		errorColor = lipgloss.Color("#8b0b2b")
		baseTextColor = lipgloss.Color("#0b0a15")
		rowHighlightBg = lipgloss.Color("#d9d4f5")
	}

	if high {
		accentColor = lipgloss.Color("#ffffff")
		accentDarkColor = lipgloss.Color("#bfbcff")
		borderColor = lipgloss.Color("#ffffff")
		appBackground = lipgloss.Color("#0b0b0b")
		baseTextColor = lipgloss.Color("#ffffff")
	}

	headerStyle = lipgloss.NewStyle().Background(accentColor).Foreground(lipgloss.Color("#0e0d19")).Bold(true).Padding(0, 2)
	if !dark {
		headerStyle = lipgloss.NewStyle().Background(accentColor).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 2)
	}
	footerStyle = lipgloss.NewStyle().Foreground(subtleColor).Padding(0, 2)
	bodyBox = lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderForeground(borderColor)
	panel = lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderForeground(borderColor)
	tag = lipgloss.NewStyle().Padding(0, 1).MarginRight(1).Background(accentDarkColor).Foreground(baseTextColor).Bold(true)
	tagWarn = tag.Copy().Background(lipgloss.Color("#4f3a10"))
	tagErr = tag.Copy().Background(lipgloss.Color("#4f1424"))
	if !dark {
		tagWarn = tag.Copy().Background(lipgloss.Color("#f1c27d")).Foreground(lipgloss.Color("#2f1a00"))
		tagErr = tag.Copy().Background(lipgloss.Color("#f4a2bc")).Foreground(lipgloss.Color("#48000e"))
	}
	sectionTitleStyle = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	bodyTextStyle = lipgloss.NewStyle().Foreground(baseTextColor)
	statBoxStyle = lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderForeground(borderColor).MarginRight(1)
	statRunningStyle = statBoxStyle.Copy().Foreground(warningColor)
	statQueuedStyle = statBoxStyle.Copy().Foreground(subtleColor)
	statFailedStyle = statBoxStyle.Copy().Foreground(errorColor)
	statSuccessStyle = statBoxStyle.Copy().Foreground(successColor)
	paletteBox = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(borderColor).Padding(0, 1).MarginLeft(2)
	helpBox = lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(accentColor).Padding(1, 2).MarginLeft(2)
	helpKeyStyle = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	helpDescStyle = lipgloss.NewStyle().Foreground(baseTextColor)
	storeBox = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(borderColor).Padding(1, 2).MarginLeft(2)
	storeEntryStyle = lipgloss.NewStyle().Foreground(baseTextColor)
	storeEntrySelected = storeEntryStyle.Copy().Background(accentDarkColor).Bold(true)
	logInfoStyle = lipgloss.NewStyle().Foreground(baseTextColor)
	logWarnStyle = lipgloss.NewStyle().Foreground(warningColor)
	logErrorStyle = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	branchMainStyle = lipgloss.NewStyle().Foreground(successColor).Bold(true)
	branchDevStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f6e58d")).Bold(true)
	branchReleaseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb3d1")).Bold(true)
	if !dark {
		branchDevStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8b5b00")).Bold(true)
		branchReleaseStyle = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	}
	branchDefaultStyle = lipgloss.NewStyle().Foreground(baseTextColor)
	appStyle = lipgloss.NewStyle().Background(appBackground).Foreground(baseTextColor).Padding(0, 1)
}

func tableStyles() table.Styles {
	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(borderColor).
		Padding(0, 1).
		Bold(true).
		Foreground(baseTextColor)
	styles.Cell = styles.Cell.
		PaddingLeft(1).
		PaddingRight(1).
		MaxWidth(50)
	styles.Selected = styles.Selected.
		Foreground(baseTextColor).
		Background(rowHighlightBg).
		Bold(false)
	return styles
}

func (m *Model) refreshStyles() {
	m.headerStyle = headerStyle
	m.footerStyle = footerStyle
	m.bodyBoxStyle = bodyBox
	m.panelStyle = panel
	m.tagStyle = tag
	m.tagWarningStyle = tagWarn
	m.tagErrorStyle = tagErr
}

func (m *Model) toggleContrast() {
	m.contrastMode = !m.contrastMode
	setTheme(m.darkBackground, m.contrastMode)
	m.refreshStyles()
	if m.contrastMode {
		m.flashMessage = "High contrast enabled"
	} else {
		m.flashMessage = "High contrast disabled"
	}
}
