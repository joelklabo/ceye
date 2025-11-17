let ws = null;
let reconnectTimeout = null;
let alerts = [];
let ruleStats = [];
let lastAlertCount = 0;

const filters = {
    severity: '',
    rule: '',
    search: ''
};

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

function connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
        console.log('WebSocket connected');
        updateConnectionStatus(true);
        fetchAlerts();
        fetchRuleStats();
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
    updateRulesStatus();
    updateAlertsTable();
    updateLastUpdate();
}

async function fetchRuleStats() {
    try {
        const response = await fetch('/api/alerts/rules/stats');
        if (!response.ok) {
            console.warn('Rule stats not available');
            return;
        }
        const data = await response.json();
        ruleStats = data.rules || [];
        updateRulesStatus();
    } catch (error) {
        console.error('Failed to fetch rule stats:', error);
    }
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

function updateRulesStatus() {
    const container = document.getElementById('rulesStatus');
    
    if (ruleStats.length === 0) {
        container.innerHTML = '<div class="empty-state">No alert rules configured</div>';
        return;
    }
    
    container.innerHTML = ruleStats.map(rule => {
        const statusClass = rule.enabled ? 'rule-enabled' : 'rule-disabled';
        const cooldownText = rule.cooldown_remaining_seconds > 0 
            ? `(cooldown: ${rule.cooldown_remaining_seconds}s)` 
            : '';
        
        return `
            <div class="rule-card ${statusClass}">
                <div class="rule-header">
                    <span class="rule-name">${escapeHtml(rule.rule_name)}</span>
                    <span class="rule-status-badge ${rule.enabled ? 'enabled' : 'disabled'}">
                        ${rule.enabled ? '✓ Active' : '○ Disabled'}
                    </span>
                </div>
                <div class="rule-stats">
                    <div class="stat-item">
                        <span class="stat-label">Total Fires:</span>
                        <span class="stat-value">${rule.total_alerts}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Last Hour:</span>
                        <span class="stat-value">${rule.fires_last_hour}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Last Fired:</span>
                        <span class="stat-value">${rule.last_fired ? formatTime(rule.last_fired) : 'Never'}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Evaluations:</span>
                        <span class="stat-value">${rule.total_evaluations.toLocaleString()}</span>
                    </div>
                </div>
                ${cooldownText ? `<div class="rule-cooldown">${cooldownText}</div>` : ''}
            </div>
        `;
    }).join('');
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
                <th>Actions</th>
            </tr>
        </thead>
        <tbody>
            ${filtered.map((alert, index) => `
                <tr>
                    <td>${formatTime(alert.triggered_at)}</td>
                    <td><span class="severity-badge severity-${alert.severity}">${alert.severity}</span></td>
                    <td><span class="rule-badge">${alert.rule_name}</span></td>
                    <td>${escapeHtml(alert.message)}</td>
                    <td>${alert.run ? escapeHtml(alert.run.Repo) : '-'}</td>
                    <td>${alert.run ? escapeHtml(alert.run.WorkflowName) : '-'}</td>
                    <td>
                        <button class="btn-details" onclick="showAlertDetails(${index})">Details</button>
                    </td>
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

function showAlertDetails(index) {
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
    
    const alert = filtered[index];
    if (!alert) return;
    
    const modal = document.getElementById('alertModal');
    const details = document.getElementById('alertDetails');
    
    const run = alert.run || {};
    const fullTime = new Date(alert.triggered_at).toLocaleString();
    
    details.innerHTML = `
        <div class="detail-section">
            <h3>Alert Information</h3>
            <div class="detail-row">
                <span class="detail-label">Rule:</span>
                <span class="rule-badge">${escapeHtml(alert.rule_name)}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Severity:</span>
                <span class="severity-badge severity-${alert.severity}">${alert.severity}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Condition:</span>
                <span>${escapeHtml(alert.condition)}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Triggered:</span>
                <span>${fullTime}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Message:</span>
                <span>${escapeHtml(alert.message)}</span>
            </div>
        </div>
        
        ${run.id ? `
            <div class="detail-section">
                <h3>Run Information</h3>
                <div class="detail-row">
                    <span class="detail-label">Repository:</span>
                    <span>${escapeHtml(run.Repo)}</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Workflow:</span>
                    <span>${escapeHtml(run.WorkflowName)}</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Branch:</span>
                    <span>${escapeHtml(run.branch)}</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Status:</span>
                    <span class="status-badge status-${run.status}">${run.status}</span>
                </div>
                ${run.conclusion ? `
                    <div class="detail-row">
                        <span class="detail-label">Conclusion:</span>
                        <span>${escapeHtml(run.conclusion)}</span>
                    </div>
                ` : ''}
                <div class="detail-row">
                    <span class="detail-label">Run ID:</span>
                    <span class="monospace">${escapeHtml(run.id)}</span>
                </div>
                ${run.url ? `
                    <div class="detail-row">
                        <span class="detail-label">URL:</span>
                        <a href="${escapeHtml(run.url)}" target="_blank" class="run-link">View Run →</a>
                    </div>
                ` : ''}
            </div>
        ` : '<p class="empty-state">No run information available</p>'}
    `;
    
    modal.classList.add('show');
}

function closeAlertModal() {
    const modal = document.getElementById('alertModal');
    modal.classList.remove('show');
}

// Keyboard shortcuts
const shortcuts = {
    'r': () => fetchAlerts(),
    '/': focusSearch,
    'Escape': handleEscape,
    'd': () => window.location.href = '/',
    '?': showKeyboardHelp,
};

function focusSearch() {
    const searchBox = document.getElementById('searchBox');
    searchBox.focus();
    searchBox.select();
}

function handleEscape() {
    // Close modal if open
    const alertModal = document.getElementById('alertModal');
    if (alertModal && alertModal.classList.contains('show')) {
        closeAlertModal();
        return;
    }
    
    const helpModal = document.getElementById('helpModal');
    if (helpModal && helpModal.classList.contains('show')) {
        closeHelpModal();
        return;
    }
    
    // Otherwise clear search/filters
    const searchBox = document.getElementById('searchBox');
    if (document.activeElement === searchBox) {
        searchBox.blur();
    } else if (filters.search || filters.severity || filters.rule) {
        filters.search = '';
        filters.severity = '';
        filters.rule = '';
        document.getElementById('searchBox').value = '';
        document.getElementById('severityFilter').value = '';
        document.getElementById('ruleFilter').value = '';
        render();
    }
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
        // But allow Escape to blur
        if (e.key === 'Escape') {
            e.target.blur();
        }
        return;
    }
    
    const handler = shortcuts[e.key];
    if (handler) {
        e.preventDefault();
        handler();
    }
});

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

// Initialize theme
initTheme();

// Initialize
connect();
