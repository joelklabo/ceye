// Default settings
const defaultSettings = {
    theme: 'dark',
    fontSize: 'medium',
    density: 'comfortable',
    autoRefresh: true,
    refreshInterval: 10,
    showNotifications: true,
    playSounds: false,
    saveFilters: true,
    debugMode: false
};

// Load settings on page load
function loadSettings() {
    const saved = localStorage.getItem('ceye-settings');
    let settings = defaultSettings;
    
    if (saved) {
        try {
            settings = { ...defaultSettings, ...JSON.parse(saved) };
        } catch (e) {
            console.error('Failed to load settings:', e);
        }
    }
    
    // Apply to UI
    document.getElementById('settingsTheme').value = settings.theme;
    document.getElementById('settingsFontSize').value = settings.fontSize;
    document.getElementById('settingsDensity').value = settings.density;
    document.getElementById('settingsAutoRefresh').checked = settings.autoRefresh;
    document.getElementById('settingsRefreshInterval').value = settings.refreshInterval;
    document.getElementById('settingsShowNotifications').checked = settings.showNotifications;
    document.getElementById('settingsPlaySounds').checked = settings.playSounds;
    document.getElementById('settingsSaveFilters').checked = settings.saveFilters;
    document.getElementById('settingsDebugMode').checked = settings.debugMode;
    
    // Apply theme
    document.documentElement.setAttribute('data-theme', settings.theme);
    
    // Apply font size
    applyFontSize(settings.fontSize);
    
    // Apply density
    applyDensity(settings.density);
    
    // Update info sections
    updateFilterInfo();
    updateStorageInfo();
}

function updateSetting(key, value) {
    const saved = localStorage.getItem('ceye-settings');
    let settings = saved ? JSON.parse(saved) : {};
    
    // Handle boolean conversion
    if (value === 'true') value = true;
    if (value === 'false') value = false;
    
    settings[key] = value;
    localStorage.setItem('ceye-settings', JSON.stringify(settings));
    
    // Apply immediately
    if (key === 'theme') {
        document.documentElement.setAttribute('data-theme', value);
        localStorage.setItem('ceye-theme', value);
    } else if (key === 'fontSize') {
        applyFontSize(value);
    } else if (key === 'density') {
        applyDensity(value);
    }
    
    console.log('Setting updated:', key, '=', value);
}

function applyFontSize(size) {
    const root = document.documentElement;
    switch (size) {
        case 'small':
            root.style.fontSize = '13px';
            break;
        case 'large':
            root.style.fontSize = '16px';
            break;
        default: // medium
            root.style.fontSize = '14px';
    }
}

function applyDensity(density) {
    document.body.className = `density-${density}`;
}

function resetAllSettings() {
    if (confirm('Reset all settings to defaults? This cannot be undone.')) {
        localStorage.setItem('ceye-settings', JSON.stringify(defaultSettings));
        localStorage.setItem('ceye-theme', defaultSettings.theme);
        location.reload();
    }
}

function exportSettings() {
    const settings = localStorage.getItem('ceye-settings') || JSON.stringify(defaultSettings);
    const filters = localStorage.getItem('ceye-filters') || '{}';
    
    const exportData = {
        settings: JSON.parse(settings),
        filters: JSON.parse(filters),
        version: '1.0',
        exportedAt: new Date().toISOString()
    };
    
    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `ceye-settings-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
}

function importSettings() {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';
    
    input.onchange = (e) => {
        const file = e.target.files[0];
        if (!file) return;
        
        const reader = new FileReader();
        reader.onload = (event) => {
            try {
                const data = JSON.parse(event.target.result);
                
                if (data.settings) {
                    localStorage.setItem('ceye-settings', JSON.stringify(data.settings));
                }
                if (data.filters) {
                    localStorage.setItem('ceye-filters', JSON.stringify(data.filters));
                }
                
                alert('Settings imported successfully!');
                location.reload();
            } catch (err) {
                alert('Failed to import settings: ' + err.message);
            }
        };
        reader.readAsText(file);
    };
    
    input.click();
}

function clearSavedFilters() {
    if (confirm('Clear all saved filters?')) {
        localStorage.removeItem('ceye-filters');
        updateFilterInfo();
        alert('Filters cleared!');
    }
}

function updateFilterInfo() {
    const filters = localStorage.getItem('ceye-filters');
    const info = document.getElementById('savedFiltersInfo');
    
    if (filters) {
        try {
            const parsed = JSON.parse(filters);
            const count = (parsed.providers?.length || 0) + (parsed.statuses?.length || 0);
            info.textContent = count > 0 ? `${count} filter(s) saved` : 'No filters saved';
        } catch (e) {
            info.textContent = 'No filters saved';
        }
    } else {
        info.textContent = 'No filters saved';
    }
}

function updateStorageInfo() {
    const info = document.getElementById('storageInfo');
    let total = 0;
    
    for (let key in localStorage) {
        if (localStorage.hasOwnProperty(key) && key.startsWith('ceye-')) {
            total += localStorage[key].length;
        }
    }
    
    const kb = (total / 1024).toFixed(2);
    info.textContent = `${kb} KB`;
}

// Load settings on page load
loadSettings();
