// E2E Tests for Tournament Search Functionality
// Tests complete user workflow: form input → API call → results display

import { test, expect } from '@playwright/test';

test.describe('Tournament Search E2E Tests', () => {
  
  test('Tournament search form submits and displays results', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Verify page loads
    await expect(page.locator('h1')).toContainText('Turniere API');
    
    // Fill in search form (default tab should be "Search Tournaments")
    await page.fill('input[name="query"]', 'Bundesliga');
    await page.selectOption('select[name="limit"]', '10');
    
    // Submit search
    await page.click('button[type="submit"]');
    
    // Wait for results to load
    await expect(page.locator('#tournament-search-results')).toBeVisible({ timeout: 10000 });
    
    // Verify API was called correctly
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/tournaments') && 
      response.url().includes('query=Bundesliga')
    );
    expect(response.ok()).toBeTruthy();
    
    // Verify results are displayed (tournaments displayed in table rows)
    const resultRows = page.locator('#tournament-search-results table tbody tr');
    await expect(resultRows.first()).toBeVisible();
  });

  test('Advanced tournament search options work correctly', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Expand advanced search
    await page.click('button.advanced-toggle');
    await expect(page.locator('.advanced-search.show')).toBeVisible();
    
    // Fill advanced options
    await page.fill('input[name="query"]', 'Championship');
    await page.fill('input[name="offset"]', '10');
    await page.selectOption('select[name="sort_by"]', 'startedOn');
    await page.selectOption('select[name="sort_order"]', 'asc');
    await page.selectOption('select[name="filter_by"]', 'year');
    await page.fill('input[name="filter_value"]', '2024');
    
    // Submit search
    await page.click('button[type="submit"]');
    
    // Verify API was called with advanced parameters
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/tournaments') && 
      response.url().includes('offset=10') &&
      response.url().includes('sort_by=startedOn') &&
      response.url().includes('sort_order=asc')
    );
    expect(response.ok()).toBeTruthy();
  });

  test('Upcoming tournaments tab functionality works', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Switch to "Upcoming" tab
    await page.click('button[data-tab="upcoming-tournaments"]');
    await expect(page.locator('#upcoming-tournaments')).toBeVisible();
    
    // Configure and submit
    await page.selectOption('#upcoming-limit', '20');
    await page.selectOption('#upcoming-format', 'json');
    await page.click('#upcoming-tournaments-form button[type="submit"]');
    
    // Wait for results
    await expect(page.locator('#upcoming-tournaments-results')).toBeVisible({ timeout: 10000 });
    
    // Verify API call
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/tournaments') && 
      response.url().includes('upcoming=true')
    );
    expect(response.ok()).toBeTruthy();
  });

  test('Recent tournaments tab functionality works', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Switch to "Recent" tab
    await page.click('button[data-tab="recent-tournaments"]');
    await expect(page.locator('#recent-tournaments')).toBeVisible();
    
    // Configure and submit
    await page.selectOption('#recent-days', '30');
    await page.selectOption('#recent-limit', '20');
    await page.selectOption('#recent-format', 'json');
    await page.click('#recent-tournaments-form button[type="submit"]');
    
    // Wait for results
    await expect(page.locator('#recent-tournaments-results')).toBeVisible({ timeout: 10000 });
    
    // Verify API call
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/tournaments') && 
      response.url().includes('days=30')
    );
    expect(response.ok()).toBeTruthy();
  });

  test('Date range tournaments tab functionality works', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Switch to "Date Range" tab
    await page.click('button[data-tab="date-range-tournaments"]');
    await expect(page.locator('#date-range-tournaments')).toBeVisible();
    
    // Fill date range
    await page.fill('#date-range-start', '2024-01-01');
    await page.fill('#date-range-end', '2024-12-31');
    
    // Submit search
    await page.click('#date-range-tournaments-form button[type="submit"]');
    
    // Wait for results
    await expect(page.locator('#date-range-tournaments-results')).toBeVisible({ timeout: 10000 });
    
    // Verify API call with date parameters
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/tournaments') && 
      response.url().includes('start_date=2024-01-01') &&
      response.url().includes('end_date=2024-12-31')
    );
    expect(response.ok()).toBeTruthy();
  });

  test('Date range advanced options work correctly', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Switch to date range tab
    await page.click('button[data-tab="date-range-tournaments"]');
    
    // Fill required dates
    await page.fill('#date-range-start', '2024-01-01');
    await page.fill('#date-range-end', '2024-06-30');
    
    // Expand advanced options
    await page.click('#date-range-tournaments .advanced-toggle');
    await expect(page.locator('#date-range-tournaments .advanced-search')).toBeVisible();
    
    // Fill advanced options
    await page.fill('#date-range-query', 'Open');
    await page.selectOption('#date-range-limit', '50');
    await page.fill('#date-range-offset', '5');
    await page.selectOption('#date-range-sort-by', 'name');
    
    // Submit search
    await page.click('#date-range-tournaments-form button[type="submit"]');
    
    // Verify API call includes advanced parameters
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/tournaments') && 
      response.url().includes('query=Open') &&
      response.url().includes('limit=50')
    );
    expect(response.ok()).toBeTruthy();
  });

  test('Tournament lookup by ID works correctly', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Switch to "Tournament Lookup" tab
    await page.click('button[data-tab="tournament-lookup"]');
    await expect(page.locator('#tournament-lookup')).toBeVisible();
    
    // Fill in tournament ID and use specific selectors
    await page.fill('#lookup-tournament-id', 'C529-K00-HT1');
    await page.selectOption('#lookup-format', 'json');
    
    // Submit lookup
    await page.click('#tournament-lookup-form button[type="submit"]');
    
    // Wait for results
    await expect(page.locator('#tournament-lookup-results')).toBeVisible({ timeout: 10000 });
    
    // Verify API call with specific tournament ID
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/tournaments/C529-K00-HT1')
    );
    expect(response.ok()).toBeTruthy();
  });

  test('Tournament lookup validation works correctly', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Switch to "Tournament Lookup" tab
    await page.click('button[data-tab="tournament-lookup"]');
    
    // Try to submit without tournament ID
    await page.click('#tournament-lookup-form button[type="submit"]');
    
    // Verify validation appears (HTML5 validation or custom validation)
    const isRequired = await page.locator('#lookup-tournament-id').getAttribute('required');
    expect(isRequired).not.toBeNull();
  });

  test('Date range validation works correctly', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Switch to date range tab
    await page.click('button[data-tab="date-range-tournaments"]');
    
    // Try to submit without dates
    await page.click('#date-range-tournaments-form button[type="submit"]');
    
    // Verify validation on required date fields
    const startRequired = await page.locator('#date-range-start').getAttribute('required');
    const endRequired = await page.locator('#date-range-end').getAttribute('required');
    expect(startRequired).not.toBeNull();
    expect(endRequired).not.toBeNull();
  });

  test('CSV export functionality works', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Perform a search with CSV format using specific selectors
    await page.fill('#search-query', 'Test');
    await page.selectOption('#search-format', 'csv');
    
    // Submit search with CSV format
    await page.click('button[type="submit"]');
    
    // Verify CSV API endpoint was called
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/tournaments') && 
      response.url().includes('format=csv')
    );
    expect(response.ok()).toBeTruthy();
    expect(response.headers()['content-type']).toMatch(/text\/csv|application\/csv/);
  });

  test('Tab navigation works correctly', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Test each tab
    const tabs = [
      { button: 'button[data-tab="search-tournaments"]', content: '#search-tournaments' },
      { button: 'button[data-tab="upcoming-tournaments"]', content: '#upcoming-tournaments' },
      { button: 'button[data-tab="recent-tournaments"]', content: '#recent-tournaments' },
      { button: 'button[data-tab="date-range-tournaments"]', content: '#date-range-tournaments' },
      { button: 'button[data-tab="tournament-lookup"]', content: '#tournament-lookup' }
    ];
    
    for (const tab of tabs) {
      await page.click(tab.button);
      await expect(page.locator(tab.content)).toBeVisible();
      
      // Verify tab button has active state
      await expect(page.locator(tab.button)).toHaveClass(/active/);
    }
  });

  test('Tournament detail modal works correctly', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Perform a search to get results
    await page.fill('input[name="query"]', 'Test');
    await page.click('button[type="submit"]');
    
    // Wait for results and click on first tournament (if any)
    await expect(page.locator('#tournament-search-results')).toBeVisible();
    
    // Check if there are tournament results to click
    const tournamentResults = await page.locator('.tournament-result').count();
    if (tournamentResults > 0) {
      const firstTournament = page.locator('.tournament-result').first();
      await firstTournament.click();
      
      // Verify modal opens
      await expect(page.locator('#tournament-detail-modal')).toBeVisible();
      await expect(page.locator('.modal-title')).toContainText('Tournament Details');
      
      // Close modal
      await page.click('.modal-close');
      await expect(page.locator('#tournament-detail-modal')).not.toBeVisible();
    }
  });

  test('Error handling displays correctly', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Mock a failing API call to test error handling
    await page.route('/api/v1/tournaments*', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Service unavailable' })
      });
    });
    
    // Try to search
    await page.fill('input[name="query"]', 'Test');
    await page.click('button[type="submit"]');
    
    // Verify error is displayed to user (use correct error class)
    await expect(page.locator('.alert.alert-error, .error, .form-error')).toBeVisible({ timeout: 5000 });
  });

  test('Back to dashboard navigation works', async ({ page }) => {
    await page.goto('/demo/tournaments.html');
    
    // Click back to dashboard link
    await page.click('a[href="index.html"]');
    
    // Verify navigation to dashboard (use correct expected URL)
    await expect(page).toHaveURL(/\/demo\/$/);
    await expect(page.locator('h1')).toContainText('Portal64 API Demo');
  });
});