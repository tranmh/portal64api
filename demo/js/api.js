// Main JavaScript functionality
class Portal64API {
    constructor() {
        this.baseURL = 'http://localhost:8080';
        this.defaultHeaders = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        };
    }

    // Generic API request method with timeout handling
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            method: 'GET',
            headers: this.defaultHeaders,
            ...options
        };

        try {
            // Create a timeout promise (increased to handle slow database)
            const timeoutPromise = new Promise((_, reject) => {
                setTimeout(() => reject(new Error('Request timeout - the server is taking longer than expected. Please try again.')), 45000); // 45 second timeout
            });

            // Race between fetch and timeout
            const response = await Promise.race([
                fetch(url, config),
                timeoutPromise
            ]);

            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || `HTTP error! status: ${response.status}`);
            }
            
            return data;
        } catch (error) {
            console.error('API Request failed:', error);
            throw error;
        }
    }

    // Health check
    async healthCheck() {
        return this.request('/health');
    }

    // Player endpoints
    async searchPlayers(params = {}) {
        // Ensure active defaults to true for demo interface
        if (!params.hasOwnProperty('active')) {
            params.active = true;
        }
        const queryString = new URLSearchParams(params).toString();
        return this.request(`/api/v1/players?${queryString}`);
    }

    async getPlayer(id, format = 'json') {
        return this.request(`/api/v1/players/${id}?format=${format}`);
    }

    async getPlayerRatingHistory(id, format = 'json') {
        return this.request(`/api/v1/players/${id}/rating-history?format=${format}`);
    }

    // Club endpoints
    async searchClubs(params = {}) {
        const queryString = new URLSearchParams(params).toString();
        return this.request(`/api/v1/clubs?${queryString}`);
    }

    async getAllClubs(format = 'json') {
        return this.request(`/api/v1/clubs/all?format=${format}`);
    }

    async getClub(id, format = 'json') {
        return this.request(`/api/v1/clubs/${id}?format=${format}`);
    }

    async getClubPlayers(id, params = {}) {
        // Ensure active defaults to true for demo interface
        if (!params.hasOwnProperty('active')) {
            params.active = true;
        }
        const queryString = new URLSearchParams(params).toString();
        return this.request(`/api/v1/clubs/${id}/players?${queryString}`);
    }

    // Tournament endpoints
    async searchTournaments(params = {}) {
        const queryString = new URLSearchParams(params).toString();
        return this.request(`/api/v1/tournaments?${queryString}`);
    }

    async getUpcomingTournaments(limit = 20, format = 'json') {
        return this.request(`/api/v1/tournaments/upcoming?limit=${limit}&format=${format}`);
    }

    async getRecentTournaments(days = 30, limit = 20, format = 'json') {
        return this.request(`/api/v1/tournaments/recent?days=${days}&limit=${limit}&format=${format}`);
    }

    async getTournamentsByDateRange(startDate, endDate, params = {}) {
        const allParams = { start_date: startDate, end_date: endDate, ...params };
        const queryString = new URLSearchParams(allParams).toString();
        return this.request(`/api/v1/tournaments/date-range?${queryString}`);
    }

    async getTournament(id, format = 'json') {
        return this.request(`/api/v1/tournaments/${id}?format=${format}`);
    }
}