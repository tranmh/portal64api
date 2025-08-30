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

        // Tournament players form handler
        const tournamentPlayersForm = document.getElementById('tournament-players-form');
        if (tournamentPlayersForm) {
            tournamentPlayersForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.getTournamentPlayers();
            });
        }

        // Check for tournament parameter in URL and auto-load tournament details
        const urlParams = new URLSearchParams(window.location.search);
        const tournamentId = urlParams.get('tournament');
        if (tournamentId) {
            // Auto-load tournament details and switch to that tab
            this.autoLoadTournamentDetails(tournamentId);
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

    async getRecentTournaments() {
        try {
            Utils.showLoading('recent-tournaments-results');
            
            const form = document.getElementById('recent-tournaments-form');
            const formData = Utils.getFormData(form);
            const days = formData.days || 30;
            const limit = formData.limit || 500;
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
                Utils.showError('tournament-lookup-results', 'Ungültiges Turnier-ID Format. Erwartetes Format: B718-A08-BEL oder C529-K00-HT1');
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

    async getTournamentPlayers() {
        try {
            Utils.showLoading('tournament-players-results');
            
            const form = document.getElementById('tournament-players-form');
            const formData = Utils.getFormData(form);
            const tournamentId = formData.tournament_id;
            const format = formData.format || 'json';
            
            // Validate tournament ID
            if (!Utils.validateTournamentID(tournamentId)) {
                Utils.showError('tournament-players-results', 'Ungültiges Turnier-ID Format. Erwartetes Format: B718-A08-BEL oder C529-K00-HT1');
                return;
            }
            
            const result = await api.getTournamentPlayers(tournamentId, format);
            
            if (format === 'json') {
                // API returns {success: true, data: {...}}
                this.displayTournamentPlayersDetail('tournament-players-results', result.data || result);
            } else {
                new CodeDisplayManager().displayResponse('tournament-players-results', result, 'Turnier-Spielerdaten (CSV)');
            }
            
        } catch (error) {
            Utils.showError('tournament-players-results', `Turnier-Spieler-Abfrage fehlgeschlagen: ${error.message}`);
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

    displayTournamentPlayersDetail(containerId, tournament) {
        const container = document.getElementById(containerId);
        if (!container) return;

        if (!tournament || !tournament.participants) {
            container.innerHTML = `
                <div class="alert alert-info">
                    <h4>Keine Teilnehmer gefunden</h4>
                    <p>Für dieses Turnier wurden keine Teilnehmerdaten gefunden.</p>
                </div>
            `;
            return;
        }

        let html = `
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Turnier: ${Utils.sanitizeHTML(tournament.name || 'Unbekannt')}</h3>
                    <p><strong>ID:</strong> <code>${Utils.sanitizeHTML(tournament.id || 'N/A')}</code> | 
                       <strong>Teilnehmer:</strong> ${tournament.participant_count || 0} | 
                       <strong>Status:</strong> <span class="badge ${tournament.status === 'computed' ? 'badge-success' : 'badge-secondary'}">${tournament.status || 'N/A'}</span></p>
                </div>
        `;

        // Tournament Players Tab Navigation
        html += `
            <div class="tabs">
                <nav class="tab-nav" style="margin-bottom: 20px;">
                    <button class="tab-nav-item active" data-tab="participants-list">Teilnehmerliste</button>
                    ${tournament.evaluations && tournament.evaluations.length > 0 ? 
                        '<button class="tab-nav-item" data-tab="dwz-changes">DWZ-Änderungen</button>' : ''}
                    ${tournament.games && tournament.games.length > 0 ? 
                        '<button class="tab-nav-item" data-tab="tournament-results">Spielergebnisse</button>' : ''}
                </nav>
        `;

        // Participants List Tab
        html += `
            <div id="participants-list" class="tab-content active">
                <h4>Teilnehmerliste (${tournament.participants.length} Spieler)</h4>
                <div class="table-container">
                    <table class="table table-striped">
                        <thead>
                            <tr>
                                <th>Nr.</th>
                                <th>Name</th>
                                <th>Verein</th>
                                <th>Start-DWZ</th>
                                <th>Geburtsjahr</th>
                                <th>Nation</th>
                                <th>FIDE-ID</th>
                                <th>Status</th>
                            </tr>
                        </thead>
                        <tbody>
        `;

        tournament.participants.forEach(participant => {
            const statusClass = participant.state === 2 ? 'badge-success' : 
                               participant.state === 1 ? 'badge-warning' : 'badge-error';
            const statusText = participant.state === 2 ? 'OK' : 
                              participant.state === 1 ? 'Unbekannt' : 'Gesperrt';
            
            html += `
                <tr>
                    <td><strong>${participant.no || 'N/A'}</strong></td>
                    <td><strong>${Utils.sanitizeHTML(participant.full_name || 'N/A')}</strong></td>
                    <td>${Utils.sanitizeHTML(participant.club?.name || 'N/A')}</td>
                    <td>${participant.rating?.use_rating ? 
                         `<span class="badge badge-primary">${participant.rating.use_rating}</span>` : 'N/A'}</td>
                    <td>${Utils.formatBirthYear(participant.birth_year)}</td>
                    <td>${Utils.sanitizeHTML(participant.nation || 'N/A')}</td>
                    <td>${participant.fide_id ? `<code>${participant.fide_id}</code>` : 'N/A'}</td>
                    <td><span class="badge ${statusClass}">${statusText}</span></td>
                </tr>
            `;
        });

        html += `
                        </tbody>
                    </table>
                </div>
            </div>
        `;

        // DWZ Changes Tab (if evaluations exist)
        if (tournament.evaluations && tournament.evaluations.length > 0) {
            html += `
                <div id="dwz-changes" class="tab-content">
                    <h4>DWZ-Änderungen (${tournament.evaluations.length} Bewertungen)</h4>
                    <p>Zeigt, wie sich die Deutsche Wertungszahl (DWZ) für jeden Spieler durch dieses Turnier geändert hat.</p>
                    <div class="table-container">
                        <table class="table table-striped">
                            <thead>
                                <tr>
                                    <th>Spieler</th>
                                    <th>DWZ Alt</th>
                                    <th>DWZ Neu</th>
                                    <th>Änderung</th>
                                    <th>Partien</th>
                                    <th>Punkte</th>
                                    <th>Leistung</th>
                                    <th>We</th>
                                </tr>
                            </thead>
                            <tbody>
            `;

            tournament.evaluations.forEach(evaluation => {
                const dwzChange = evaluation.dwz_new - evaluation.dwz_old;
                const changeClass = dwzChange > 0 ? 'text-success' : dwzChange < 0 ? 'text-danger' : '';
                const changeText = dwzChange > 0 ? `+${dwzChange}` : dwzChange.toString();
                
                html += `
                    <tr>
                        <td><strong>${Utils.sanitizeHTML(evaluation.player_name || 'N/A')}</strong></td>
                        <td><span class="badge badge-secondary">${evaluation.dwz_old || 'N/A'}</span></td>
                        <td><span class="badge badge-primary">${evaluation.dwz_new || 'N/A'}</span></td>
                        <td><strong class="${changeClass}">${changeText}</strong></td>
                        <td>${evaluation.games || 0}</td>
                        <td>${evaluation.points || 0}</td>
                        <td>${evaluation.achievement || 'N/A'}</td>
                        <td>${evaluation.we ? evaluation.we.toFixed(2) : 'N/A'}</td>
                    </tr>
                `;
            });

            html += `
                            </tbody>
                        </table>
                    </div>
                </div>
            `;
        }

        // Tournament Results Tab (if games exist)
        if (tournament.games && tournament.games.length > 0) {
            html += `
                <div id="tournament-results" class="tab-content">
                    <h4>Spielergebnisse (${tournament.games.length} Runden)</h4>
                    <p>Detaillierte Ergebnisse aller Spiele nach Runden sortiert.</p>
            `;

            tournament.games.forEach(round => {
                html += `
                    <div class="card" style="margin-bottom: 20px;">
                        <div class="card-header">
                            <h5>Runde ${round.round}</h5>
                            ${round.appointment ? `<p><strong>Termin:</strong> ${Utils.sanitizeHTML(round.appointment)}</p>` : ''}
                        </div>
                        <div class="table-container">
                            <table class="table table-striped">
                                <thead>
                                    <tr>
                                        <th>Brett</th>
                                        <th>Weiß</th>
                                        <th>Schwarz</th>
                                        <th>Ergebnis</th>
                                        <th>Punkte W</th>
                                        <th>Punkte S</th>
                                    </tr>
                                </thead>
                                <tbody>
                `;

                round.games.forEach(game => {
                    html += `
                        <tr>
                            <td><strong>${game.board}</strong></td>
                            <td>${Utils.sanitizeHTML(game.white?.name || 'N/A')} (${game.white?.no || 'N/A'})</td>
                            <td>${Utils.sanitizeHTML(game.black?.name || 'N/A')} (${game.black?.no || 'N/A'})</td>
                            <td><strong>${Utils.sanitizeHTML(game.result || 'N/A')}</strong></td>
                            <td>${game.white_points !== undefined ? game.white_points : 'N/A'}</td>
                            <td>${game.black_points !== undefined ? game.black_points : 'N/A'}</td>
                        </tr>
                    `;
                });

                html += `
                                </tbody>
                            </table>
                        </div>
                    </div>
                `;
            });

            html += `</div>`;
        }

        html += `
                </div>
            </div>
        `;

        container.innerHTML = html;

        // Initialize sub-tabs functionality
        this.initializeSubTabs(containerId);
    }

    // Initialize sub-tabs functionality for tournament players detail
    initializeSubTabs(containerId) {
        const container = document.getElementById(containerId);
        if (!container) return;

        const tabButtons = container.querySelectorAll('.tab-nav-item');
        const tabContents = container.querySelectorAll('.tab-content');

        tabButtons.forEach(button => {
            button.addEventListener('click', () => {
                // Remove active class from all buttons and tabs
                tabButtons.forEach(btn => btn.classList.remove('active'));
                tabContents.forEach(content => content.classList.remove('active'));

                // Add active class to clicked button and corresponding tab
                button.classList.add('active');
                const targetTab = container.querySelector(`#${button.dataset.tab}`);
                if (targetTab) {
                    targetTab.classList.add('active');
                }
            });
        });
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

    autoLoadTournamentDetails(tournamentId) {
        // Pre-fill the tournament lookup form with the tournament ID
        const tournamentForm = document.getElementById('tournament-lookup-form');
        if (tournamentForm) {
            const tournamentIdInput = tournamentForm.querySelector('input[name="tournament_id"]');
            if (tournamentIdInput) {
                tournamentIdInput.value = tournamentId;
            }
        }

        // Automatically trigger the tournament lookup
        this.loadSpecificTournament(tournamentId);
        
        // Switch to the tournament lookup tab if tabs exist
        const tournamentTab = document.querySelector('a[href="#tournament-lookup"]');
        if (tournamentTab) {
            tournamentTab.click();
        }
    }

    async loadSpecificTournament(tournamentId) {
        try {
            Utils.showLoading('tournament-lookup-results');
            
            const result = await api.getTournament(tournamentId);
            this.displayTournamentDetail('tournament-lookup-results', result.data || result);
            
            // Also show tournament players automatically
            const playersResult = await api.getTournamentPlayers(tournamentId);
            
            // Create a combined display showing both tournament details and players
            let combinedHtml = `
                <div class="tournament-detail-wrapper">
                    <h3>Turnier-Details</h3>
                    <div id="tournament-info-section"></div>
                    
                    <h3>Turnier-Spieler und Ergebnisse</h3>
                    <div id="tournament-players-section"></div>
                </div>
            `;
            
            document.getElementById('tournament-lookup-results').innerHTML = combinedHtml;
            
            this.displayTournamentDetail('tournament-info-section', result.data || result);
            this.displayTournamentPlayers('tournament-players-section', playersResult.data || playersResult);
            
        } catch (error) {
            Utils.showError('tournament-lookup-results', `Turnier-Details laden fehlgeschlagen: ${error.message}`);
        }
    }
}

// Initialize tournament manager when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    if (document.querySelector('.tournament-page')) {
        window.tournamentManager = new TournamentManager();
    }
});