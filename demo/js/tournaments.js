// Tournament-specific functionality

class TournamentManager {
    constructor() {
        this.currentResults = null;
        this.currentMeta = null;
        this.init();
    }

    init() {
        // Search tournaments form handler
        const searchForm = document.getElementById('tournament-search-form');
        if (searchForm) {
            searchForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.searchTournaments();
            });
        }

        // Upcoming tournaments form handler
        const upcomingForm = document.getElementById('upcoming-tournaments-form');
        if (upcomingForm) {
            upcomingForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.getUpcomingTournaments();
            });
        }

        // Recent tournaments form handler
        const recentForm = document.getElementById('recent-tournaments-form');
        if (recentForm) {
            recentForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.getRecentTournaments();
            });
        }

        // Date range tournaments form handler
        const dateRangeForm = document.getElementById('date-range-tournaments-form');
        if (dateRangeForm) {
            dateRangeForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.getTournamentsByDateRange();
            });
        }

        // Individual tournament lookup form
        const tournamentForm = document.getElementById('tournament-lookup-form');
        if (tournamentForm) {
            tournamentForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.getTournament();
            });
        }
    }

    async searchTournaments() {
        try {
            Utils.showLoading('tournament-search-results');
            
            const form = document.getElementById('tournament-search-form');
            const params = Utils.getFormData(form);
            
            const result = await api.searchTournaments(params);
            // Fix: API returns {success: true, data: {data: [...], meta: {...}}}
            this.currentResults = result.data?.data || result.data || result;
            this.currentMeta = result.data?.meta || result.meta;
            
            this.displayTournamentResults('tournament-search-results', this.currentResults, this.currentMeta);
            
        } catch (error) {
            Utils.showError('tournament-search-results', `Suche fehlgeschlagen: ${error.message}`);
        }
    }

    async getUpcomingTournaments() {
        try {
            Utils.showLoading('upcoming-tournaments-results');
            
            const form = document.getElementById('upcoming-tournaments-form');
            const formData = Utils.getFormData(form);
            const limit = formData.limit || 20;
            const format = formData.format || 'json';
            
            const result = await api.getUpcomingTournaments(limit, format);
            
            if (format === 'json') {
                // Fix: API returns {success: true, data: [...]} for upcoming tournaments
                this.displayTournamentResults('upcoming-tournaments-results', result.data || result);
            } else {
                new CodeDisplayManager().displayResponse('upcoming-tournaments-results', result, 'Kommende Turniere (CSV)');
            }
            
        } catch (error) {
            Utils.showError('upcoming-tournaments-results', `Fehler beim Abrufen kommender Turniere: ${error.message}`);
        }
    }

    async getRecentTournaments() {
        try {
            Utils.showLoading('recent-tournaments-results');
            
            const form = document.getElementById('recent-tournaments-form');
            const formData = Utils.getFormData(form);
            const days = formData.days || 30;
            const limit = formData.limit || 20;
            const format = formData.format || 'json';
            
            const result = await api.getRecentTournaments(days, limit, format);
            
            if (format === 'json') {
                // Fix: API returns {success: true, data: [...]} for recent tournaments
                this.displayTournamentResults('recent-tournaments-results', result.data || result);
            } else {
                new CodeDisplayManager().displayResponse('recent-tournaments-results', result, 'Kürzliche Turniere (CSV)');
            }
            
        } catch (error) {
            Utils.showError('recent-tournaments-results', `Fehler beim Abrufen kürzlicher Turniere: ${error.message}`);
        }
    }

    async getTournamentsByDateRange() {
        try {
            Utils.showLoading('date-range-tournaments-results');
            
            const form = document.getElementById('date-range-tournaments-form');
            const formData = Utils.getFormData(form);
            const startDate = formData.start_date;
            const endDate = formData.end_date;
            
            if (!startDate || !endDate) {
                Utils.showError('date-range-tournaments-results', 'Sowohl Start- als auch Enddatum sind erforderlich.');
                return;
            }
            
            // Remove dates from params before passing to API
            delete formData.start_date;
            delete formData.end_date;
            
            const result = await api.getTournamentsByDateRange(startDate, endDate, formData);
            // Fix: API returns {success: true, data: {data: [...], meta: {...}}}
            this.currentResults = result.data?.data || result.data || result;
            this.currentMeta = result.data?.meta || result.meta;
            
            this.displayTournamentResults('date-range-tournaments-results', this.currentResults, this.currentMeta);
            
        } catch (error) {
            Utils.showError('date-range-tournaments-results', `Fehler beim Abrufen der Turniere nach Datumsbereich: ${error.message}`);
        }
    }

    async getTournament() {
        try {
            Utils.showLoading('tournament-lookup-results');
            
            const form = document.getElementById('tournament-lookup-form');
            const formData = Utils.getFormData(form);
            const tournamentId = formData.tournament_id;
            const format = formData.format || 'json';
            
            // Validate tournament ID
            if (!Utils.validateTournamentID(tournamentId)) {
                Utils.showError('tournament-lookup-results', 'Ungültiges Turnier-ID Format. Erwartetes Format: C529-K00-HT1 oder C339-400-442');
                return;
            }
            
            const result = await api.getTournament(tournamentId, format);
            
            if (format === 'json') {
                // API returns {success: true, data: {...}}
                this.displayTournamentDetail('tournament-lookup-results', result.data || result);
            } else {
                new CodeDisplayManager().displayResponse('tournament-lookup-results', result, 'Turnierdaten (CSV)');
            }
            
        } catch (error) {
            Utils.showError('tournament-lookup-results', `Turnier-Abfrage fehlgeschlagen: ${error.message}`);
        }
    }

    displayTournamentResults(containerId, tournaments, meta) {
        const container = document.getElementById(containerId);
        if (!container) return;

        if (!tournaments || tournaments.length === 0) {
            container.innerHTML = `
                <div class="alert alert-info">
                    <h4>Keine Ergebnisse</h4>
                    <p>Keine Turniere gefunden, die Ihren Kriterien entsprechen.</p>
                </div>
            `;
            return;
        }

        let html = '';
        
        if (meta) {
            html += `
                <div class="alert alert-info">
                    <p>${Utils.getPaginationInfo(meta)}</p>
                </div>
            `;
        }

        html += `
            <div class="table-container">
                <table class="table table-striped">
                    <thead>
                        <tr>
                            <th>Turnier-ID</th>
                            <th>Name</th>
                            <th>Startdatum</th>
                            <th>Enddatum</th>
                            <th>Veranstalter</th>
                            <th>Aktionen</th>
                        </tr>
                    </thead>
                    <tbody>
        `;

        tournaments.forEach(tournament => {
            html += `
                <tr>
                    <td><code>${Utils.sanitizeHTML(tournament.id || tournament.code || 'N/A')}</code></td>
                    <td><strong>${Utils.sanitizeHTML(tournament.name || 'N/A')}</strong></td>
                    <td>${Utils.formatDate(tournament.start_date)}</td>
                    <td>${Utils.formatDate(tournament.end_date)}</td>
                    <td>${Utils.sanitizeHTML(tournament.organization || 'N/A')}</td>
                    <td>
                        <button onclick="tournamentManager.viewTournamentDetail('${tournament.id || tournament.code}')" class="btn btn-small btn-secondary">
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

        // Add pagination controls if meta exists
        if (meta) {
            const paginationCallback = this.getPaginationCallback(containerId);
            html += Utils.createPaginationControls(meta, containerId, paginationCallback);
        }

        container.innerHTML = html;
    }

    // Get the appropriate pagination callback function name based on container ID
    getPaginationCallback(containerId) {
        switch(containerId) {
            case 'tournament-search-results':
                return 'tournamentManager.searchTournamentsWithOffset';
            case 'date-range-tournaments-results':
                return 'tournamentManager.getTournamentsByDateRangeWithOffset';
            default:
                return 'tournamentManager.searchTournamentsWithOffset';
        }
    }

    // Pagination handler for tournament search
    async searchTournamentsWithOffset(offset, containerId) {
        try {
            Utils.showLoading(containerId);
            
            const form = document.getElementById('tournament-search-form');
            const params = Utils.getFormData(form);
            params.offset = offset;
            
            const result = await api.searchTournaments(params);
            this.currentResults = result.data?.data || result.data || result;
            this.currentMeta = result.data?.meta || result.meta;
            
            this.displayTournamentResults(containerId, this.currentResults, this.currentMeta);
            
        } catch (error) {
            Utils.showError(containerId, `Suche fehlgeschlagen: ${error.message}`);
        }
    }

    // Pagination handler for date range tournaments
    async getTournamentsByDateRangeWithOffset(offset, containerId) {
        try {
            Utils.showLoading(containerId);
            
            const form = document.getElementById('date-range-tournaments-form');
            const formData = Utils.getFormData(form);
            const startDate = formData.start_date;
            const endDate = formData.end_date;
            
            if (!startDate || !endDate) {
                Utils.showError(containerId, 'Both start date and end date are required.');
                return;
            }
            
            // Remove dates from params and add offset
            delete formData.start_date;
            delete formData.end_date;
            formData.offset = offset;
            
            const result = await api.getTournamentsByDateRange(startDate, endDate, formData);
            this.currentResults = result.data?.data || result.data || result;
            this.currentMeta = result.data?.meta || result.meta;
            
            this.displayTournamentResults(containerId, this.currentResults, this.currentMeta);
            
        } catch (error) {
            Utils.showError(containerId, `Fehler beim Abrufen der Turniere nach Datumsbereich: ${error.message}`);
        }
    }

    displayTournamentDetail(containerId, tournament) {
        const container = document.getElementById(containerId);
        if (!container) return;

        const html = `
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Turnierdetails: ${Utils.sanitizeHTML(tournament.name || 'Unbekannt')}</h3>
                </div>
                <div class="form-row">
                    <div>
                        <h4>Basic Information</h4>
                        <p><strong>Tournament ID:</strong> <code>${Utils.sanitizeHTML(tournament.tournament_id || tournament.code || 'N/A')}</code></p>
                        <p><strong>Name:</strong> ${Utils.sanitizeHTML(tournament.name || 'N/A')}</p>
                        <p><strong>Type:</strong> ${Utils.sanitizeHTML(tournament.type || 'N/A')}</p>
                        <p><strong>Format:</strong> ${Utils.sanitizeHTML(tournament.format || 'N/A')}</p>
                    </div>
                    <div>
                        <h4>Schedule</h4>
                        <p><strong>Start Date:</strong> ${Utils.formatDate(tournament.start_date || tournament.startedOn)}</p>
                        <p><strong>End Date:</strong> ${Utils.formatDate(tournament.end_date || tournament.finishedOn)}</p>
                        <p><strong>Registration Deadline:</strong> ${Utils.formatDate(tournament.registration_deadline)}</p>
                        <p><strong>Status:</strong> 
                            <span class="badge ${tournament.status === 'finished' ? 'badge-secondary' : 'badge-primary'}">
                                ${Utils.sanitizeHTML(tournament.status || 'N/A')}
                            </span>
                        </p>
                    </div>
                </div>
                <div class="form-row">
                    <div>
                        <h4>Organizer Information</h4>
                        <p><strong>Organizer:</strong> ${Utils.sanitizeHTML(tournament.organizer || tournament.club_name || 'N/A')}</p>
                        <p><strong>Club ID:</strong> <code>${Utils.sanitizeHTML(tournament.organizer_club_id || tournament.club_id || 'N/A')}</code></p>
                        <p><strong>Contact:</strong> ${Utils.sanitizeHTML(tournament.contact_person || 'N/A')}</p>
                        <p><strong>Email:</strong> ${tournament.contact_email ? `<a href="mailto:${tournament.contact_email}">${Utils.sanitizeHTML(tournament.contact_email)}</a>` : 'N/A'}</p>
                    </div>
                    <div>
                        <h4>Turnierdetails</h4>
                        <p><strong>Teilnehmer:</strong> 
                            ${tournament.participant_count ? `<span class="badge badge-primary">${tournament.participant_count}</span>` : 'N/A'}
                        </p>
                        <p><strong>Rounds:</strong> ${Utils.sanitizeHTML(tournament.rounds || 'N/A')}</p>
                        <p><strong>Time Control:</strong> ${Utils.sanitizeHTML(tournament.time_control || 'N/A')}</p>
                        <p><strong>Prize Fund:</strong> ${Utils.sanitizeHTML(tournament.prize_fund || 'N/A')}</p>
                    </div>
                </div>
                ${tournament.description ? `
                    <div class="mt-3">
                        <h4>Description</h4>
                        <p>${Utils.sanitizeHTML(tournament.description)}</p>
                    </div>
                ` : ''}
                ${tournament.website ? `
                    <div class="mt-3">
                        <a href="${tournament.website}" target="_blank" class="btn btn-primary">
                            Visit Tournament Website
                        </a>
                    </div>
                ` : ''}
            </div>
        `;

        container.innerHTML = html;
    }

    async viewTournamentDetail(tournamentId) {
        try {
            const modal = document.getElementById('tournament-detail-modal');
            const modalBody = modal.querySelector('.modal-body');
            
            modalBody.innerHTML = '<div class="loading show"><div class="spinner"></div><p>Lade Turnierdetails...</p></div>';
            
            new ModalManager().openModal('tournament-detail-modal');
            
            const result = await api.getTournament(tournamentId);
            modalBody.innerHTML = '<div id="tournament-detail-content"></div>';
            this.displayTournamentDetail('tournament-detail-content', result.data || result);
            
        } catch (error) {
            const modalBody = document.querySelector('#tournament-detail-modal .modal-body');
            modalBody.innerHTML = `<div class="alert alert-error"><p>Failed to load tournament details: ${error.message}</p></div>`;
        }
    }
}

// Initialize tournament manager when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    if (document.querySelector('.tournament-page')) {
        window.tournamentManager = new TournamentManager();
    }
});