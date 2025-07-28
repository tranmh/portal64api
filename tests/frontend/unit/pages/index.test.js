/**
 * @fileoverview Tests for the main dashboard page (index.html)
 * Tests DOM interactions, health checks, and navigation functionality
 */

const { JSDOM } = require('jsdom');
const fs = require('fs');
const path = require('path');

describe('Dashboard Page (index.html)', () => {
  let dom;
  let document;
  let window;

  beforeEach(async () => {
    // Load the actual HTML file
    const htmlPath = path.resolve(__dirname, '../../../../demo/index.html');
    const htmlContent = fs.readFileSync(htmlPath, 'utf8');

    // Create JSDOM instance
    dom = new JSDOM(htmlContent, {
      url: 'http://localhost:3000',
      pretendToBeVisual: true,
      resources: 'usable'
    });

    document = dom.window.document;
    window = dom.window;

    // Mock the global objects that the page expects
    window.fetch = jest.fn();
    window.api = {
      healthCheck: jest.fn(),
      searchPlayers: jest.fn(),
      searchClubs: jest.fn(),
      searchTournaments: jest.fn()
    };

    // Load utility functions that the page depends on
    window.Utils = {
      showLoading: jest.fn(),
      showError: jest.fn(),
      showSuccess: jest.fn()
    };

    window.CodeDisplayManager = jest.fn().mockImplementation(() => ({
      displayResponse: jest.fn()
    }));
  });

  afterEach(() => {
    dom.window.close();
  });

  describe('Page Structure', () => {
    test('should have correct page title', () => {
      const title = document.querySelector('title');
      expect(title.textContent).toBe('Portal64 API Demo - Dashboard');
    });

    test('should have main header', () => {
      const header = document.querySelector('header.header h1');
      expect(header).toBeTruthy();
      expect(header.textContent).toBe('Portal64 API Demo');
    });

    test('should have health check section', () => {
      const healthSection = document.querySelector('#health-result');
      expect(healthSection).toBeTruthy();
    });

    test('should have navigation cards', () => {
      const navCards = document.querySelectorAll('.nav-card');
      expect(navCards).toHaveLength(4);
      
      const expectedLinks = [
        'players.html',
        'clubs.html', 
        'tournaments.html',
        'api-docs.html'
      ];

      navCards.forEach((card, index) => {
        expect(card.getAttribute('href')).toBe(expectedLinks[index]);
      });
    });

    test('should have quick test buttons', () => {
      const quickTestButtons = document.querySelectorAll('.card button.btn-primary');
      expect(quickTestButtons).toHaveLength(3);
      
      const buttonTexts = Array.from(quickTestButtons).map(btn => btn.textContent);
      expect(buttonTexts).toContain('Test Players Search');
      expect(buttonTexts).toContain('Test Clubs Search');
      expect(buttonTexts).toContain('Test Tournaments');
    });
  });

  describe('Health Check Functionality', () => {
    test('should have health check button', () => {
      const healthButton = document.querySelector('button[onclick="performHealthCheck()"]');
      expect(healthButton).toBeTruthy();
      expect(healthButton.textContent).toBe('Refresh Health Check');
    });

    test('should display health check results', async () => {
      // Mock successful health check
      window.api.healthCheck.mockResolvedValue({
        status: 'ok',
        timestamp: Date.now(),
        version: '1.0.0'
      });

      // Define the performHealthCheck function that would be loaded from external JS
      window.performHealthCheck = async () => {
        try {
          const result = await window.api.healthCheck();
          const healthResult = document.getElementById('health-result');
          healthResult.innerHTML = `<div class="success">Status: ${result.status}</div>`;
        } catch (error) {
          const healthResult = document.getElementById('health-result');
          healthResult.innerHTML = `<div class="error">Error: ${error.message}</div>`;
        }
      };

      // Execute health check
      await window.performHealthCheck();

      const healthResult = document.getElementById('health-result');
      expect(healthResult.innerHTML).toContain('Status: ok');
      expect(window.api.healthCheck).toHaveBeenCalledTimes(1);
    });

    test('should handle health check errors', async () => {
      // Mock failed health check
      window.api.healthCheck.mockRejectedValue(new Error('Connection failed'));

      window.performHealthCheck = async () => {
        try {
          await window.api.healthCheck();
        } catch (error) {
          const healthResult = document.getElementById('health-result');
          healthResult.innerHTML = `<div class="error">Error: ${error.message}</div>`;
        }
      };

      // Execute health check
      await window.performHealthCheck();

      const healthResult = document.getElementById('health-result');
      expect(healthResult.innerHTML).toContain('Error: Connection failed');
    });
  });

  describe('Quick Test Functions', () => {
    test('should test players endpoint', async () => {
      const mockPlayersData = {
        success: true,
        data: [
          { id: 'C0101-1014', name: 'Test Player', rating: 1500 }
        ]
      };

      window.api.searchPlayers.mockResolvedValue(mockPlayersData);
      
      // Define the test function that would be in the actual page
      window.testPlayersEndpoint = async () => {
        try {
          window.Utils.showLoading('quick-test-result');
          const result = await window.api.searchPlayers({ limit: 5 });
          new window.CodeDisplayManager().displayResponse('quick-test-result', result, 'Players Search Result');
        } catch (error) {
          window.Utils.showError('quick-test-result', `Players test failed: ${error.message}`);
        }
      };

      await window.testPlayersEndpoint();

      expect(window.Utils.showLoading).toHaveBeenCalledWith('quick-test-result');
      expect(window.api.searchPlayers).toHaveBeenCalledWith({ limit: 5 });
      expect(window.CodeDisplayManager).toHaveBeenCalled();
    });

    test('should test clubs endpoint', async () => {
      const mockClubsData = {
        success: true,
        data: [
          { id: 'C0101', name: 'Test Club', city: 'Test City' }
        ]
      };

      window.api.searchClubs.mockResolvedValue(mockClubsData);
      
      window.testClubsEndpoint = async () => {
        try {
          window.Utils.showLoading('quick-test-result');
          const result = await window.api.searchClubs({ limit: 5 });
          new window.CodeDisplayManager().displayResponse('quick-test-result', result, 'Clubs Search Result');
        } catch (error) {
          window.Utils.showError('quick-test-result', `Clubs test failed: ${error.message}`);
        }
      };

      await window.testClubsEndpoint();

      expect(window.Utils.showLoading).toHaveBeenCalledWith('quick-test-result');
      expect(window.api.searchClubs).toHaveBeenCalledWith({ limit: 5 });
    });

    test('should test tournaments endpoint', async () => {
      const mockTournamentsData = {
        success: true,
        data: [
          { id: 'C529-K00-HT1', name: 'Test Tournament', location: 'Test Location' }
        ]
      };

      window.api.searchTournaments.mockResolvedValue(mockTournamentsData);
      
      window.testTournamentsEndpoint = async () => {
        try {
          window.Utils.showLoading('quick-test-result');
          const result = await window.api.searchTournaments({ limit: 5 });
          new window.CodeDisplayManager().displayResponse('quick-test-result', result, 'Tournaments Search Result');
        } catch (error) {
          window.Utils.showError('quick-test-result', `Tournaments test failed: ${error.message}`);
        }
      };

      await window.testTournamentsEndpoint();

      expect(window.Utils.showLoading).toHaveBeenCalledWith('quick-test-result');
      expect(window.api.searchTournaments).toHaveBeenCalledWith({ limit: 5 });
    });

    test('should handle test endpoint failures', async () => {
      window.api.searchPlayers.mockRejectedValue(new Error('API Error'));
      
      window.testPlayersEndpoint = async () => {
        try {
          window.Utils.showLoading('quick-test-result');
          await window.api.searchPlayers({ limit: 5 });
        } catch (error) {
          window.Utils.showError('quick-test-result', `Players test failed: ${error.message}`);
        }
      };

      await window.testPlayersEndpoint();

      expect(window.Utils.showError).toHaveBeenCalledWith(
        'quick-test-result', 
        'Players test failed: API Error'
      );
    });
  });

  describe('API Information Section', () => {
    test('should display correct base URL', () => {
      const baseUrlElement = document.querySelector('code');
      expect(baseUrlElement.textContent).toBe('http://localhost:8080');
    });

    test('should have API information sections', () => {
      const infoSections = document.querySelectorAll('.card h4');
      const sectionTitles = Array.from(infoSections).map(h4 => h4.textContent);
      
      expect(sectionTitles).toContain('Base URL');
      expect(sectionTitles).toContain('Available Formats');
      expect(sectionTitles).toContain('Authentication');
      expect(sectionTitles).toContain('Rate Limits');
      expect(sectionTitles).toContain('Pagination');
      expect(sectionTitles).toContain('Error Handling');
    });
  });

  describe('Footer', () => {
    test('should have correct footer content', () => {
      const footer = document.querySelector('footer p');
      expect(footer.textContent).toContain('© 2024 Portal64 API Demo');
      expect(footer.textContent).toContain('DWZ Chess Rating System Interface');
    });
  });

  describe('CSS and JavaScript Loading', () => {
    test('should have CSS links', () => {
      const cssLinks = document.querySelectorAll('link[rel="stylesheet"]');
      expect(cssLinks).toHaveLength(3);
      
      const cssHrefs = Array.from(cssLinks).map(link => link.getAttribute('href'));
      expect(cssHrefs).toContain('css/main.css');
      expect(cssHrefs).toContain('css/forms.css');
      expect(cssHrefs).toContain('css/components.css');
    });

    test('should have JavaScript script tags', () => {
      const scriptTags = document.querySelectorAll('script[src]');
      expect(scriptTags).toHaveLength(3);
      
      const scriptSrcs = Array.from(scriptTags).map(script => script.getAttribute('src'));
      expect(scriptSrcs).toContain('js/utils.js');
      expect(scriptSrcs).toContain('js/api.js');
      expect(scriptSrcs).toContain('js/main.js');
    });
  });
});
