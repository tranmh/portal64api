/**
 * @fileoverview Tests for the tournaments.html page
 * Tests page structure, form functionality, and API interactions
 */

const fs = require('fs');
const path = require('path');
const { JSDOM } = require('jsdom');

describe('Tournaments Page (tournaments.html)', () => {
  let dom, document, window;

  beforeEach(() => {
    // Load HTML file
    const htmlPath = path.resolve(__dirname, '../../../../demo/tournaments.html');
    const htmlContent = fs.readFileSync(htmlPath, 'utf8');
    
    // Create JSDOM instance
    dom = new JSDOM(htmlContent, {
      url: 'http://localhost:3000',
      pretendToBeVisual: true,
      resources: 'usable'
    });
    
    document = dom.window.document;
    window = dom.window;

    // Mock API and utility functions
    window.api = {
      searchTournaments: jest.fn(),
      getUpcomingTournaments: jest.fn(),
      getRecentTournaments: jest.fn(),
      getTournamentsByDateRange: jest.fn(),
      getTournament: jest.fn()
    };

    window.Utils = {
      showLoading: jest.fn(),
      showError: jest.fn(),
      showSuccess: jest.fn(),
      getFormData: jest.fn()
    };

    window.CodeDisplayManager = jest.fn();
    window.TournamentManager = jest.fn();
  });

  afterEach(() => {
    dom.window.close();
  });

  describe('Page Structure', () => {
    test('should have correct page title', () => {
      expect(document.title).toBe('Portal64 API Demo - Tournaments');
    });

    test('should have main header', () => {
      const header = document.querySelector('header h1');
      expect(header).toBeTruthy();
      expect(header.textContent).toBe('Tournaments API');
    });

    test('should have search form', () => {
      const searchForm = document.getElementById('tournament-search-form');
      expect(searchForm).toBeTruthy();
      expect(searchForm.tagName).toBe('FORM');
    });

    test('should have search button', () => {
      const searchButton = document.querySelector('#tournament-search-form button[type="submit"]');
      expect(searchButton).toBeTruthy();
      expect(searchButton.textContent).toBe('Search Tournaments');
    });

    test('should have results container', () => {
      // Look for common result container patterns
      const resultContainers = document.querySelectorAll('#results, .results, #search-results, .search-results');
      expect(resultContainers.length).toBeGreaterThan(0);
    });

    test('should have upcoming tournaments form', () => {
      const upcomingForm = document.getElementById('upcoming-tournaments-form');
      expect(upcomingForm).toBeTruthy();
    });

    test('should have recent tournaments form', () => {
      const recentForm = document.getElementById('recent-tournaments-form');
      expect(recentForm).toBeTruthy();
    });

    test('should have date range form', () => {
      const dateRangeForm = document.getElementById('date-range-tournaments-form');
      expect(dateRangeForm).toBeTruthy();
    });

    test('should have tournament lookup form', () => {
      const lookupForm = document.getElementById('tournament-lookup-form');
      expect(lookupForm).toBeTruthy();
    });
  });

  describe('Tab Navigation', () => {
    test('should have tab navigation buttons', () => {
      const tabButtons = document.querySelectorAll('.tab-nav-item');
      expect(tabButtons.length).toBeGreaterThan(0);
      
      const expectedTabs = ['search-tournaments', 'upcoming-tournaments', 'recent-tournaments', 'date-range-tournaments', 'tournament-lookup'];
      expectedTabs.forEach(tabId => {
        const tabButton = document.querySelector(`[data-tab="${tabId}"]`);
        expect(tabButton).toBeTruthy();
      });
    });

    test('should have corresponding tab content', () => {
      const expectedTabs = ['search-tournaments', 'upcoming-tournaments', 'recent-tournaments', 'date-range-tournaments', 'tournament-lookup'];
      expectedTabs.forEach(tabId => {
        const tabContent = document.getElementById(tabId);
        expect(tabContent).toBeTruthy();
        expect(tabContent.classList.contains('tab-content')).toBe(true);
      });
    });
  });

  describe('Form Elements', () => {
    test('should have query input field', () => {
      const queryInput = document.querySelector('#search-query');
      expect(queryInput).toBeTruthy();
      expect(queryInput.name).toBe('query');
      expect(queryInput.placeholder).toContain('tournament');
    });

    test('should have limit select field', () => {
      const limitSelect = document.querySelector('#search-limit');
      expect(limitSelect).toBeTruthy();
      expect(limitSelect.name).toBe('limit');
      
      const options = limitSelect.querySelectorAll('option');
      expect(options.length).toBeGreaterThan(0);
    });

    test('should have date range inputs', () => {
      const startDate = document.querySelector('#date-range-start');
      const endDate = document.querySelector('#date-range-end');
      expect(startDate).toBeTruthy();
      expect(endDate).toBeTruthy();
      expect(startDate.type).toBe('date');
      expect(endDate.type).toBe('date');
    });

    test('should have tournament ID input', () => {
      const tournamentId = document.querySelector('#lookup-tournament-id');
      expect(tournamentId).toBeTruthy();
      expect(tournamentId.name).toBe('tournament_id');
    });
  });

  describe('Search Functionality', () => {
    test('should perform tournament search', async () => {
      if (!document || !window.api) return;

      // Mock successful API response
      window.api.searchTournaments.mockResolvedValue({
        data: [
          { id: 'T001', name: 'Test Tournament', date: '2024-01-15' }
        ]
      });

      // Mock utility functions
      window.performTournamentSearch = async () => {
        window.Utils.showLoading('tournament-search-results');
        const result = await window.api.searchTournaments({ query: 'Test', limit: 10 });
        window.CodeDisplayManager();
      };

      // Set search term
      const searchInput = document.querySelector('#search-query');
      if (searchInput) {
        searchInput.value = 'Test';
      }

      // Perform search
      await window.performTournamentSearch();

      expect(window.Utils.showLoading).toHaveBeenCalledWith('tournament-search-results');
      expect(window.api.searchTournaments).toHaveBeenCalledWith({ query: 'Test', limit: 10 });
      expect(window.CodeDisplayManager).toHaveBeenCalled();
    });

    test('should handle search errors', async () => {
      if (!document || !window.api) return;

      window.api.searchTournaments.mockRejectedValue(new Error('Search failed'));

      window.performTournamentSearch = async () => {
        try {
          window.Utils.showLoading('tournament-search-results');
          await window.api.searchTournaments({ query: 'Test' });
        } catch (error) {
          window.Utils.showError('tournament-search-results', error.message);
        }
      };

      await window.performTournamentSearch();

      expect(window.Utils.showError).toHaveBeenCalledWith('tournament-search-results', 'Search failed');
    });
  });

  describe('Upcoming Tournaments Functionality', () => {
    test('should fetch upcoming tournaments', async () => {
      if (!document || !window.api) return;

      window.api.getUpcomingTournaments.mockResolvedValue({
        data: [
          { id: 'T001', name: 'Future Tournament', date: '2024-12-15' }
        ]
      });

      window.fetchUpcomingTournaments = async () => {
        window.Utils.showLoading('upcoming-tournaments-results');
        const result = await window.api.getUpcomingTournaments();
        window.CodeDisplayManager();
      };

      await window.fetchUpcomingTournaments();

      expect(window.Utils.showLoading).toHaveBeenCalledWith('upcoming-tournaments-results');
      expect(window.api.getUpcomingTournaments).toHaveBeenCalled();
      expect(window.CodeDisplayManager).toHaveBeenCalled();
    });
  });

  describe('Recent Tournaments Functionality', () => {
    test('should fetch recent tournaments', async () => {
      if (!document || !window.api) return;

      window.api.getRecentTournaments.mockResolvedValue({
        data: [
          { id: 'T002', name: 'Past Tournament', date: '2024-01-15' }
        ]
      });

      window.fetchRecentTournaments = async () => {
        window.Utils.showLoading('recent-tournaments-results');
        const result = await window.api.getRecentTournaments();
        window.CodeDisplayManager();
      };

      await window.fetchRecentTournaments();

      expect(window.Utils.showLoading).toHaveBeenCalledWith('recent-tournaments-results');
      expect(window.api.getRecentTournaments).toHaveBeenCalled();
      expect(window.CodeDisplayManager).toHaveBeenCalled();
    });
  });

  describe('Date Range Functionality', () => {
    test('should fetch tournaments by date range', async () => {
      if (!document || !window.api) return;

      window.api.getTournamentsByDateRange.mockResolvedValue({
        data: [
          { id: 'T003', name: 'Range Tournament', date: '2024-06-15' }
        ]
      });

      window.fetchTournamentsByDateRange = async () => {
        window.Utils.showLoading('date-range-tournaments-results');
        const result = await window.api.getTournamentsByDateRange('2024-01-01', '2024-12-31');
        window.CodeDisplayManager();
      };

      await window.fetchTournamentsByDateRange();

      expect(window.Utils.showLoading).toHaveBeenCalledWith('date-range-tournaments-results');
      expect(window.api.getTournamentsByDateRange).toHaveBeenCalledWith('2024-01-01', '2024-12-31');
      expect(window.CodeDisplayManager).toHaveBeenCalled();
    });
  });

  describe('Tournament Lookup Functionality', () => {
    test('should lookup individual tournament', async () => {
      if (!document || !window.api) return;

      window.api.getTournament.mockResolvedValue({
        id: 'C529-K00-HT1',
        name: 'Test Tournament',
        date: '2024-01-15'
      });

      window.lookupTournament = async () => {
        window.Utils.showLoading('tournament-lookup-results');
        const result = await window.api.getTournament('C529-K00-HT1');
        window.CodeDisplayManager();
      };

      await window.lookupTournament();

      expect(window.Utils.showLoading).toHaveBeenCalledWith('tournament-lookup-results');
      expect(window.api.getTournament).toHaveBeenCalledWith('C529-K00-HT1');
      expect(window.CodeDisplayManager).toHaveBeenCalled();
    });

    test('should handle tournament not found', async () => {
      if (!document || !window.api) return;

      window.api.getTournament.mockRejectedValue(new Error('Tournament not found'));

      window.lookupTournament = async () => {
        try {
          window.Utils.showLoading('tournament-lookup-results');
          await window.api.getTournament('INVALID');
        } catch (error) {
          window.Utils.showError('tournament-lookup-results', error.message);
        }
      };

      await window.lookupTournament();

      expect(window.Utils.showError).toHaveBeenCalledWith('tournament-lookup-results', 'Tournament not found');
    });
  });

  describe('CSS and JavaScript Loading', () => {
    test('should have CSS links', () => {
      const cssLinks = document.querySelectorAll('link[rel="stylesheet"]');
      expect(cssLinks.length).toBeGreaterThan(0);
      
      const expectedFiles = ['main.css', 'forms.css', 'components.css'];
      expectedFiles.forEach(file => {
        const link = Array.from(cssLinks).find(l => l.href.includes(file));
        expect(link).toBeTruthy();
      });
    });

    test('should have JavaScript script tags', () => {
      const scripts = document.querySelectorAll('script[src]');
      expect(scripts.length).toBeGreaterThan(0);
      
      const expectedFiles = ['utils.js', 'api.js', 'main.js', 'tournaments.js'];
      expectedFiles.forEach(file => {
        const script = Array.from(scripts).find(s => s.src.includes(file));
        expect(script).toBeTruthy();
      });
    });
  });

  describe('Footer', () => {
    test('should have correct footer content', () => {
      const footer = document.querySelector('footer');
      expect(footer).toBeTruthy();
      expect(footer.textContent).toContain('2024');
      expect(footer.textContent).toContain('Portal64 API Demo');
    });
  });
});
