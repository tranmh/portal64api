// E2E Tests for Club Search Functionality
// Tests complete user workflow: form input → API call → results display

import { test, expect } from '@playwright/test';

test.describe('Club Search E2E Tests', () => {
  
  test('Club search form submits and displays results', async ({ page }) => {
    await page.goto('/demo/clubs.html');
    
    // Verify page loads
    await expect(page.locator('h1')).toContainText('Vereine API');
    
    // Fill in search form (default tab should be "Search Clubs")
    await page.fill('input[name="query"]', 'Ulm');
    await page.selectOption('select[name="limit"]', '10');
    
    // Submit search
    await page.click('button[type="submit"]');
    
    // Wait for results to load
    await expect(page.locator('#club-search-results')).toBeVisible({ timeout: 10000 });
    
    // Verify API was called correctly
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/clubs') && 
      response.url().includes('query=Ulm')
    );
    expect(response.ok()).toBeTruthy();
    
    // Verify results are displayed (clubs are displayed in table rows)
    const resultRows = page.locator('#club-search-results table tbody tr');
    await expect(resultRows.first()).toBeVisible(); // At least 1 result row should be visible
  });

  test('Advanced search options work correctly', async ({ page }) => {
    await page.goto('/demo/clubs.html');
    
    // Expand advanced search
    await page.click('button.advanced-toggle');
    await expect(page.locator('.advanced-search')).toBeVisible();
    
    // Fill advanced options
    await page.fill('input[name="query"]', 'Test');
    await page.fill('input[name="offset"]', '5');
    await page.selectOption('select[name="sort_by"]', 'name');
    await page.selectOption('select[name="sort_order"]', 'desc');
    await page.selectOption('select[name="filter_by"]', 'region');
    await page.fill('input[name="filter_value"]', 'Bayern');
    
    // Submit search
    await page.click('button[type="submit"]');
    
    // Verify API was called with advanced parameters
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/clubs') && 
      response.url().includes('offset=5') &&
      response.url().includes('sort_by=name') &&
      response.url().includes('sort_order=desc')
    );
    expect(response.ok()).toBeTruthy();
  });

  test('All clubs tab functionality works', async ({ page }) => {
    await page.goto('/demo/clubs.html');
    
    // Switch to "All Clubs" tab
    await page.click('button[data-tab="all-clubs"]');
    await expect(page.locator('#all-clubs')).toBeVisible();
    
    // Select format and submit
    await page.selectOption('#all-clubs-format', 'json');
    await page.click('#all-clubs-form button[type="submit"]');
    
    // Wait for results
    await expect(page.locator('#all-clubs-results')).toBeVisible({ timeout: 15000 });
    
    // Verify API call
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/clubs') && 
      !response.url().includes('query=')
    );
    expect(response.ok()).toBeTruthy();
  });

  test('Club lookup by ID works correctly', async ({ page }) => {
    await page.goto('/demo/clubs.html');
    
    // Switch to "Club Lookup" tab
    await page.click('button[data-tab="club-lookup"]');
    await expect(page.locator('#club-lookup')).toBeVisible();
    
    // Fill in club ID
    await page.fill('input[name="club_id"]', 'C0101');
    
    // The format selector should be visible in this tab (not advanced search)
    await page.selectOption('#club-lookup-form select[name="format"]', 'json');
    
    // Submit lookup
    await page.click('#club-lookup-form button[type="submit"]');
    
    // Wait for results
    await expect(page.locator('#club-lookup-results')).toBeVisible({ timeout: 10000 });
    
    // Verify API call with specific club ID
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/clubs/C0101')
    );
    expect(response.ok()).toBeTruthy();
  });

  test('Club lookup validation works correctly', async ({ page }) => {
    await page.goto('/demo/clubs.html');
    
    // Switch to "Club Lookup" tab
    await page.click('button[data-tab="club-lookup"]');
    
    // Try to submit without club ID
    await page.click('#club-lookup-form button[type="submit"]');
    
    // Verify validation appears (HTML5 validation or custom validation)
    const isRequired = await page.locator('input[name="club_id"]').getAttribute('required');
    expect(isRequired).not.toBeNull();
  });

  test('CSV export functionality works', async ({ page }) => {
    await page.goto('/demo/clubs.html');
    
    // Perform a search first - expand advanced search to access format option
    await page.fill('input[name="query"]', 'Test');
    
    // Click advanced toggle to show format option
    await page.click('button.advanced-toggle');
    await expect(page.locator('.advanced-search.show')).toBeVisible();
    
    // Now select CSV format
    await page.selectOption('#search-format', 'csv');
    
    // Submit search with CSV format
    await page.click('button[type="submit"]');
    
    // Verify CSV API endpoint was called
    const response = await page.waitForResponse(response => 
      response.url().includes('/api/v1/clubs') && 
      response.url().includes('format=csv')
    );
    expect(response.ok()).toBeTruthy();
    expect(response.headers()['content-type']).toMatch(/text\/csv|application\/csv/);
  });

  test('Tab navigation works correctly', async ({ page }) => {
    await page.goto('/demo/clubs.html');
    
    // Test each tab
    const tabs = [
      { button: 'button[data-tab="search-clubs"]', content: '#search-clubs' },
      { button: 'button[data-tab="all-clubs"]', content: '#all-clubs' },
      { button: 'button[data-tab="club-lookup"]', content: '#club-lookup' }
    ];
    
    for (const tab of tabs) {
      await page.click(tab.button);
      await expect(page.locator(tab.content)).toBeVisible();
      
      // Verify tab button has active state
      await expect(page.locator(tab.button)).toHaveClass(/active/);
    }
  });

  test('Club detail modal works correctly', async ({ page }) => {
    await page.goto('/demo/clubs.html');
    
    // Perform a search to get results
    await page.fill('input[name="query"]', 'Test');
    await page.click('button[type="submit"]');
    
    // Wait for results and click on first club (if any)
    await expect(page.locator('#club-search-results')).toBeVisible();
    
    // Check if there are club results to click
    const clubResults = await page.locator('.club-result').count();
    if (clubResults > 0) {
      const firstClub = page.locator('.club-result').first();
      await firstClub.click();
      
      // Verify modal opens
      await expect(page.locator('#club-detail-modal')).toBeVisible();
      await expect(page.locator('.modal-title')).toContainText('Club Details');
      
      // Close modal
      await page.click('.modal-close');
      await expect(page.locator('#club-detail-modal')).not.toBeVisible();
    }
  });

  test('Error handling displays correctly', async ({ page }) => {
    await page.goto('/demo/clubs.html');
    
    // Mock a failing API call to test error handling
    await page.route('/api/v1/clubs*', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Service unavailable' })
      });
    });
    
    // Try to search
    await page.fill('input[name="query"]', 'Test');
    await page.click('button[type="submit"]');
    
    // Verify error is displayed to user
    await expect(page.locator('.alert.alert-error')).toBeVisible({ timeout: 5000 });
  });

  test('Back to dashboard navigation works', async ({ page }) => {
    await page.goto('/demo/clubs.html');
    
    // Click back to dashboard link
    await page.click('a[href="index.html"]');
    
    // Verify navigation to dashboard (URL should contain demo and either index.html or just be the demo root)
    await expect(page).toHaveURL(/\/demo\/?(?:index\.html)?$/);
    await expect(page.locator('h1')).toContainText('Portal64 API Demo');
  });
});