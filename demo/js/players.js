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
            Utils.showError('player-search-results', `Search failed: ${error.message}`);
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
                Utils.showError('player-lookup-results', 'Invalid player ID format. Expected format: CLUBID-PERSONID (e.g., C0101-1014, UNKNOWN-10799083)');
                return;
            }
            
            const result = await api.getPlayer(playerId, format);
            
            if (format === 'json') {
                // API returns {success: true, data: {...}}
                this.displayPlayerDetail('player-lookup-results', result.data || result);
            } else {
                new CodeDisplayManager().displayResponse('player-lookup-results', result, 'Player Data (CSV)');
            }
            
        } catch (error) {
            Utils.showError('player-lookup-results', `Player lookup failed: ${error.message}`);
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
                Utils.showError('rating-history-results', 'Invalid player ID format. Expected format: CLUBID-PERSONID (e.g., C0101-1014, UNKNOWN-10799083)');
                return;
            }
            
            const result = await api.getPlayerRatingHistory(playerId, format);
            
            if (format === 'json') {
                // API returns {success: true, data: [...]}
                this.displayRatingHistory('rating-history-results', result.data || result);
            } else {
                new CodeDisplayManager().displayResponse('rating-history-results', result, 'Rating History (CSV)');
            }
            
        } catch (error) {
            Utils.showError('rating-history-results', `Rating history lookup failed: ${error.message}`);
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
                Utils.showError('club-players-results', 'Invalid club ID format. Expected format: C0101');
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
            Utils.showError('club-players-results', `Club players lookup failed: ${error.message}`);
        }
    }

    displayPlayerResults(containerId, players, meta) {
        const container = document.getElementById(containerId);
        if (!container) return;

        if (!players || players.length === 0) {
            container.innerHTML = `
                <div class="alert alert-info">
                    <h4>No Results</h4>
                    <p>No players found matching your search criteria.</p>
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
                            <th>Player ID</th>
                            <th>Name</th>
                            <th>Club</th>
                            <th>DWZ Rating</th>
                            <th>Birth Year</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
        `;

        players.forEach(player => {
            // Extract birth year from birth date
            let birthYear = 'N/A';
            if (player.birth) {
                const birthDate = new Date(player.birth);
                if (!isNaN(birthDate.getTime())) {
                    birthYear = birthDate.getFullYear().toString();
                }
            }

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
                            View Details
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

        container.innerHTML = html;
    }

    displayPlayerDetail(containerId, player) {
        const container = document.getElementById(containerId);
        if (!container) return;

        // Extract birth year from birth date
        let birthYear = 'N/A';
        if (player.birth) {
            const birthDate = new Date(player.birth);
            if (!isNaN(birthDate.getTime())) {
                birthYear = birthDate.getFullYear().toString();
            }
        }

        const html = `
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Player Details: ${Utils.sanitizeHTML(((player.firstname || '') + ' ' + (player.name || '')).trim() || 'Unknown')}</h3>
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
                        <p><strong>Birth Date:</strong> ${Utils.formatDate(player.birth)}</p>
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
                            <th>Evaluation ID</th>
                            <th>Old DWZ</th>
                            <th>New DWZ</th>
                            <th>Rating Change</th>
                            <th>Games</th>
                            <th>Points</th>
                            <th>Achievement</th>
                            <th>Tournament ID</th>
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
                    <td>${Utils.sanitizeHTML(entry.id_master || 'N/A')}</td>
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
            
            modalBody.innerHTML = '<div class="loading show"><div class="spinner"></div><p>Loading player details...</p></div>';
            
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