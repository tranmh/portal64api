// Club Profile functionality

class ClubProfileManager {
    constructor() {
        this.currentProfile = null;
        this.init();
    }

    init() {
        const profileForm = document.getElementById('club-profile-form');
        if (profileForm) {
            profileForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.loadClubProfile();
            });
        }

        // Check if club ID is in URL parameters
        const urlParams = new URLSearchParams(window.location.search);
        const clubId = urlParams.get('club_id');
        if (clubId) {
            document.getElementById('club-id').value = clubId;
            this.loadClubProfile();
        }
    }

    async loadClubProfile() {
        try {
            Utils.showLoading('club-profile-results');
            
            const form = document.getElementById('club-profile-form');
            const formData = Utils.getFormData(form);
            const clubId = formData.club_id;
            
            // Validate club ID
            if (!Utils.validateClubID(clubId)) {
                Utils.showError('club-profile-results', 'Ungültiges Vereins-ID Format. Erwartetes Format: alphanumerische Zeichen (z.B., C0101, B000E)');
                return;
            }
            
            const result = await api.getClubProfile(clubId);
            this.currentProfile = result.data || result;
            this.displayClubProfile('club-profile-results', this.currentProfile);
            
        } catch (error) {
            Utils.showError('club-profile-results', `Fehler beim Laden des Vereinsprofils: ${error.message}`);
        }
    }

    displayClubProfile(containerId, profile) {
        const container = document.getElementById(containerId);
        if (!container) return;

        if (!profile || !profile.club) {
            container.innerHTML = `
                <div class="alert alert-error">
                    <h4>Kein Vereinsprofil gefunden</h4>
                    <p>Das angegebene Vereinsprofil konnte nicht geladen werden.</p>
                </div>
            `;
            return;
        }

        const club = profile.club;
        
        let html = `
            <!-- Club Header -->
            <div class="card club-header">
                <div class="card-header">
                    <h2 class="card-title">${Utils.sanitizeHTML(club.name || 'Unbekannter Verein')}</h2>
                    <div class="club-badges">
                        <span class="badge badge-primary">VKZ: ${Utils.sanitizeHTML(club.id)}</span>
                        <span class="badge badge-secondary">Status: ${Utils.sanitizeHTML(club.status)}</span>
                    </div>
                </div>
                
                <div class="form-row">
                    <div class="club-basic-info">
                        <h4>Grundinformationen</h4>
                        <p><strong>Vollständiger Name:</strong> ${Utils.sanitizeHTML(club.name || 'N/A')}</p>
                        <p><strong>Kurzname:</strong> ${Utils.sanitizeHTML(club.short_name || 'N/A')}</p>
                        <p><strong>Region:</strong> ${Utils.sanitizeHTML(club.region || 'N/A')}</p>
                        <p><strong>Bezirk:</strong> ${Utils.sanitizeHTML(club.district || 'N/A')}</p>
                        <p><strong>Gründungsdatum:</strong> ${Utils.formatDate(club.founding_date)}</p>
                    </div>
                    
                    <div class="club-stats">
                        <h4>Vereinsstatistiken</h4>
                        <p><strong>Gesamtmitglieder:</strong> 
                            <span class="badge badge-info">${profile.player_count || 0}</span>
                        </p>
                        <p><strong>Aktive Spieler:</strong> 
                            <span class="badge badge-success">${profile.active_player_count || 0}</span>
                        </p>
                        <p><strong>Durchschnittliche DWZ:</strong> 
                            <span class="badge badge-secondary">${profile.rating_stats.average_dwz ? Math.round(profile.rating_stats.average_dwz) : 'N/A'}</span>
                        </p>
                        <p><strong>Median DWZ:</strong> 
                            <span class="badge badge-secondary">${profile.rating_stats.median_dwz ? Math.round(profile.rating_stats.median_dwz) : 'N/A'}</span>
                        </p>
                    </div>
                </div>
            </div>
        `;

        // Rating Statistics
        if (profile.rating_stats && Object.keys(profile.rating_stats.rating_distribution || {}).length > 0) {
            html += `
                <div class="card">
                    <div class="card-header">
                        <h3 class="card-title">Bewertungsstatistiken</h3>
                    </div>
                    
                    <div class="form-row">
                        <div>
                            <h4>DWZ Bereich</h4>
                            <p><strong>Höchste DWZ:</strong> <span class="badge badge-primary">${profile.rating_stats.highest_dwz || 'N/A'}</span></p>
                            <p><strong>Niedrigste DWZ:</strong> <span class="badge badge-secondary">${profile.rating_stats.lowest_dwz || 'N/A'}</span></p>
                            <p><strong>Spieler mit DWZ:</strong> <span class="badge badge-info">${profile.rating_stats.players_with_dwz || 0}</span></p>
                        </div>
                        
                        <div>
                            <h4>Bewertungsverteilung</h4>
                            ${this.createRatingDistribution(profile.rating_stats.rating_distribution)}
                        </div>
                    </div>
                </div>
            `;
        }

        // Players List
        if (profile.players && profile.players.length > 0) {
            html += `
                <div class="card">
                    <div class="card-header">
                        <h3 class="card-title">Vereinsspieler (Top ${profile.players.length})</h3>
                        <p>Sortiert nach DWZ-Bewertung (höchste zuerst)</p>
                    </div>
                    
                    <div class="table-container">
                        <table class="table table-striped">
                            <thead>
                                <tr>
                                    <th>Name</th>
                                    <th>Vorname</th>
                                    <th>DWZ</th>
                                    <th>Index</th>
                                    <th>Geburtsjahr</th>
                                    <th>Geschlecht</th>
                                    <th>Status</th>
                                </tr>
                            </thead>
                            <tbody>
            `;

            profile.players.forEach(player => {
                const birthYear = player.birth ? new Date(player.birth).getFullYear() : 'N/A';
                const dwzBadge = player.current_dwz > 0 ? 
                    `<span class="badge badge-primary">${player.current_dwz}</span>` : 
                    '<span class="badge badge-secondary">Unbewertet</span>';
                
                html += `
                    <tr>
                        <td><strong>${Utils.sanitizeHTML(player.name || 'N/A')}</strong></td>
                        <td>${Utils.sanitizeHTML(player.firstname || 'N/A')}</td>
                        <td>${dwzBadge}</td>
                        <td>${player.dwz_index || 'N/A'}</td>
                        <td>${birthYear}</td>
                        <td>${Utils.sanitizeHTML(player.gender || 'N/A')}</td>
                        <td>
                            <span class="badge ${player.status === 'active' ? 'badge-success' : 'badge-secondary'}">
                                ${Utils.sanitizeHTML(player.status || 'N/A')}
                            </span>
                        </td>
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

        // Actions
        html += `
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Kontakt & Links</h3>
                </div>
                
                <div class="form-row">
                    <div>
                        <h4>Vereinswebsite</h4>
                        ${profile.contact && profile.contact.website ? `
                            <p><a href="${profile.contact.website}" target="_blank" class="btn btn-primary">
                                🌐 Zur Vereinswebsite
                            </a></p>
                        ` : `
                            <div class="contact-info">
                                <p><strong>🌐 Vereinswebsite</strong></p>
                                <p class="text-muted">
                                    <em>Vereinswebsite derzeit nicht verfügbar</em><br>
                                    <small>Tipp: Suchen Sie online nach "${Utils.sanitizeHTML(club.name)}" oder kontaktieren Sie den Deutschen Schachbund für Kontaktdaten.</small>
                                </p>
                                <div class="contact-links">
                                    <a href="https://www.google.com/search?q=${encodeURIComponent(club.name + ' schachverein')}" target="_blank" class="external-link">
                                        🔍 Bei Google suchen
                                    </a>
                                </div>
                            </div>
                        `}
                        
                        ${profile.contact && profile.contact.email ? `
                            <p><strong>E-Mail:</strong> 
                                <a href="mailto:${profile.contact.email}">${Utils.sanitizeHTML(profile.contact.email)}</a>
                            </p>
                        ` : ''}
                        
                        ${profile.contact && profile.contact.phone ? `
                            <p><strong>Telefon:</strong> ${Utils.sanitizeHTML(profile.contact.phone)}</p>
                        ` : ''}
                    </div>
                    
                    <div>
                        ${profile.contact && profile.contact.meeting_place ? `
                            <h4>Spiellokal</h4>
                            <p><strong>Ort:</strong> ${Utils.sanitizeHTML(profile.contact.meeting_place)}</p>
                            ${profile.contact.meeting_time ? `
                                <p><strong>Spielzeit:</strong> ${Utils.sanitizeHTML(profile.contact.meeting_time)}</p>
                            ` : ''}
                        ` : `
                            <h4>Weitere Informationen</h4>
                            <div class="contact-info">
                                <p><strong>📋 Offizielle Vereinsdaten</strong></p>
                                <div class="contact-links">
                                    <a href="https://www.schachbund.de" target="_blank" class="external-link">
                                        🏛️ Deutscher Schachbund
                                    </a>
                                    <a href="https://www.schachverband-wuerttemberg.de" target="_blank" class="external-link">
                                        🏛️ Schachverband Württemberg
                                    </a>
                                    <small class="text-muted">
                                        Hier finden Sie offizielle Kontaktdaten und weitere Informationen über den Verein.
                                    </small>
                                </div>
                            </div>
                        `}
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Weitere Aktionen</h3>
                </div>
                
                <div class="form-row">
                    <div class="form-group">
                        <a href="players.html?club_id=${encodeURIComponent(club.id)}" class="btn btn-primary">
                            Alle Vereinsspieler anzeigen
                        </a>
                    </div>
                    <div class="form-group">
                        <a href="clubs.html" class="btn btn-secondary">
                            Andere Vereine suchen
                        </a>
                    </div>
                    <div class="form-group">
                        <button onclick="clubProfileManager.shareProfile('${club.id}')" class="btn btn-secondary">
                            Profil teilen
                        </button>
                    </div>
                </div>
            </div>
        `;

        container.innerHTML = html;
    }

    createRatingDistribution(distribution) {
        if (!distribution || Object.keys(distribution).length === 0) {
            return '<p>Keine Bewertungsverteilung verfügbar</p>';
        }

        let html = '<div class="rating-distribution">';
        
        // Sort by rating categories (highest first)
        const sortedEntries = Object.entries(distribution).sort((a, b) => {
            // Extract the rating value from category strings like "Expert (2200+)"
            const extractRating = (category) => {
                const match = category.match(/\((\d+)[-+]/);
                return match ? parseInt(match[1]) : 0;
            };
            return extractRating(b[0]) - extractRating(a[0]);
        });

        sortedEntries.forEach(([category, count]) => {
            const percentage = count > 0 ? ((count / Object.values(distribution).reduce((a, b) => a + b, 0)) * 100).toFixed(1) : 0;
            html += `
                <div class="rating-category">
                    <span class="category-name">${Utils.sanitizeHTML(category)}</span>
                    <span class="badge badge-info">${count} Spieler (${percentage}%)</span>
                </div>
            `;
        });

        html += '</div>';
        return html;
    }

    shareProfile(clubId) {
        const url = `${window.location.origin}${window.location.pathname}?club_id=${encodeURIComponent(clubId)}`;
        
        if (navigator.share) {
            navigator.share({
                title: `Vereinsprofil ${clubId}`,
                text: `Schau dir das Profil von Schachverein ${clubId} an`,
                url: url
            });
        } else {
            // Fallback: copy to clipboard
            navigator.clipboard.writeText(url).then(() => {
                alert('Profil-URL wurde in die Zwischenablage kopiert!');
            }).catch(() => {
                prompt('Kopiere diese URL:', url);
            });
        }
    }
}

// Initialize club profile manager when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    if (document.querySelector('.club-profile-page')) {
        window.clubProfileManager = new ClubProfileManager();
    }
});