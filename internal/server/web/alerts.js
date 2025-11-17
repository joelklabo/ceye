let ws = null;
let reconnectTimeout = null;
let alerts = [];
let lastAlertCount = 0;

const filters = {
    severity: '',
    rule: '',
    search: ''
};

function connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
        console.log('WebSocket connected');
        updateConnectionStatus(true);
        fetchAlerts();
    };
    
    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        
        // Check for new alerts
        if (data.alert_count && data.alert_count > lastAlertCount) {
            showToast('New alert fired!', 'critical');
            fetchAlerts();
        }
        lastAlertCount = data.alert_count || 0;
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

async function fetchAlerts() {
    try {
        const params = new URLSearchParams();
        if (filters.severity) params.append('severity', filters.severity);
        if (filters.rule) params.append('rule', filters.rule);
        
        const url = `/api/alerts/history?${params.toString()}`;
        const response = await fetch(url);
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        
        const data = await response.json();
        alerts = data.alerts || [];
        render();
    } catch (error) {
        console.error('Failed to fetch alerts:', error);
        showToast('Failed to fetch alerts', 'critical');
    }
}

function render() {
    updateStats();
    updateRuleFilter();
    updateAlertsTable();
    updateLastUpdate();
}

function updateStats() {
    const total = alerts.length;
    const critical = alerts.filter(a => a.severity === 'critical').length;
    const warning = alerts.filter(a => a.severity === 'warning').length;
    
    // Count alerts in last 5 minutes
    const fiveMinAgo = new Date(Date.now() - 5 * 60 * 1000);
    const recent = alerts.filter(a => new Date(a.triggered_at) > fiveMinAgo).length;
    
    document.getElementById('totalCount').textContent = total;
    document.getElementById('criticalCount').textContent = critical;
    document.getElementById('warningCount').textContent = warning;
    document.getElementById('recentCount').textContent = recent;
}

function updateRuleFilter() {
    const select = document.getElementById('ruleFilter');
    const currentValue = select.value;
    
    // Get unique rule names
    const rules = [...new Set(alerts.map(a => a.rule_name))].sort();
    
    select.innerHTML = '<option value="">All Rules</option>';
    rules.forEach(rule => {
        const option = document.createElement('option');
        option.value = rule;
        option.textContent = rule;
        select.appendChild(option);
    });
    
    if (currentValue) {
        select.value = currentValue;
    }
}

function updateAlertsTable() {
    const container = document.getElementById('alertsTable');
    
    const filtered = alerts.filter(alert => {
        if (filters.severity && alert.severity !== filters.severity) return false;
        if (filters.rule && alert.rule_name !== filters.rule) return false;
        if (filters.search) {
            const search = filters.search.toLowerCase();
            return alert.message.toLowerCase().includes(search) ||
                   alert.rule_name.toLowerCase().includes(search) ||
                   (alert.run && alert.run.Repo && alert.run.Repo.toLowerCase().includes(search));
        }
        return true;
    });
    
    if (filtered.length === 0) {
        container.innerHTML = '<div class="empty-state">No alerts found</div>';
        return;
    }
    
    const table = document.createElement('table');
    table.innerHTML = `
        <thead>
            <tr>
                <th>Time</th>
                <th>Severity</th>
                <th>Rule</th>
                <th>Message</th>
                <th>Repo</th>
                <th>Workflow</th>
            </tr>
        </thead>
        <tbody>
            ${filtered.map(alert => `
                <tr>
                    <td>${formatTime(alert.triggered_at)}</td>
                    <td><span class="severity-badge severity-${alert.severity}">${alert.severity}</span></td>
                    <td><span class="rule-badge">${alert.rule_name}</span></td>
                    <td>${escapeHtml(alert.message)}</td>
                    <td>${alert.run ? escapeHtml(alert.run.Repo) : '-'}</td>
                    <td>${alert.run ? escapeHtml(alert.run.WorkflowName) : '-'}</td>
                </tr>
            `).join('')}
        </tbody>
    `;
    
    container.innerHTML = '';
    container.appendChild(table);
}

function formatTime(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    
    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;
    
    const diffDays = Math.floor(diffHours / 24);
    if (diffDays < 7) return `${diffDays}d ago`;
    
    return date.toLocaleDateString();
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function updateLastUpdate() {
    const elem = document.getElementById('lastUpdate');
    elem.textContent = `Updated: ${new Date().toLocaleTimeString()}`;
}

function showToast(message, severity = 'info') {
    const container = document.getElementById('toastContainer');
    const toast = document.createElement('div');
    toast.className = `toast toast-${severity}`;
    toast.textContent = message;
    
    container.appendChild(toast);
    
    // Animate in
    setTimeout(() => toast.classList.add('show'), 10);
    
    // Remove after 4 seconds
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 300);
    }, 4000);
}

// Event listeners
document.getElementById('refreshBtn').addEventListener('click', fetchAlerts);

document.getElementById('severityFilter').addEventListener('change', (e) => {
    filters.severity = e.target.value;
    render();
});

document.getElementById('ruleFilter').addEventListener('change', (e) => {
    filters.rule = e.target.value;
    render();
});

document.getElementById('searchBox').addEventListener('input', (e) => {
    filters.search = e.target.value;
    render();
});

// Initialize
connect();
