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

// Cleanup on page unload
window.addEventListener('beforeunload', function() {
    if (statusRefreshInterval) {
        clearInterval(statusRefreshInterval);
    }
});