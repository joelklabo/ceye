let ws = null;
let reconnectTimeout = null;
let currentData = null;

const filters = {
    provider: '',
    status: '',
    search: ''
};

function connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
        console.log('WebSocket connected');
        updateConnectionStatus(true);
    };
    
    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        currentData = data;
        render(data);
    };
    
    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        updateConnectionStatus(false);
    };
    
    ws.onclose = () => {
        console.log('WebSocket closed, reconnecting...');
        updateConnectionStatus(false);
        reconnectTimeout = setTimeout(connect, 3000);
    };
}

function updateConnectionStatus(connected) {
    const status = document.getElementById('connectionStatus');
    status.textContent = connected ? '● Connected' : '○ Disconnected';
    status.className = connected ? 'connected' : 'disconnected';
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
    const select = document.getElementById('providerFilter');
    const currentValue = select.value;
    
    select.innerHTML = '<option value="">All Providers</option>';
    providers.forEach(provider => {
        const option = document.createElement('option');
        option.value = provider;
        option.textContent = provider;
        select.appendChild(option);
    });
    
    if (currentValue) {
        select.value = currentValue;
    }
}

function updateRunsTable(runs) {
    const container = document.getElementById('runsTable');
    
    const filtered = runs.filter(run => {
        if (filters.provider && run.Provider !== filters.provider) return false;
        if (filters.status && run.Status !== filters.status) return false;
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
    } else if (filters.search || filters.provider || filters.status) {
        filters.search = '';
        filters.provider = '';
        filters.status = '';
        document.getElementById('searchBox').value = '';
        document.getElementById('providerFilter').value = '';
        document.getElementById('statusFilter').value = '';
        if (currentData) render(currentData);
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
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send('refresh');
    }
});

document.getElementById('providerFilter').addEventListener('change', (e) => {
    filters.provider = e.target.value;
    if (currentData) render(currentData);
});

document.getElementById('statusFilter').addEventListener('change', (e) => {
    filters.status = e.target.value;
    if (currentData) render(currentData);
});

document.getElementById('searchBox').addEventListener('input', (e) => {
    filters.search = e.target.value;
    if (currentData) render(currentData);
});

// Initialize
connect();
