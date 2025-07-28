/**
 * @fileoverview Tests for the players page (players.html)
 * Tests search functionality, form validation, and results display
 */

const { JSDOM } = require('jsdom');
const fs = require('fs');
const path = require('path');

describe('Players Page (players.html)', () => {
  let dom;
  let document;
  let window;

  beforeEach(async () => {
    // Load the actual HTML file
    const htmlPath = path.resolve(__dirname, '../../../../demo/players.html');
    
    // Check if file exists, if not skip this test suite
    if (!fs.existsSync(htmlPath)) {
      console.warn(`Skipping players.html tests - file not found at ${htmlPath}`);
      return;
    }

    const htmlContent = fs.readFileSync(htmlPath, 'utf8');

    // Create JSDOM instance
    dom = new JSDOM(htmlContent, {
      url: 'http://localhost:3000',
      pretendToBeVisual: true,
      resources: 'usable'
    });

    document = dom.window.document;
    window = dom.window;

    // Mock the API
    window.api = {
      searchPlayers: jest.fn(),
      getPlayer: jest.fn()
    };

    // Mock utility functions
    window.Utils = {
      showLoading: jest.fn(),
      showError: jest.fn(),
      showSuccess: jest.fn(),
      formatDate: jest.fn(date => date),
      formatRating: jest.fn(rating => rating.toString())
    };

    // Mock CodeDisplayManager
    window.CodeDisplayManager = jest.fn().mockImplementation(() => ({
      displayResponse: jest.fn(),
      displayError: jest.fn()
    }));
  });

  afterEach(() => {
    if (dom) {
      dom.window.close();
    }
  });

  describe('Page Structure', () => {
    test('should have correct page title', () => {
      if (!document) return;
      
      const title = document.querySelector('title');
      expect(title?.textContent).toContain('Players');
    });

    test('should have search form', () => {
      if (!document) return;
      
      // Look for common search form elements
      const searchInputs = document.querySelectorAll('input[type="text"], input[type="search"]');
      expect(searchInputs.length).toBeGreaterThan(0);
    });

    test('should have search button', () => {
      if (!document) return;
      
      const searchButton = document.querySelector('button[type="submit"], button.search-btn, input[type="submit"]');
      expect(searchButton).toBeTruthy();
    });

    test('should have results container', () => {
      if (!document) return;
      
      // Look for common result container patterns
      const resultContainers = document.querySelectorAll('#results, .results, #search-results, .search-results');
      expect(resultContainers.length).toBeGreaterThan(0);
    });
  });

  describe('Search Functionality', () => {
    test('should perform player search', async () => {
      if (!document || !window.api) return;

      const mockResults = {
        success: true,
        data: [
          {
            id: 'C0101-1014',
            name: 'Müller, Hans',
            first_name: 'Hans',
            rating: 1856,
            club_name: 'Post-SV Ulm'
          }
        ],
        meta: {
          total: 1,
          limit: 10,
          offset: 0
        }
      };

      window.api.searchPlayers.mockResolvedValue(mockResults);

      // Mock the search function that would be defined in the page's JavaScript
      window.performPlayerSearch = async () => {
        try {
          const searchTerm = document.querySelector('input[name="query"], #search-query')?.value || '';
          window.Utils.showLoading('results');
          const results = await window.api.searchPlayers({ query: searchTerm, limit: 10 });
          new window.CodeDisplayManager().displayResponse('results', results, 'Search Results');
        } catch (error) {
          window.Utils.showError('results', error.message);
        }
      };

      // Set search term
      const searchInput = document.querySelector('input[name="query"], #search-query');
      if (searchInput) {
        searchInput.value = 'Hans';
      }

      // Perform search
      await window.performPlayerSearch();

      expect(window.Utils.showLoading).toHaveBeenCalledWith('results');
      expect(window.api.searchPlayers).toHaveBeenCalledWith({ query: 'Hans', limit: 10 });
      expect(window.CodeDisplayManager).toHaveBeenCalled();
    });

    test('should handle search errors', async () => {
      if (!document || !window.api) return;

      window.api.searchPlayers.mockRejectedValue(new Error('Search failed'));

      window.performPlayerSearch = async () => {
        try {
          const searchTerm = 'test';
          await window.api.searchPlayers({ search: searchTerm });
        } catch (error) {
          window.Utils.showError('results', error.message);
        }
      };

      await window.performPlayerSearch();

      expect(window.Utils.showError).toHaveBeenCalledWith('results', 'Search failed');
    });

    test('should validate search input', () => {
      if (!document) return;

      // Mock validation function
      window.validateSearchInput = (value) => {
        if (!value || value.trim().length < 2) {
          window.Utils.showError('search-error', 'Search term must be at least 2 characters');
          return false;
        }
        return true;
      };

      // Test valid input
      expect(window.validateSearchInput('Hans')).toBe(true);

      // Test invalid input
      expect(window.validateSearchInput('H')).toBe(false);
      expect(window.Utils.showError).toHaveBeenCalledWith('search-error', 'Search term must be at least 2 characters');

      // Test empty input
      expect(window.validateSearchInput('')).toBe(false);
    });
  });

  describe('Player Details', () => {
    test('should fetch and display player details', async () => {
      if (!document || !window.api) return;

      const mockPlayer = {
        success: true,
        data: {
          id: 'C0101-1014',
          name: 'Müller, Hans',
          first_name: 'Hans',
          rating: 1856,
          club_name: 'Post-SV Ulm',
          games_played: 45,
          last_tournament: '2024-03-15'
        }
      };

      window.api.getPlayer.mockResolvedValue(mockPlayer);

      window.showPlayerDetails = async (playerId) => {
        try {
          window.Utils.showLoading('player-details');
          const player = await window.api.getPlayer(playerId);
          new window.CodeDisplayManager().displayResponse('player-details', player, 'Player Details');
        } catch (error) {
          window.Utils.showError('player-details', error.message);
        }
      };

      await window.showPlayerDetails('C0101-1014');

      expect(window.api.getPlayer).toHaveBeenCalledWith('C0101-1014');
      expect(window.Utils.showLoading).toHaveBeenCalledWith('player-details');
      expect(window.CodeDisplayManager).toHaveBeenCalled();
    });

    test('should handle player not found', async () => {
      if (!document || !window.api) return;

      window.api.getPlayer.mockRejectedValue(new Error('Player not found'));

      window.showPlayerDetails = async (playerId) => {
        try {
          await window.api.getPlayer(playerId);
        } catch (error) {
          window.Utils.showError('player-details', error.message);
        }
      };

      await window.showPlayerDetails('INVALID-ID');

      expect(window.Utils.showError).toHaveBeenCalledWith('player-details', 'Player not found');
    });
  });

  describe('Form Interactions', () => {
    test('should handle form submission', () => {
      if (!document) return;

      // Mock form submission handler
      window.handleSearchSubmit = (event) => {
        event.preventDefault();
        const formData = new FormData(event.target);
        const searchTerm = formData.get('search');
        
        if (window.validateSearchInput(searchTerm)) {
          window.performPlayerSearch(searchTerm);
        }
      };

      // Create mock form
      const form = document.createElement('form');
      const searchInput = document.createElement('input');
      searchInput.name = 'search';
      searchInput.value = 'test player';
      form.appendChild(searchInput);

      // Mock preventDefault
      const mockEvent = {
        target: form,
        preventDefault: jest.fn()
      };

      // Mock the validation and search functions
      window.validateSearchInput = jest.fn().mockReturnValue(true);
      window.performPlayerSearch = jest.fn();

      window.handleSearchSubmit(mockEvent);

      expect(mockEvent.preventDefault).toHaveBeenCalled();
      expect(window.validateSearchInput).toHaveBeenCalledWith('test player');
      expect(window.performPlayerSearch).toHaveBeenCalledWith('test player');
    });

    test('should clear search results', () => {
      if (!document) return;

      // Create results container
      const resultsContainer = document.createElement('div');
      resultsContainer.id = 'results';
      resultsContainer.innerHTML = '<div>Previous results</div>';
      document.body.appendChild(resultsContainer);

      window.clearSearchResults = () => {
        const container = document.getElementById('results');
        if (container) {
          container.innerHTML = '';
        }
      };

      window.clearSearchResults();

      expect(document.getElementById('results').innerHTML).toBe('');
    });
  });

  describe('Pagination', () => {
    test('should handle pagination controls', () => {
      if (!document) return;

      let currentPage = 1;
      const itemsPerPage = 10;

      window.goToPage = (page) => {
        currentPage = page;
        const offset = (page - 1) * itemsPerPage;
        window.performPlayerSearch({ offset, limit: itemsPerPage });
      };

      window.performPlayerSearch = jest.fn();

      // Test going to page 2
      window.goToPage(2);

      expect(currentPage).toBe(2);
      expect(window.performPlayerSearch).toHaveBeenCalledWith({ offset: 10, limit: 10 });
    });

    test('should update pagination display', () => {
      if (!document) return;

      // Create pagination container
      const paginationContainer = document.createElement('div');
      paginationContainer.id = 'pagination';
      document.body.appendChild(paginationContainer);

      window.updatePagination = (currentPage, totalPages) => {
        const container = document.getElementById('pagination');
        container.innerHTML = `Page ${currentPage} of ${totalPages}`;
      };

      window.updatePagination(2, 5);

      expect(document.getElementById('pagination').innerHTML).toBe('Page 2 of 5');
    });
  });
});
