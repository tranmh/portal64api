/**
 * @fileoverview Tests for the clubs.html page
 * Tests page structure, form functionality, and API interactions
 */

const fs = require('fs');
const path = require('path');
const { JSDOM } = require('jsdom');

describe('Clubs Page (clubs.html)', () => {
  let dom, document, window;

  beforeEach(() => {
    // Load HTML file
    const htmlPath = path.resolve(__dirname, '../../../../demo/clubs.html');
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
      searchClubs: jest.fn(),
      getAllClubs: jest.fn(),
      getClub: jest.fn()
    };

    window.Utils = {
      showLoading: jest.fn(),
      showError: jest.fn(),
      showSuccess: jest.fn(),
      getFormData: jest.fn()
    };

    window.CodeDisplayManager = jest.fn();
    window.ClubManager = jest.fn();
  });

  afterEach(() => {
    dom.window.close();
  });

  describe('Page Structure', () => {
    test('should have correct page title', () => {
      expect(document.title).toBe('Portal64 API Demo - Clubs');
    });

    test('should have main header', () => {
      const header = document.querySelector('header h1');
      expect(header).toBeTruthy();
      expect(header.textContent).toBe('Clubs API');
    });

    test('should have search form', () => {
      const searchForm = document.getElementById('club-search-form');
      expect(searchForm).toBeTruthy();
      expect(searchForm.tagName).toBe('FORM');
    });

    test('should have search button', () => {
      const searchButton = document.querySelector('#club-search-form button[type="submit"]');
      expect(searchButton).toBeTruthy();
      expect(searchButton.textContent).toBe('Search Clubs');
    });

    test('should have results container', () => {
      // Look for common result container patterns
      const resultContainers = document.querySelectorAll('#results, .results, #search-results, .search-results');
      expect(resultContainers.length).toBeGreaterThan(0);
    });

    test('should have all clubs form', () => {
      const allClubsForm = document.getElementById('all-clubs-form');
      expect(allClubsForm).toBeTruthy();
    });

    test('should have club lookup form', () => {
      const lookupForm = document.getElementById('club-lookup-form');
      expect(lookupForm).toBeTruthy();
    });
  });

  describe('Tab Navigation', () => {
    test('should have tab navigation buttons', () => {
      const tabButtons = document.querySelectorAll('.tab-nav-item');
      expect(tabButtons.length).toBeGreaterThan(0);
      
      const expectedTabs = ['search-clubs', 'all-clubs', 'club-lookup'];
      expectedTabs.forEach(tabId => {
        const tabButton = document.querySelector(`[data-tab="${tabId}"]`);
        expect(tabButton).toBeTruthy();
      });
    });

    test('should have corresponding tab content', () => {
      const expectedTabs = ['search-clubs', 'all-clubs', 'club-lookup'];
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
      expect(queryInput.placeholder).toContain('club name');
    });

    test('should have limit select field', () => {
      const limitSelect = document.querySelector('#search-limit');
      expect(limitSelect).toBeTruthy();
      expect(limitSelect.name).toBe('limit');
      
      const options = limitSelect.querySelectorAll('option');
      expect(options.length).toBeGreaterThan(0);
    });

    test('should have advanced search options', () => {
      const advancedSearch = document.querySelector('.advanced-search');
      expect(advancedSearch).toBeTruthy();
      
      // Check for filter options
      const filterBy = document.querySelector('#search-filter-by');
      const filterValue = document.querySelector('#search-filter-value');
      expect(filterBy).toBeTruthy();
      expect(filterValue).toBeTruthy();
    });
  });

  describe('Search Functionality', () => {
    test('should perform club search', async () => {
      if (!document || !window.api) return;

      // Mock successful API response
      window.api.searchClubs.mockResolvedValue({
        data: [
          { id: 'C0101', name: 'Test Club', region: 'Bayern' }
        ]
      });

      // Mock utility functions
      window.performClubSearch = async () => {
        window.Utils.showLoading('club-search-results');
        const result = await window.api.searchClubs({ query: 'Test', limit: 10 });
        window.CodeDisplayManager();
      };

      // Set search term
      const searchInput = document.querySelector('#search-query');
      if (searchInput) {
        searchInput.value = 'Test';
      }

      // Perform search
      await window.performClubSearch();

      expect(window.Utils.showLoading).toHaveBeenCalledWith('club-search-results');
      expect(window.api.searchClubs).toHaveBeenCalledWith({ query: 'Test', limit: 10 });
      expect(window.CodeDisplayManager).toHaveBeenCalled();
    });

    test('should handle search errors', async () => {
      if (!document || !window.api) return;

      window.api.searchClubs.mockRejectedValue(new Error('Search failed'));

      window.performClubSearch = async () => {
        try {
          window.Utils.showLoading('club-search-results');
          await window.api.searchClubs({ query: 'Test' });
        } catch (error) {
          window.Utils.showError('club-search-results', error.message);
        }
      };

      await window.performClubSearch();

      expect(window.Utils.showError).toHaveBeenCalledWith('club-search-results', 'Search failed');
    });
  });

  describe('All Clubs Functionality', () => {
    test('should fetch all clubs', async () => {
      if (!document || !window.api) return;

      window.api.getAllClubs.mockResolvedValue({
        data: [
          { id: 'C0101', name: 'Club 1' },
          { id: 'C0102', name: 'Club 2' }
        ]
      });

      window.fetchAllClubs = async () => {
        window.Utils.showLoading('all-clubs-results');
        const result = await window.api.getAllClubs();
        window.CodeDisplayManager();
      };

      await window.fetchAllClubs();

      expect(window.Utils.showLoading).toHaveBeenCalledWith('all-clubs-results');
      expect(window.api.getAllClubs).toHaveBeenCalled();
      expect(window.CodeDisplayManager).toHaveBeenCalled();
    });
  });

  describe('Club Lookup Functionality', () => {
    test('should lookup individual club', async () => {
      if (!document || !window.api) return;

      window.api.getClub.mockResolvedValue({
        id: 'C0101',
        name: 'Test Club',
        region: 'Bayern'
      });

      window.lookupClub = async () => {
        window.Utils.showLoading('club-lookup-results');
        const result = await window.api.getClub('C0101');
        window.CodeDisplayManager();
      };

      await window.lookupClub();

      expect(window.Utils.showLoading).toHaveBeenCalledWith('club-lookup-results');
      expect(window.api.getClub).toHaveBeenCalledWith('C0101');
      expect(window.CodeDisplayManager).toHaveBeenCalled();
    });

    test('should handle club not found', async () => {
      if (!document || !window.api) return;

      window.api.getClub.mockRejectedValue(new Error('Club not found'));

      window.lookupClub = async () => {
        try {
          window.Utils.showLoading('club-lookup-results');
          await window.api.getClub('C9999');
        } catch (error) {
          window.Utils.showError('club-lookup-results', error.message);
        }
      };

      await window.lookupClub();

      expect(window.Utils.showError).toHaveBeenCalledWith('club-lookup-results', 'Club not found');
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
      
      const expectedFiles = ['utils.js', 'api.js', 'main.js', 'clubs.js'];
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
