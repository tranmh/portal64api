// Club-specific functionality

class ClubManager {
    constructor() {
        this.currentResults = null;
        this.currentMeta = null;
        this.init();
    }

    init() {
        // Search clubs form handler
        const searchForm = document.getElementById('club-search-form');
        if (searchForm) {
            searchForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.searchClubs();
            });
        }

        // All clubs form handler
        const allClubsForm = document.getElementById('all-clubs-form');
        if (allClubsForm) {
            allClubsForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.getAllClubs();
            });
        }

        // Individual club lookup form
        const clubForm = document.getElementById('club-lookup-form');
        if (clubForm) {
            clubForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.getClub();
            });
        }
    }

    async searchClubs() {
        try {
            Utils.showLoading('club-search-results');
            
            const form = document.getElementById('club-search-form');
            const params = Utils.getFormData(form);
            
            const result = await api.searchClubs(params);
            // Fix: API returns {success: true, data: {data: [...], meta: {...}}}
            this.currentResults = result.data?.data || result.data || result;
            this.currentMeta = result.data?.meta || result.meta;
            
            this.displayClubResults('club-search-results', this.currentResults, this.currentMeta);
            
        } catch (error) {
            Utils.showError('club-search-results', `Search failed: ${error.message}`);
        }
    }

    async getAllClubs() {
        try {
            Utils.showLoading('all-clubs-results');
            
            const form = document.getElementById('all-clubs-form');
            const formData = Utils.getFormData(form);
            const format = formData.format || 'json';
            
            const result = await api.getAllClubs(format);
            
            if (format === 'json') {
                // Fix: API returns {success: true, data: [...]} for getAllClubs
                this.displayClubResults('all-clubs-results', result.data || result);
            } else {
                new CodeDisplayManager().displayResponse('all-clubs-results', result, 'All Clubs (CSV)');
            }
            
        } catch (error) {
            Utils.showError('all-clubs-results', `Failed to get all clubs: ${error.message}`);
        }
    }

    async getClub() {
        try {
            Utils.showLoading('club-lookup-results');
            
            const form = document.getElementById('club-lookup-form');
            const formData = Utils.getFormData(form);
            const clubId = formData.club_id;
            const format = formData.format || 'json';
            
            // Validate club ID
            if (!Utils.validateClubID(clubId)) {
                Utils.showError('club-lookup-results', 'Invalid club ID format. Expected format: alphanumeric characters (e.g., C0101, B000E, UNKNOWN)');
                return;
            }
            
            const result = await api.getClub(clubId, format);
            
            if (format === 'json') {
                // API returns {success: true, data: {...}}
                this.displayClubDetail('club-lookup-results', result.data || result);
            } else {
                new CodeDisplayManager().displayResponse('club-lookup-results', result, 'Club Data (CSV)');
            }
            
        } catch (error) {
            Utils.showError('club-lookup-results', `Club lookup failed: ${error.message}`);
        }
    }

    displayClubResults(containerId, clubs, meta) {
        const container = document.getElementById(containerId);
        if (!container) return;

        if (!clubs || clubs.length === 0) {
            container.innerHTML = `
                <div class="alert alert-info">
                    <h4>No Results</h4>
                    <p>No clubs found matching your search criteria.</p>
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
                            <th>Club ID</th>
                            <th>Name</th>
                            <th>Region</th>
                            <th>District</th>
                            <th>Members</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
        `;

        clubs.forEach(club => {
            html += `
                <tr>
                    <td><code>${Utils.sanitizeHTML(club.id || 'N/A')}</code></td>
                    <td><strong>${Utils.sanitizeHTML(club.name || 'N/A')}</strong></td>
                    <td>${Utils.sanitizeHTML(club.region || 'N/A')}</td>
                    <td>${Utils.sanitizeHTML(club.district || 'N/A')}</td>
                    <td>
                        ${club.member_count ? `<span class="badge badge-secondary">${club.member_count}</span>` : 'N/A'}
                    </td>
                    <td>
                        <button onclick="clubManager.viewClubDetail('${club.id}')" class="btn btn-small btn-secondary">
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
            case 'club-search-results':
                return 'clubManager.searchClubsWithOffset';
            default:
                return 'clubManager.searchClubsWithOffset';
        }
    }

    // Pagination handler for club search
    async searchClubsWithOffset(offset, containerId) {
        try {
            Utils.showLoading(containerId);
            
            const form = document.getElementById('club-search-form');
            const params = Utils.getFormData(form);
            params.offset = offset;
            
            const result = await api.searchClubs(params);
            this.currentResults = result.data?.data || result.data || result;
            this.currentMeta = result.data?.meta || result.meta;
            
            this.displayClubResults(containerId, this.currentResults, this.currentMeta);
            
        } catch (error) {
            Utils.showError(containerId, `Search failed: ${error.message}`);
        }
    }

    displayClubDetail(containerId, club) {
        const container = document.getElementById(containerId);
        if (!container) return;

        const html = `
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Club Details: ${Utils.sanitizeHTML(club.name || 'Unknown')}</h3>
                </div>
                <div class="form-row">
                    <div>
                        <h4>Basic Information</h4>
                        <p><strong>Club ID:</strong> <code>${Utils.sanitizeHTML(club.id || 'N/A')}</code></p>
                        <p><strong>Name:</strong> ${Utils.sanitizeHTML(club.name || 'N/A')}</p>
                        <p><strong>Short Name:</strong> ${Utils.sanitizeHTML(club.short_name || 'N/A')}</p>
                        <p><strong>Status:</strong> ${Utils.sanitizeHTML(club.status || 'N/A')}</p>
                    </div>
                    <div>
                        <h4>Location Information</h4>
                        <p><strong>Region:</strong> ${Utils.sanitizeHTML(club.region || 'N/A')}</p>
                        <p><strong>District:</strong> ${Utils.sanitizeHTML(club.district || 'N/A')}</p>
                        <p><strong>Founding Date:</strong> ${Utils.formatDate(club.founding_date)}</p>
                    </div>
                </div>
                <div class="form-row">
                    <div>
                        <h4>Statistics</h4>
                        <p><strong>Member Count:</strong> 
                            ${club.member_count ? `<span class="badge badge-primary">${club.member_count}</span>` : 'N/A'}
                        </p>
                        <p><strong>Average DWZ:</strong> 
                            ${club.average_dwz ? `<span class="badge badge-secondary">${Math.round(club.average_dwz)}</span>` : 'N/A'}
                        </p>
                    </div>
                    <div>
                        <h4>Actions</h4>
                        ${club.id ? `
                            <a href="players.html?club_id=${encodeURIComponent(club.id)}#club-players" class="btn btn-primary">
                                View Club Players
                            </a>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;

        container.innerHTML = html;
    }

    async viewClubDetail(clubId) {
        try {
            const modal = document.getElementById('club-detail-modal');
            const modalBody = modal.querySelector('.modal-body');
            
            modalBody.innerHTML = '<div class="loading show"><div class="spinner"></div><p>Loading club details...</p></div>';
            
            new ModalManager().openModal('club-detail-modal');
            
            const result = await api.getClub(clubId);
            modalBody.innerHTML = '<div id="club-detail-content"></div>';
            this.displayClubDetail('club-detail-content', result.data || result);
            
        } catch (error) {
            const modalBody = document.querySelector('#club-detail-modal .modal-body');
            modalBody.innerHTML = `<div class="alert alert-error"><p>Failed to load club details: ${error.message}</p></div>`;
        }
    }
}

// Initialize club manager when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    if (document.querySelector('.club-page')) {
        window.clubManager = new ClubManager();
    }
});