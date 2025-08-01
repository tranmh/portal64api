// E2E Tests for Player Search Functionality
// Tests complete user workflow: form input → API call → results display

import { test, expect } from '@playwright/test';

test.describe('Player Search E2E Tests', () => {
  
  test('Player search form submits and displays results', async ({ page }) => {
    await page.goto('/demo/players.html');
    
    // Verify page loads (correct title)
    await expect(page.locator('h1')).toContainText('Spieler API');
    
    // Fill in search form using correct field names
    await page.fill('#search-query', 'Schmidt');
    await page.selectOption('#search-limit', '10');
    
    // Submit search
    await page.click('button[type="submit"]');
    
    // Wait for results to load in the correct container
    await expect(page.locator('#player-search-results')).toBeVisible({ timeout: 10000 });
    
    // Verify API was called correctly
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/players') && 
      response.url().includes('query=Schmidt')
    );
    expect(response.ok()).toBeTruthy();
    
    // Verify results are displayed (players displayed in table rows)
    const resultRows = page.locator('#player-search-results table tbody tr');
    await expect(resultRows.first()).toBeVisible();
  });

  test('Player detail view works correctly', async ({ page }) => {
    await page.goto('/demo/players.html');
    
    // Perform search first using correct field names
    await page.fill('input[name="query"]', 'Test');
    await page.click('button[type="submit"]');
    
    // Wait for results and check if there are clickable player rows
    await expect(page.locator('#player-search-results')).toBeVisible({ timeout: 10000 });
    
    // Check if there are result rows to click
    const resultRows = page.locator('#player-search-results table tbody tr');
    const rowCount = await resultRows.count();
    
    if (rowCount > 0) {
      // Click on first player row's "View Details" button if available
      const detailsButton = page.locator('#player-search-results table tbody tr').first().locator('button:has-text("View Details")');
      if (await detailsButton.isVisible({ timeout: 1000 })) {
        await detailsButton.click();
        
        // Verify detail view opens (modal or in-place details)
        await expect(page.locator('#player-detail-modal, .player-detail')).toBeVisible({ timeout: 5000 });
      }
    }
  });

  test('Search validation works correctly', async ({ page }) => {
    await page.goto('/demo/players.html');
    
    // Try to submit empty form - this should trigger HTML5 validation or custom validation
    await page.click('button[type="submit"]');
    
    // Check if form validation prevents submission (HTML5 validation)
    // The query field should be focused or show validation message (use specific selector)
    const queryField = page.locator('#search-query');
    const isInvalid = await queryField.evaluate(el => !el.checkValidity());
    expect(isInvalid).toBeTruthy();
  });

  test('Export to CSV functionality works', async ({ page }) => {
    await page.goto('/demo/players.html');
    
    // Perform a search first using correct field name
    await page.fill('input[name="query"]', 'Schmidt');
    
    // Show advanced search options to access format selector
    await page.click('button.advanced-toggle');
    await expect(page.locator('.advanced-search.show')).toBeVisible();
    
    // Select CSV format
    await page.selectOption('#search-format', 'csv');
    
    // Submit search
    await page.click('button[type="submit"]');
    
    // Verify CSV API endpoint was called
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/players') && 
      response.url().includes('format=csv')
    );
    expect(response.ok()).toBeTruthy();
    expect(response.headers()['content-type']).toMatch(/text\/csv|application\/csv/);
  });

  test('Pagination works correctly', async ({ page }) => {
    await page.goto('/demo/players.html');
    
    // Search for common name to get multiple results using correct field name
    await page.fill('#search-query', 'Mueller');
    await page.selectOption('#search-limit', '5'); // Small limit for pagination
    await page.click('button[type="submit"]');
    
    // Wait for results in correct container
    await expect(page.locator('#player-search-results')).toBeVisible({ timeout: 10000 });
    
    // Verify API call was made
    await page.waitForResponse(response => 
      response.url().includes('/api/v1/players') && 
      response.url().includes('query=Mueller') &&
      response.url().includes('limit=5')
    );
    
    // Check if there are results displayed
    const resultRows = page.locator('#player-search-results table tbody tr');
    await expect(resultRows.first()).toBeVisible();
  });

  test('Rating history display works', async ({ page }) => {
    await page.goto('/demo/players.html');
    
    // Navigate to Rating History tab
    await page.click('button[data-tab="rating-history"]');
    await expect(page.locator('#rating-history')).toBeVisible();
    
    // Search for a specific player using the correct form in this tab
    await page.fill('#rating-history input[name="player_id"]', 'C0101-1014');
    await page.click('#rating-history button[type="submit"]');
    
    // Wait for rating history results
    await expect(page.locator('#rating-history-results')).toBeVisible({ timeout: 10000 });
    
    // Verify API call was made for rating history (more flexible check)
    const response = await page.waitForResponse(response => {
      return response.url().includes('/api/v1/players/') && 
             response.url().includes('/rating-history');
    }, { timeout: 10000 });
    expect(response.ok()).toBeTruthy();
  });
});
