// Kader-Planung JavaScript functionality

// Global status refresh interval (every 30 seconds when running)
let statusRefreshInterval = null;

// API calls for Kader-Planung
async function getKaderPlanungStatus() {
    return api.request('/api/v1/kader-planung/status');
}

async function getKaderPlanungFiles() {
    return api.request('/api/v1/kader-planung/files');
}

async function getAnalysisCapabilities() {
    return api.request('/api/v1/kader-planung/capabilities');
}

async function startDetailedAnalysis(params) {
    return api.request('/api/v1/kader-planung/start', {
        method: 'POST',
        body: JSON.stringify(params)
    });
}

async function startStatisticalAnalysis(params) {
    return api.request('/api/v1/kader-planung/statistical', {
        method: 'POST',
        body: JSON.stringify(params)
    });
}

async function startHybridAnalysis(params) {
    return api.request('/api/v1/kader-planung/hybrid', {
        method: 'POST',
        body: JSON.stringify(params)
    });
}

async function getStatisticalResults() {
    return api.request('/api/v1/kader-planung/statistical/files');
}

// Status display functions
async function refreshStatus() {
    try {
        Utils.showLoading('status-result');
        const response = await getKaderPlanungStatus();
        
        // Debug logging
        console.log('Status API response:', response);
        
        // Handle the API wrapper format {success: true, data: {...}}
        if (!response.success || !response.data) {
            console.error('Status API error response:', response);
            throw new Error(response.error || response.message || 'API call failed');
        }
        
        // The actual status data is in response.data
        const statusResponse = {
            status: response.data.status,
            available_files: response.data.available_files || []
        };
        
        displayStatus(statusResponse);
        
        // Setup auto-refresh if running
        setupStatusAutoRefresh(response.data.status);
    } catch (error) {
        console.error('Status refresh failed:', error);
        Utils.showError('status-result', `Status abrufen fehlgeschlagen: ${error.message}`);
    }
}

function displayStatus(response) {
    const statusContainer = document.getElementById('status-result');
    const status = response.status; // This is the ExecutionStatus object
    
    // Determine status indicator and text
    let statusIndicator, statusText, statusClass;
    
    if (status.running || status.Running) { // Handle both possible field names
        statusIndicator = '🔄';
        statusText = 'Läuft gerade...';
        statusClass = 'status-running';
    } else {
        statusIndicator = '✅';
        statusText = 'Abgeschlossen';
        statusClass = 'status-completed';
    }
    
    // Format dates - use actual field names from ExecutionStatus with fallbacks
    const lastExecution = (status.last_execution || status.LastExecution) ? 
        formatDateTime(status.last_execution || status.LastExecution) : 'Noch nie ausgeführt';
    const nextScheduled = 'Täglich nach Datenbankimport'; // No next_scheduled field in backend
    const startTime = (status.start_time || status.StartTime) ? 
        formatDateTime(status.start_time || status.StartTime) : null;
    const lastSuccess = (status.last_success || status.LastSuccess) ? 
        formatDateTime(status.last_success || status.LastSuccess) : null;
    const lastError = status.last_error || status.LastError;
    
    statusContainer.innerHTML = `
        <div class="status-info">
            <div class="status-row ${statusClass}">
                <span class="status-indicator">${statusIndicator}</span>
                <strong>Status:</strong> ${statusText}
            </div>
            <div class="status-row">
                <span class="status-indicator">📅</span>
                <strong>Letzte Ausführung:</strong> ${lastExecution}
            </div>
            <div class="status-row">
                <span class="status-indicator">⏰</span>
                <strong>Nächste Ausführung:</strong> ${nextScheduled}
            </div>
            ${(status.running || status.Running) && startTime ? `
                <div class="status-row">
                    <span class="status-indicator">⏱️</span>
                    <strong>Gestartet:</strong> ${startTime}
                </div>
            ` : ''}
            ${lastSuccess ? `
                <div class="status-row">
                    <span class="status-indicator">✅</span>
                    <strong>Letzter Erfolg:</strong> ${lastSuccess}
                </div>
            ` : ''}
            ${lastError ? `
                <div class="status-row">
                    <span class="status-indicator">⚠️</span>
                    <strong>Letzter Fehler:</strong> ${lastError}
                </div>
            ` : ''}
            <div class="status-row">
                <span class="status-indicator">📊</span>
                <strong>Verfügbare Dateien:</strong> ${response.available_files ? response.available_files.length : 0}
            </div>
        </div>
    `;
}

function setupStatusAutoRefresh(status) {
    // Clear existing interval
    if (statusRefreshInterval) {
        clearInterval(statusRefreshInterval);
        statusRefreshInterval = null;
    }
    
    // Setup auto-refresh if running - handle both possible field names
    if (status.running || status.Running) {
        statusRefreshInterval = setInterval(() => {
            refreshStatus();
        }, 30000); // Refresh every 30 seconds
    }
}

// Files display functions
async function refreshFiles() {
    try {
        Utils.showLoading('files-result');
        const response = await getKaderPlanungFiles();
        
        // Debug logging
        console.log('Files API response:', response);
        
        // Handle the API wrapper format {success: true, data: [...]}
        if (!response.success || !Array.isArray(response.data)) {
            console.error('Files API error response:', response);
            throw new Error(response.error || response.message || 'API call failed or invalid data format');
        }
        
        // The actual files array is in response.data
        displayFiles(response.data);
    } catch (error) {
        console.error('Files refresh failed:', error);
        Utils.showError('files-result', `Dateiliste abrufen fehlgeschlagen: ${error.message}`);
    }
}

function displayFiles(files) {
    const filesContainer = document.getElementById('files-result');
    
    if (!files || files.length === 0) {
        filesContainer.innerHTML = `
            <div class="info-message">
                <p>🗂️ Noch keine Export-Dateien verfügbar.</p>
                <p>Die erste Datei wird nach dem nächsten automatischen Import erstellt.</p>
            </div>
        `;
        return;
    }
    
    // Sort files by modification time (newest first) - handle both possible field names
    files.sort((a, b) => {
        const dateA = new Date(a.mod_time || a.ModTime || a.modified || 0);
        const dateB = new Date(b.mod_time || b.ModTime || b.modified || 0);
        return dateB - dateA;
    });
    
    // Find the latest file (newest one)
    const latestFile = files[0];
    
    let tableHTML = `
        <div class="table-responsive">
            <table class="data-table">
                <thead>
                    <tr>
                        <th>📁 Dateiname</th>
                        <th>📅 Erstellt am</th>
                        <th>📊 Dateigröße</th>
                        <th>🏛️ Vereine</th>
                        <th>👥 Spieler</th>
                        <th>⬇️ Download</th>
                    </tr>
                </thead>
                <tbody>
    `;
    
    files.forEach(file => {
        const isLatest = file.name === latestFile.name;
        const rowClass = isLatest ? 'latest-file' : '';
        
        // Handle multiple possible field names for date and size
        const modTime = file.mod_time || file.ModTime || file.modified;
        const fileSize = file.size || file.Size;
        const fileName = file.name || file.Name;
        
        tableHTML += `
            <tr class="${rowClass}">
                <td>
                    ${isLatest ? '<span class="latest-badge">📌 AKTUELL</span>' : ''}
                    <span class="filename">${fileName}</span>
                </td>
                <td>${modTime ? formatDateTime(modTime) : 'N/A'}</td>
                <td>${fileSize ? formatFileSize(fileSize) : 'N/A'}</td>
                <td>N/A</td>
                <td>N/A</td>
                <td>
                    <button onclick="downloadFile('${fileName}')" class="btn btn-primary btn-sm">
                        ⬇️ Herunterladen
                    </button>
                </td>
            </tr>
        `;
    });
    
    tableHTML += `
                </tbody>
            </table>
        </div>
        <div class="files-summary">
            <p><strong>Gesamt:</strong> ${files.length} verfügbare Dateien | <strong>Neueste:</strong> ${latestFile.name || latestFile.Name}</p>
        </div>
    `;
    
    filesContainer.innerHTML = tableHTML;
}

// File download function
function downloadFile(filename) {
    const downloadUrl = `${api.baseURL}/api/v1/kader-planung/download/${encodeURIComponent(filename)}`;
    
    // Create temporary link element for download
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = filename;
    link.style.display = 'none';
    
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    
    Utils.showSuccess('files-result', `Download für "${filename}" wurde gestartet.`, 3000);
}

// Utility functions
function formatDateTime(dateString) {
    if (!dateString) return 'N/A';
    
    const date = new Date(dateString);
    return date.toLocaleString('de-DE', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

function formatFileSize(bytes) {
    if (!bytes) return 'N/A';
    
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    if (bytes === 0) return '0 Bytes';
    
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
}

// UI functionality for new analysis modes
function initializeModeSelector() {
    const modeSelector = document.getElementById('processing-mode');
    if (modeSelector) {
        modeSelector.addEventListener('change', function() {
            toggleModeOptions(this.value);
        });

        // Initialize with default mode
        toggleModeOptions('detailed');
    }
}

function toggleModeOptions(mode) {
    const statisticalOptions = document.getElementById('statistical-options');
    const detailedOptions = document.getElementById('detailed-options');

    if (!statisticalOptions || !detailedOptions) return;

    // Show/hide options based on selected mode
    switch (mode) {
        case 'statistical':
            statisticalOptions.style.display = 'block';
            detailedOptions.style.display = 'none';
            break;
        case 'hybrid':
            statisticalOptions.style.display = 'block';
            detailedOptions.style.display = 'block';
            break;
        case 'detailed':
        case 'efficient':
        default:
            statisticalOptions.style.display = 'none';
            detailedOptions.style.display = 'block';
            break;
    }
}

async function loadAnalysisCapabilities() {
    try {
        const response = await getAnalysisCapabilities();
        if (response.success && response.data) {
            console.log('Analysis capabilities:', response.data);
            // Could update UI based on capabilities
        }
    } catch (error) {
        console.warn('Failed to load analysis capabilities:', error);
    }
}

async function startExecution() {
    const startBtn = document.getElementById('start-btn');
    if (!startBtn) return;

    // Disable button during execution
    startBtn.disabled = true;
    startBtn.textContent = '⏳ Starte...';

    try {
        const mode = document.getElementById('processing-mode')?.value || 'detailed';
        const params = collectExecutionParams(mode);

        let response;
        switch (mode) {
            case 'statistical':
                response = await startStatisticalAnalysis(params);
                break;
            case 'hybrid':
                response = await startHybridAnalysis(params);
                break;
            case 'detailed':
            case 'efficient':
            default:
                response = await startDetailedAnalysis(params);
                break;
        }

        if (response.success) {
            Utils.showSuccess('status-result', `${getModeDisplayName(mode)} wurde erfolgreich gestartet!`, 5000);
            // Refresh status to show running state
            setTimeout(() => refreshStatus(), 1000);
        } else {
            throw new Error(response.error || response.message || 'Unbekannter Fehler');
        }
    } catch (error) {
        console.error('Execution failed:', error);
        Utils.showError('status-result', `Start fehlgeschlagen: ${error.message}`);
    } finally {
        // Re-enable button
        startBtn.disabled = false;
        startBtn.textContent = '🚀 Analyse starten';
    }
}

function collectExecutionParams(mode) {
    const params = {};

    // Basic parameters for all modes
    const clubPrefix = document.getElementById('club-prefix')?.value;
    if (clubPrefix) params.club_prefix = clubPrefix;

    const timeout = parseInt(document.getElementById('timeout')?.value) || 30;
    params.timeout = timeout;

    const concurrency = parseInt(document.getElementById('concurrency')?.value) || 4;
    params.concurrency = concurrency;

    const verbose = document.getElementById('verbose')?.checked;
    if (verbose) params.verbose = verbose;

    // Mode-specific parameters
    if (mode === 'statistical' || mode === 'hybrid') {
        const minSampleSize = parseInt(document.getElementById('min-sample-size')?.value) || 100;
        params.min_sample_size = minSampleSize;

        // Collect statistical formats
        const formats = [];
        if (document.getElementById('csv-stats')?.checked) formats.push('csv');
        if (document.getElementById('json-stats')?.checked) formats.push('json');
        if (formats.length > 0) params.formats = formats;
    }

    if (mode === 'detailed' || mode === 'hybrid' || mode === 'efficient') {
        const outputFormat = document.getElementById('output-format')?.value || 'csv';
        params.output_format = outputFormat;
    }

    return params;
}

function getModeDisplayName(mode) {
    switch (mode) {
        case 'statistical': return 'Statistische Analyse';
        case 'hybrid': return 'Hybrid-Analyse';
        case 'efficient': return 'Effizienter Modus';
        case 'detailed':
        default: return 'Detailierte Analyse';
    }
}

// Enhanced status display with processing mode information
function displayStatusEnhanced(response) {
    const statusContainer = document.getElementById('status-result');
    const status = response.status;

    // Determine status indicator and text
    let statusIndicator, statusText, statusClass;

    if (status.running || status.Running) {
        statusIndicator = '🔄';
        statusText = 'Läuft gerade...';
        statusClass = 'status-running';
    } else {
        statusIndicator = '✅';
        statusText = 'Abgeschlossen';
        statusClass = 'status-completed';
    }

    // Processing mode info
    const processingMode = status.processing_mode || status.ProcessingMode || 0;
    let modeText = 'Detailierte Analyse';
    switch (processingMode) {
        case 1: modeText = 'Statistische Analyse'; break;
        case 2: modeText = 'Hybrid-Analyse'; break;
        case 3: modeText = 'Effizienter Modus'; break;
    }

    // Format dates
    const lastExecution = (status.last_execution || status.LastExecution) ?
        formatDateTime(status.last_execution || status.LastExecution) : 'Noch nie ausgeführt';
    const nextScheduled = 'Täglich nach Datenbankimport';
    const startTime = (status.start_time || status.StartTime) ?
        formatDateTime(status.start_time || status.StartTime) : null;
    const lastSuccess = (status.last_success || status.LastSuccess) ?
        formatDateTime(status.last_success || status.LastSuccess) : null;
    const lastError = status.last_error || status.LastError;

    statusContainer.innerHTML = `
        <div class="status-info">
            <div class="status-row ${statusClass}">
                <span class="status-indicator">${statusIndicator}</span>
                <strong>Status:</strong> ${statusText}
            </div>
            <div class="status-row">
                <span class="status-indicator">⚙️</span>
                <strong>Modus:</strong> ${modeText}
            </div>
            <div class="status-row">
                <span class="status-indicator">📅</span>
                <strong>Letzte Ausführung:</strong> ${lastExecution}
            </div>
            <div class="status-row">
                <span class="status-indicator">⏰</span>
                <strong>Nächste Ausführung:</strong> ${nextScheduled}
            </div>
            ${(status.running || status.Running) && startTime ? `
                <div class="status-row">
                    <span class="status-indicator">⏱️</span>
                    <strong>Gestartet:</strong> ${startTime}
                </div>
            ` : ''}
            ${lastSuccess ? `
                <div class="status-row">
                    <span class="status-indicator">✅</span>
                    <strong>Letzter Erfolg:</strong> ${lastSuccess}
                </div>
            ` : ''}
            ${lastError ? `
                <div class="status-row">
                    <span class="status-indicator">⚠️</span>
                    <strong>Letzter Fehler:</strong> ${lastError}
                </div>
            ` : ''}
            <div class="status-row">
                <span class="status-indicator">📊</span>
                <strong>Verfügbare Dateien:</strong> ${response.available_files ? response.available_files.length : 0}
            </div>
            ${status.output_files && status.output_files.length > 0 ? `
                <div class="status-row">
                    <span class="status-indicator">📁</span>
                    <strong>Letzte Ausgabedateien:</strong> ${status.output_files.slice(0, 3).join(', ')}${status.output_files.length > 3 ? '...' : ''}
                </div>
            ` : ''}
        </div>
    `;
}

// Update existing displayStatus function to use enhanced version
function displayStatus(response) {
    displayStatusEnhanced(response);
}

// Cleanup on page unload
window.addEventListener('beforeunload', function() {
    if (statusRefreshInterval) {
        clearInterval(statusRefreshInterval);
    }
});