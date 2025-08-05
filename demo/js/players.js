// Player-specific functionality

class PlayerManager {
    constructor() {
        this.currentResults = null;
        this.currentMeta = null;
        this.init();
    }

    init() {
        // Search form handler
        const searchForm = document.getElementById('player-search-form');
        if (searchForm) {
            searchForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.searchPlayers();
            });
        }

        // Individual player lookup form
        const playerForm = document.getElementById('player-lookup-form');
        if (playerForm) {
            playerForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.getPlayer();
            });
        }

        // Rating history form
        const historyForm = document.getElementById('rating-history-form');
        if (historyForm) {
            historyForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.getPlayerRatingHistory();
            });
        }

        // Club players form
        const clubPlayersForm = document.getElementById('club-players-form');
        if (clubPlayersForm) {
            clubPlayersForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.getClubPlayers();
            });
        }

        // Handle URL parameters for direct navigation from clubs page
        this.handleURLParameters();
    }

    handleURLParameters() {
        const urlParams = new URLSearchParams(window.location.search);
        const hash = window.location.hash;
        
        console.log('URL Params:', urlParams.toString());
        console.log('Hash:', hash);
        
        // Check if we have club_id parameter and club-players hash
        const clubId = urlParams.get('club_id');
        console.log('Club ID from URL:', clubId);
        
        if (clubId && hash.includes('club-players')) {
            console.log('Auto-loading club players for:', clubId);
            
            // Switch to club-players tab
            const clubPlayersTab = document.querySelector('[data-tab="club-players"]');
            if (clubPlayersTab) {
                console.log('Found club-players tab, clicking...');
                clubPlayersTab.click();
            } else {
                console.log('Club-players tab not found');
            }
            
            // Populate club ID field and automatically search
            setTimeout(() => {
                const clubIdField = document.getElementById('club-id');
                if (clubIdField) {
                    console.log('Found club-id field, setting value:', clubId);
                    clubIdField.value = clubId;
                    // Automatically trigger the search
                    console.log('Triggering getClubPlayers...');
                    this.getClubPlayers();
                } else {
                    console.log('Club-id field not found');
                }
            }, 500); // Increased timeout to 500ms
        } else {
            console.log('No auto-load conditions met. ClubId:', clubId, 'Hash includes club-players:', hash.includes('club-players'));
        }
    }

    async searchPlayers() {
        try {
            Utils.showLoading('player-search-results');
            
            const form = document.getElementById('player-search-form');
            const params = Utils.getFormData(form);
            
            const result = await api.searchPlayers(params);
            // Fix: API returns {success: true, data: {data: [...], meta: {...}}}
            this.currentResults = result.data?.data || result.data || result;
            this.currentMeta = result.data?.meta || result.meta;
            
            this.displayPlayerResults('player-search-results', this.currentResults, this.currentMeta);
            
        } catch (error) {
            Utils.showError('player-search-results', `Suche fehlgeschlagen: ${error.message}`);
        }
    }

    async getPlayer() {
        try {
            Utils.showLoading('player-lookup-results');
            
            const form = document.getElementById('player-lookup-form');
            const formData = Utils.getFormData(form);
            const playerId = formData.player_id;
            const format = formData.format || 'json';
            
            // Validate player ID
            if (!Utils.validatePlayerID(playerId)) {
                Utils.showError('player-lookup-results', 'Ungültiges Spieler-ID Format. Erwartetes Format: CLUBID-PERSONID (z.B., C0101-1014, UNKNOWN-10799083)');
                return;
            }
            
            const result = await api.getPlayer(playerId, format);
            
            if (format === 'json') {
                // API returns {success: true, data: {...}}
                this.displayPlayerDetail('player-lookup-results', result.data || result);
            } else {
                new CodeDisplayManager().displayResponse('player-lookup-results', result, 'Spielerdaten (CSV)');
            }
            
        } catch (error) {
            Utils.showError('player-lookup-results', `Spieler-Abfrage fehlgeschlagen: ${error.message}`);
        }
    }

    async getPlayerRatingHistory() {
        try {
            Utils.showLoading('rating-history-results');
            
            const form = document.getElementById('rating-history-form');
            const formData = Utils.getFormData(form);
            const playerId = formData.player_id;
            const format = formData.format || 'json';
            
            // Validate player ID
            if (!Utils.validatePlayerID(playerId)) {
                Utils.showError('rating-history-results', 'Ungültiges Spieler-ID Format. Erwartetes Format: CLUBID-PERSONID (z.B., C0101-1014, UNKNOWN-10799083)');
                return;
            }
            
            const result = await api.getPlayerRatingHistory(playerId, format);
            
            if (format === 'json') {
                // API returns {success: true, data: [...]}
                this.displayRatingHistory('rating-history-results', result.data || result);
            } else {
                new CodeDisplayManager().displayResponse('rating-history-results', result, 'Bewertungsverlauf (CSV)');
            }
            
        } catch (error) {
            Utils.showError('rating-history-results', `Bewertungsverlauf-Abfrage fehlgeschlagen: ${error.message}`);
        }
    }

    async getClubPlayers() {
        try {
            Utils.showLoading('club-players-results');
            
            const form = document.getElementById('club-players-form');
            const formData = Utils.getFormData(form);
            const clubId = formData.club_id;
            
            // Validate club ID
            if (!Utils.validateClubID(clubId)) {
                Utils.showError('club-players-results', 'Ungültiges Vereins-ID Format. Erwartetes Format: C0101');
                return;
            }
            
            // Remove club_id from params before passing to API
            delete formData.club_id;
            
            const result = await api.getClubPlayers(clubId, formData);
            // Fix: API returns {success: true, data: {data: [...], meta: {...}}}
            this.currentResults = result.data?.data || result.data || result;
            this.currentMeta = result.data?.meta || result.meta;
            
            this.displayPlayerResults('club-players-results', this.currentResults, this.currentMeta);
            
        } catch (error) {
            Utils.showError('club-players-results', `Vereinsspieler-Abfrage fehlgeschlagen: ${error.message}`);
        }
    }

    displayPlayerResults(containerId, players, meta) {
        const container = document.getElementById(containerId);
        if (!container) return;

        if (!players || players.length === 0) {
            container.innerHTML = `
                <div class="alert alert-info">
                    <h4>Keine Ergebnisse</h4>
                    <p>Keine Spieler gefunden, die Ihren Suchkriterien entsprechen.</p>
                </div>
            `;
            return;
        }

        let html = `
            <div class="alert alert-info">
                <p>${Utils.getPaginationInfo(meta)}</p>
            </div>
            <div class="table-container">
                <table class="table table-striped">
                    <thead>
                        <tr>
                            <th>Spieler-ID</th>
                            <th>Name</th>
                            <th>Verein</th>
                            <th>DWZ Bewertung</th>
                            <th>Geburtsjahr</th>
                            <th>Aktionen</th>
                        </tr>
                    </thead>
                    <tbody>
        `;

        players.forEach(player => {
            // Use birth year directly from API (GDPR compliant)
            const birthYear = Utils.formatBirthYear(player.birth_year);

            html += `
                <tr>
                    <td><code>${Utils.sanitizeHTML(player.id || 'N/A')}</code></td>
                    <td><strong>${Utils.sanitizeHTML(((player.firstname || '') + ' ' + (player.name || '')).trim() || 'N/A')}</strong></td>
                    <td>${Utils.sanitizeHTML(player.club || 'N/A')}</td>
                    <td>
                        ${player.current_dwz ? `<span class="badge badge-primary">${player.current_dwz}</span>` : 'N/A'}
                    </td>
                    <td>${birthYear}</td>
                    <td>
                        <button onclick="playerManager.viewPlayerDetail('${player.id}')" class="btn btn-small btn-secondary">
                            Details anzeigen
                        </button>
                    </td>
                </tr>
            `;
        });

        html += `
                    </tbody>
                </table>
            </div>
        `;

        // Add pagination controls
        const paginationCallback = this.getPaginationCallback(containerId);
        html += Utils.createPaginationControls(meta, containerId, paginationCallback);

        container.innerHTML = html;
    }

    // Get the appropriate pagination callback function name based on container ID
    getPaginationCallback(containerId) {
        switch(containerId) {
            case 'player-search-results':
                return 'playerManager.searchPlayersWithOffset';
            case 'club-players-results':
                return 'playerManager.getClubPlayersWithOffset';
            default:
                return 'playerManager.searchPlayersWithOffset';
        }
    }

    // Pagination handler for player search
    async searchPlayersWithOffset(offset, containerId) {
        try {
            Utils.showLoading(containerId);
            
            const form = document.getElementById('player-search-form');
            const params = Utils.getFormData(form);
            params.offset = offset;
            
            const result = await api.searchPlayers(params);
            this.currentResults = result.data?.data || result.data || result;
            this.currentMeta = result.data?.meta || result.meta;
            
            this.displayPlayerResults(containerId, this.currentResults, this.currentMeta);
            
        } catch (error) {
            Utils.showError(containerId, `Suche fehlgeschlagen: ${error.message}`);
        }
    }

    // Pagination handler for club players
    async getClubPlayersWithOffset(offset, containerId) {
        try {
            Utils.showLoading(containerId);
            
            const form = document.getElementById('club-players-form');
            const formData = Utils.getFormData(form);
            const clubId = formData.club_id;
            
            // Validate club ID
            if (!Utils.validateClubID(clubId)) {
                Utils.showError(containerId, 'Ungültiges Vereins-ID Format. Erwartetes Format: C0101');
                return;
            }
            
            // Remove club_id from params and add offset
            delete formData.club_id;
            formData.offset = offset;
            
            const result = await api.getClubPlayers(clubId, formData);
            this.currentResults = result.data?.data || result.data || result;
            this.currentMeta = result.data?.meta || result.meta;
            
            this.displayPlayerResults(containerId, this.currentResults, this.currentMeta);
            
        } catch (error) {
            Utils.showError(containerId, `Vereinsspieler-Abfrage fehlgeschlagen: ${error.message}`);
        }
    }

    displayPlayerDetail(containerId, player) {
        const container = document.getElementById(containerId);
        if (!container) return;

        // Use birth year directly from API (GDPR compliant) 
        const birthYear = Utils.formatBirthYear(player.birth_year);

        const html = `
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Spielerdetails: ${Utils.sanitizeHTML(((player.firstname || '') + ' ' + (player.name || '')).trim() || 'Unbekannt')}</h3>
                </div>
                <div class="form-row">
                    <div>
                        <h4>Basic Information</h4>
                        <p><strong>Player ID:</strong> <code>${Utils.sanitizeHTML(player.id || 'N/A')}</code></p>
                        <p><strong>Name:</strong> ${Utils.sanitizeHTML(player.name || 'N/A')}</p>
                        <p><strong>First Name:</strong> ${Utils.sanitizeHTML(player.firstname || 'N/A')}</p>
                        <p><strong>Birth Year:</strong> ${birthYear}</p>
                        <p><strong>Gender:</strong> ${Utils.sanitizeHTML(player.gender || 'N/A')}</p>
                        <p><strong>Nation:</strong> ${Utils.sanitizeHTML(player.nation || 'N/A')}</p>
                    </div>
                    <div>
                        <h4>Club Information</h4>
                        <p><strong>Club ID:</strong> <code>${Utils.sanitizeHTML(player.club_id || 'N/A')}</code></p>
                        <p><strong>Club Name:</strong> ${Utils.sanitizeHTML(player.club || 'N/A')}</p>
                        <p><strong>Status:</strong> ${Utils.sanitizeHTML(player.status || 'N/A')}</p>
                    </div>
                </div>
                <div class="form-row">
                    <div>
                        <h4>DWZ Rating</h4>
                        <p><strong>Current DWZ:</strong> 
                            ${player.current_dwz ? `<span class="badge badge-primary">${player.current_dwz}</span>` : 'N/A'}
                        </p>
                        <p><strong>DWZ Index:</strong> ${Utils.sanitizeHTML(player.dwz_index || 'N/A')}</p>
                    </div>
                    <div>
                        <h4>Additional Information</h4>
                        <p><strong>FIDE ID:</strong> ${Utils.sanitizeHTML(player.fide_id || 'N/A')}</p>
                        <p><strong>Geburtsjahr:</strong> ${Utils.formatBirthYear(player.birth_year)}</p>
                    </div>
                </div>
            </div>
        `;

        container.innerHTML = html;
    }

    displayRatingHistory(containerId, history) {
        const container = document.getElementById(containerId);
        if (!container) return;

        if (!history || history.length === 0) {
            container.innerHTML = `
                <div class="alert alert-info">
                    <h4>No Rating History</h4>
                    <p>No rating history found for this player.</p>
                </div>
            `;
            return;
        }

        let html = `
            <div class="table-container">
                <table class="table table-striped">
                    <thead>
                        <tr>
                            <th>Auswertungs-ID</th>
                            <th>Alte DWZ</th>
                            <th>Neue DWZ</th>
                            <th>Bewertungsänderung</th>
                            <th>Spiele</th>
                            <th>Punkte</th>
                            <th>Leistung</th>
                            <th>Turnier-ID</th>
                        </tr>
                    </thead>
                    <tbody>
        `;

        history.forEach(entry => {
            const ratingChange = entry.dwz_new - entry.dwz_old;
            
            html += `
                <tr>
                    <td>${Utils.sanitizeHTML(entry.id || 'N/A')}</td>
                    <td><span class="badge badge-secondary">${entry.dwz_old || 'N/A'}</span></td>
                    <td><span class="badge badge-primary">${entry.dwz_new || 'N/A'}</span></td>
                    <td>
                        ${ratingChange !== 0 ? 
                            `<span class="badge ${ratingChange > 0 ? 'badge-success' : 'badge-danger'}">
                                ${ratingChange > 0 ? '+' : ''}${ratingChange}
                            </span>` : 
                            '<span class="badge badge-secondary">0</span>'
                        }
                    </td>
                    <td>${Utils.sanitizeHTML(entry.games || 'N/A')}</td>
                    <td>${Utils.sanitizeHTML(entry.points || 'N/A')}</td>
                    <td>${Utils.sanitizeHTML(entry.achievement || 'N/A')}</td>
                    <td><code>${Utils.sanitizeHTML(entry.tournament_id || 'N/A')}</code></td>
                </tr>
            `;
        });

        html += `
                    </tbody>
                </table>
            </div>
        `;

        container.innerHTML = html;
    }

    async viewPlayerDetail(playerId) {
        try {
            const modal = document.getElementById('player-detail-modal');
            const modalBody = modal.querySelector('.modal-body');
            
            modalBody.innerHTML = '<div class="loading show"><div class="spinner"></div><p>Lade Spielerdetails...</p></div>';
            
            new ModalManager().openModal('player-detail-modal');
            
            const result = await api.getPlayer(playerId);
            this.displayPlayerDetail('player-detail-content', result.data || result);
            modalBody.innerHTML = '<div id="player-detail-content"></div>';
            this.displayPlayerDetail('player-detail-content', result.data || result);
            
        } catch (error) {
            const modalBody = document.querySelector('#player-detail-modal .modal-body');
            modalBody.innerHTML = `<div class="alert alert-error"><p>Failed to load player details: ${error.message}</p></div>`;
        }
    }
}

// Initialize player manager when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    if (document.querySelector('.player-page')) {
        window.playerManager = new PlayerManager();
    }
});