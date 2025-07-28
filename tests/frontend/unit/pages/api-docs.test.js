/**
 * @fileoverview Tests for the api-docs.html page
 * Tests documentation page structure and content
 */

const fs = require('fs');
const path = require('path');
const { JSDOM } = require('jsdom');

describe('API Documentation Page (api-docs.html)', () => {
  let dom, document, window;

  beforeEach(() => {
    // Load HTML file
    const htmlPath = path.resolve(__dirname, '../../../../demo/api-docs.html');
    const htmlContent = fs.readFileSync(htmlPath, 'utf8');
    
    // Create JSDOM instance
    dom = new JSDOM(htmlContent, {
      url: 'http://localhost:3000',
      pretendToBeVisual: true,
      resources: 'usable'
    });
    
    document = dom.window.document;
    window = dom.window;
  });

  afterEach(() => {
    dom.window.close();
  });

  describe('Page Structure', () => {
    test('should have correct page title', () => {
      expect(document.title).toBe('Portal64 API Demo - Documentation');
    });

    test('should have main header', () => {
      const header = document.querySelector('header h1');
      expect(header).toBeTruthy();
      expect(header.textContent).toBe('API Documentation');
    });

    test('should have back to dashboard link', () => {
      const backLink = document.querySelector('header a[href="index.html"]');
      expect(backLink).toBeTruthy();
      expect(backLink.textContent).toContain('Back to Dashboard');
    });
  });

  describe('Tab Navigation', () => {
    test('should have tab navigation buttons', () => {
      const tabButtons = document.querySelectorAll('.tab-nav-item');
      expect(tabButtons.length).toBeGreaterThan(0);
      
      const expectedTabs = ['overview', 'health', 'players-api', 'clubs-api', 'tournaments-api', 'swagger'];
      expectedTabs.forEach(tabId => {
        const tabButton = document.querySelector(`[data-tab="${tabId}"]`);
        expect(tabButton).toBeTruthy();
      });
    });

    test('should have corresponding tab content', () => {
      const expectedTabs = ['overview', 'health', 'players-api', 'clubs-api', 'tournaments-api', 'swagger'];
      expectedTabs.forEach(tabId => {
        const tabContent = document.getElementById(tabId);
        expect(tabContent).toBeTruthy();
        expect(tabContent.classList.contains('tab-content')).toBe(true);
      });
    });
  });

  describe('API Documentation Content', () => {
    test('should have base URL information', () => {
      const baseUrlCell = document.querySelector('td code');
      expect(baseUrlCell).toBeTruthy();
      expect(baseUrlCell.textContent).toBe('http://localhost:8080');
    });

    test('should have API version information', () => {
      const content = document.body.textContent;
      expect(content).toContain('v1');
    });

    test('should have pagination documentation', () => {
      const content = document.body.textContent;
      expect(content).toContain('Pagination');
      expect(content).toContain('limit');
      expect(content).toContain('offset');
    });

    test('should have players API documentation', () => {
      const playersSection = document.getElementById('players-api');
      expect(playersSection).toBeTruthy();
      expect(playersSection.textContent).toContain('/api/v1/players');
      expect(playersSection.textContent).toContain('query');
    });

    test('should have clubs API documentation', () => {
      const clubsSection = document.getElementById('clubs-api');
      expect(clubsSection).toBeTruthy();
      expect(clubsSection.textContent).toContain('/api/v1/clubs');
    });

    test('should have tournaments API documentation', () => {
      const tournamentsSection = document.getElementById('tournaments-api');
      expect(tournamentsSection).toBeTruthy();
      expect(tournamentsSection.textContent).toContain('/api/v1/tournaments');
    });
  });

  describe('Response Format Documentation', () => {
    test('should document JSON and CSV formats', () => {
      const content = document.body.textContent;
      expect(content).toContain('JSON');
      expect(content).toContain('CSV');
      expect(content).toContain('format=csv');
    });

    test('should have example responses', () => {
      const codeBlocks = document.querySelectorAll('pre, code');
      expect(codeBlocks.length).toBeGreaterThan(0);
    });
  });

  describe('Health Check Documentation', () => {
    test('should have health endpoint documentation', () => {
      const healthSection = document.getElementById('health');
      expect(healthSection).toBeTruthy();
      expect(healthSection.textContent).toContain('/health');
    });
  });

  describe('Parameter Documentation', () => {
    test('should have parameter tables', () => {
      const tables = document.querySelectorAll('table');
      expect(tables.length).toBeGreaterThan(0);
      
      // Check for common parameter documentation
      const content = document.body.textContent;
      expect(content).toContain('Parameter');
      expect(content).toContain('Type');
      expect(content).toContain('Required');
      expect(content).toContain('Description');
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
      
      const expectedFiles = ['main.js'];
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
