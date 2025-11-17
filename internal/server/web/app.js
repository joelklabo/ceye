let ws = null;
let reconnectTimeout = null;
let currentData = null;

const filters = {
    providers: [],    // Multi-select
    statuses: [],     // Multi-select
    search: ''
};

// Available filter options (populated from data)
let availableProviders = [];
let availableStatuses = ['in_progress', 'queued', 'completed', 'failed'];

// Theme management
function initTheme() {
    const savedTheme = localStorage.getItem('ceye-theme') || 'dark';
    setTheme(savedTheme);
    
    const selector = document.getElementById('themeSelector');
    if (selector) {
        selector.value = savedTheme;
        selector.addEventListener('change', (e) => {
            setTheme(e.target.value);
        });
    }
}

function setTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('ceye-theme', theme);
}

// Load saved filters
function loadSavedFilters() {
    const saved = localStorage.getItem('ceye-filters');
    if (saved) {
        try {
            const parsed = JSON.parse(saved);
            filters.providers = parsed.providers || [];
            filters.statuses = parsed.statuses || [];
            filters.search = parsed.search || '';
        } catch (e) {
            console.error('Failed to load saved filters:', e);
        }
    }
}

// Save filters
function saveFilters() {
    localStorage.setItem('ceye-filters', JSON.stringify(filters));
}

// Workspaces (named filter presets)
function loadWorkspaces() {
    const saved = localStorage.getItem('ceye-workspaces');
    return saved ? JSON.parse(saved) : [];
}

function saveWorkspaces(workspaces) {
    localStorage.setItem('ceye-workspaces', JSON.stringify(workspaces));
}

function saveCurrentAsWorkspace() {
    const name = prompt('Enter workspace name:');
    if (!name) return;
    
    const workspaces = loadWorkspaces();
    const workspace = {
        name: name,
        filters: { ...filters },
        createdAt: new Date().toISOString()
    };
    
    // Replace if exists
    const index = workspaces.findIndex(w => w.name === name);
    if (index >= 0) {
        if (!confirm(`Workspace "${name}" already exists. Overwrite?`)) return;
        workspaces[index] = workspace;
    } else {
        workspaces.push(workspace);
    }
    
    saveWorkspaces(workspaces);
    updateWorkspaceSelector();
    alert(`Workspace "${name}" saved!`);
}

function loadWorkspace(name) {
    const workspaces = loadWorkspaces();
    const workspace = workspaces.find(w => w.name === name);
    
    if (workspace) {
        filters.providers = workspace.filters.providers || [];
        filters.statuses = workspace.filters.statuses || [];
        filters.search = workspace.filters.search || '';
        
        document.getElementById('searchBox').value = filters.search;
        saveFilters();
        
        if (currentData) render(currentData);
    }
}

function deleteWorkspace(name) {
    if (!confirm(`Delete workspace "${name}"?`)) return;
    
    const workspaces = loadWorkspaces();
    const filtered = workspaces.filter(w => w.name !== name);
    saveWorkspaces(filtered);
    updateWorkspaceSelector();
    alert(`Workspace "${name}" deleted!`);
}

function updateWorkspaceSelector() {
    const selector = document.getElementById('workspaceSelector');
    if (!selector) return;
    
    const workspaces = loadWorkspaces();
    selector.innerHTML = '<option value="">Select Workspace...</option>';
    
    workspaces.forEach(ws => {
        const option = document.createElement('option');
        option.value = ws.name;
        option.textContent = ws.name;
        selector.appendChild(option);
    });
}

function connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
        console.log('WebSocket connected');
        updateConnectionStatus(true);
        addActivityItem('websocket', 'Connected to server');
    };
    
    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        currentData = data;
        
        // Log activity
        const runCount = data.runs ? data.runs.length : 0;
        const activeCount = data.totals ? (data.totals.running + data.totals.queued) : 0;
        addActivityItem('success', `Received update: ${runCount} runs (${activeCount} active)`);
        
        // Update last update timestamp
        updateLastUpdate();
        
        render(data);
    };
    
    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        updateConnectionStatus(false);
    };
    
    ws.onclose = () => {
        console.log('WebSocket closed, reconnecting...');
        updateConnectionStatus(false);
        addActivityItem('error', 'Disconnected from server, reconnecting...');
        reconnectTimeout = setTimeout(connect, 3000);
    };
}

function updateConnectionStatus(connected) {
    const status = document.getElementById('connectionStatus');
    if (connected) {
        status.textContent = '● Connected';
        status.className = 'connection-indicator connected';
    } else {
        status.textContent = '○ Disconnected';
        status.className = 'connection-indicator disconnected';
    }
}

function updateLastUpdate() {
    const now = new Date();
    const timeStr = now.toLocaleTimeString();
    const lastUpdate = document.getElementById('lastUpdate');
    if (lastUpdate) {
        lastUpdate.textContent = `Last update: ${timeStr}`;
    }
}

// Activity Log
let activityLogOpen = false;
const maxActivityItems = 50;

function toggleActivityLog() {
    const log = document.getElementById('activityLog');
    const toggle = document.getElementById('activityToggle');
    activityLogOpen = !activityLogOpen;
    log.style.display = activityLogOpen ? 'block' : 'none';
    toggle.textContent = activityLogOpen ? '▲' : '▼';
}

function addActivityItem(type, message, details) {
    const log = document.getElementById('activityLog');
    if (!log) return;
    
    const now = new Date();
    const timestamp = now.toLocaleTimeString();
    
    const item = document.createElement('div');
    item.className = `activity-item ${type} activity-pulse`;
    
    const content = `<span class="timestamp">${timestamp}</span><span class="message">${message}</span>`;
    item.innerHTML = details ? `${content}<div class="details">${details}</div>` : content;
    
    // Remove first item if over limit
    if (log.children.length >= maxActivityItems) {
        log.removeChild(log.firstChild);
    }
    
    // Clear "waiting" message on first real item
    if (log.children.length === 1 && log.firstChild && log.firstChild.classList && log.firstChild.classList.contains('muted')) {
        log.innerHTML = '';
    }
    
    log.appendChild(item);
    
    // Auto-scroll to bottom if near bottom
    if (log.scrollHeight - log.scrollTop <= log.clientHeight + 50) {
        log.scrollTop = log.scrollHeight;
    }
    
    // Remove pulse after animation
    setTimeout(() => item.classList.remove('activity-pulse'), 500);
}

function render(data) {
    updateStats(data.totals || {});
    updateProviderFilter(data.providers || []);
    updateRunsTable(data.runs || []);
    updateProviderHealth(data.status || {}, data.health || {});
    updateAlertBadge(data.alert_count || 0);
    updateLastUpdate(data.timestamp);
}

function updateStats(totals) {
    document.getElementById('runningCount').textContent = totals.running || 0;
    document.getElementById('queuedCount').textContent = totals.queued || 0;
    document.getElementById('successCount').textContent = totals.success || 0;
    document.getElementById('failedCount').textContent = totals.failed || 0;
}

function updateProviderFilter(providers) {
    availableProviders = providers;
    renderFilterPills();
}

function renderFilterPills() {
    const container = document.getElementById('filterPills');
    if (!container) return;
    
    const pills = [];
    
    // Provider pills
    filters.providers.forEach(provider => {
        pills.push(`
            <span class="filter-pill">
                ${escapeHtml(provider)}
                <button onclick="removeProviderFilter('${escapeHtml(provider)}')" class="pill-remove">&times;</button>
            </span>
        `);
    });
    
    // Status pills
    filters.statuses.forEach(status => {
        pills.push(`
            <span class="filter-pill status-${status}">
                ${escapeHtml(status)}
                <button onclick="removeStatusFilter('${escapeHtml(status)}')" class="pill-remove">&times;</button>
            </span>
        `);
    });
    
    // Clear all button
    if (pills.length > 0 || filters.search) {
        pills.push(`<button onclick="clearAllFilters()" class="btn-clear-filters">Clear All</button>`);
    }
    
    container.innerHTML = pills.join('');
}

function removeProviderFilter(provider) {
    filters.providers = filters.providers.filter(p => p !== provider);
    saveFilters();
    if (currentData) render(currentData);
}

function removeStatusFilter(status) {
    filters.statuses = filters.statuses.filter(s => s !== status);
    saveFilters();
    if (currentData) render(currentData);
}

function clearAllFilters() {
    filters.providers = [];
    filters.statuses = [];
    filters.search = '';
    document.getElementById('searchBox').value = '';
    saveFilters();
    if (currentData) render(currentData);
}

function updateRunsTable(runs) {
    const container = document.getElementById('runsTable');
    
    const filtered = runs.filter(run => {
        // Provider filter
        if (filters.providers.length > 0 && !filters.providers.includes(run.Provider)) {
            return false;
        }
        
        // Status filter
        if (filters.statuses.length > 0 && !filters.statuses.includes(run.Status)) {
            return false;
        }
        
        // Search filter
        if (filters.search) {
            const search = filters.search.toLowerCase();
            return run.WorkflowName.toLowerCase().includes(search) ||
                   run.Repo.toLowerCase().includes(search) ||
                   run.Branch.toLowerCase().includes(search);
        }
        
        return true;
    });
    
    if (filtered.length === 0) {
        container.innerHTML = '<div class="empty-state">No runs found</div>';
        return;
    }
    
    const table = document.createElement('table');
    table.innerHTML = `
        <thead>
            <tr>
                <th>Provider</th>
                <th>Repository</th>
                <th>Workflow</th>
                <th>Branch</th>
                <th>Status</th>
                <th>Duration</th>
                <th>Updated</th>
            </tr>
        </thead>
        <tbody>
            ${filtered.map(run => `
                <tr>
                    <td><span class="provider-badge">${escapeHtml(run.Provider)}</span></td>
                    <td>${escapeHtml(run.Repo)}</td>
                    <td>
                        ${run.URL ? `<a href="${escapeHtml(run.URL)}" target="_blank" class="run-link">${escapeHtml(run.WorkflowName)}</a>` : escapeHtml(run.WorkflowName)}
                    </td>
                    <td>${escapeHtml(run.Branch)}</td>
                    <td><span class="status-badge status-${run.Status}">${formatStatus(run.Status)}</span></td>
                    <td>${formatDuration(run.Duration)}</td>
                    <td>${formatTimestamp(run.UpdatedAt)}</td>
                </tr>
            `).join('')}
        </tbody>
    `;
    
    container.innerHTML = '';
    container.appendChild(table);
}

function updateProviderHealth(status, health) {
    const container = document.getElementById('providerHealth');
    
    const providers = Object.keys(status);
    if (providers.length === 0) {
        container.innerHTML = '<div class="empty-state">No providers</div>';
        return;
    }
    
    container.innerHTML = providers.map(provider => {
        const error = status[provider];
        const h = health[provider] || {};
        const isHealthy = !error;
        
        return `
            <div class="health-item ${isHealthy ? 'healthy' : 'error'}">
                <div class="health-name">${escapeHtml(provider)}</div>
                <div class="health-status">
                    ${error ? `❌ ${escapeHtml(error)}` : '✅ Healthy'}
                </div>
                ${h.LastSuccess ? `<div class="health-status">Last success: ${formatTimestamp(h.LastSuccess)}</div>` : ''}
                ${h.ErrorCount > 0 ? `<div class="health-status">Errors: ${h.ErrorCount}</div>` : ''}
            </div>
        `;
    }).join('');
}

function updateLastUpdate(timestamp) {
    if (timestamp) {
        const date = new Date(timestamp);
        document.getElementById('lastUpdate').textContent = `Last update: ${date.toLocaleTimeString()}`;
    }
}

function formatStatus(status) {
    const map = {
        'in_progress': 'Running',
        'queued': 'Queued',
        'completed': 'Completed',
        'failed': 'Failed',
        'cancelled': 'Cancelled'
    };
    return map[status] || status;
}

function formatDuration(duration) {
    if (!duration) return '-';
    const seconds = Math.floor(duration / 1000000000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) {
        return `${hours}h ${minutes % 60}m`;
    }
    if (minutes > 0) {
        return `${minutes}m ${seconds % 60}s`;
    }
    return `${seconds}s`;
}

function formatTimestamp(timestamp) {
    if (!timestamp) return '-';
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now - date;
    
    if (diff < 60000) return 'just now';
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    return date.toLocaleString();
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function updateAlertBadge(count) {
    const badge = document.getElementById('alertBadge');
    if (count > 0) {
        badge.textContent = count;
        badge.style.display = 'inline-block';
    } else {
        badge.style.display = 'none';
    }
}

// Keyboard shortcuts
let selectedRowIndex = -1;

const shortcuts = {
    'r': refresh,
    '/': focusSearch,
    'Escape': clearSearch,
    'a': goToAlerts,
    '?': showKeyboardHelp,
};

function refresh() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        addActivityItem('websocket', 'Requesting refresh from server');
        ws.send('refresh');
    }
}

function focusSearch() {
    const searchBox = document.getElementById('searchBox');
    searchBox.focus();
    searchBox.select();
}

function clearSearch() {
    const searchBox = document.getElementById('searchBox');
    if (document.activeElement === searchBox) {
        searchBox.blur();
    } else if (filters.search || filters.providers.length > 0 || filters.statuses.length > 0) {
        clearAllFilters();
    }
}

function goToAlerts() {
    window.location.href = '/alerts.html';
}

function showKeyboardHelp() {
    const modal = document.getElementById('helpModal');
    if (modal) {
        modal.classList.add('show');
    }
}

function closeHelpModal() {
    const modal = document.getElementById('helpModal');
    if (modal) {
        modal.classList.remove('show');
    }
}

// Filter management
function addProviderFilter() {
    const provider = prompt('Enter provider name (e.g., github, azure):');
    if (provider && !filters.providers.includes(provider)) {
        filters.providers.push(provider);
        saveFilters();
        addActivityItem('success', `Added provider filter: ${provider}`);
        if (currentData) render(currentData);
    }
}

function addStatusFilter() {
    const status = prompt('Enter status (in_progress, queued, completed, failed):');
    if (status && availableStatuses.includes(status) && !filters.statuses.includes(status)) {
        filters.statuses.push(status);
        saveFilters();
        addActivityItem('success', `Added status filter: ${status}`);
        if (currentData) render(currentData);
    }
}

// Keyboard event handler
document.addEventListener('keydown', (e) => {
    // Don't intercept if typing in input/textarea
    if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') {
        return;
    }
    
    const handler = shortcuts[e.key];
    if (handler) {
        e.preventDefault();
        handler();
    }
});

// Event listeners
document.getElementById('refreshBtn').addEventListener('click', () => {
    refresh();
});

document.getElementById('searchBox').addEventListener('input', (e) => {
    filters.search = e.target.value;
    saveFilters();
    if (currentData) render(currentData);
});

// Initialize theme
initTheme();

// Load saved filters
loadSavedFilters();

// Initialize workspace selector
updateWorkspaceSelector();

// Initialize
connect();
