// E2E Tests for Portal64 API Dashboard
// Tests the complete user journey from frontend UI to backend API

import { test, expect } from '@playwright/test';

test.describe('Portal64 Dashboard E2E Tests', () => {
  
  test('Dashboard loads and shows health status', async ({ page }) => {
    // Navigate to dashboard
    await page.goto('/demo/index.html');
    
    // Verify page loads correctly
    await expect(page).toHaveTitle(/Portal64 API/);
    await expect(page.locator('h1')).toContainText('Portal64 API Demo');
    
    // Test health check functionality (correct button text)
    await page.click('button:has-text("Refresh Health Check")');
    
    // Wait for health result container to be populated
    await expect(page.locator('#health-result')).not.toBeEmpty({ timeout: 5000 });
    
    // Verify health endpoint returns expected data
    const response = await page.request.get('/health');
    expect(response.ok()).toBeTruthy();
    const healthData = await response.json();
    expect(healthData).toHaveProperty('status');
  });

  test('Quick API test buttons work correctly', async ({ page }) => {
    await page.goto('/demo/index.html');
    
    // Test "Test Players Search" button (correct button text)
    await page.click('button:has-text("Test Players Search")');
    
    // Verify API result appears in the quick test container
    await expect(page.locator('#quick-test-result')).not.toBeEmpty({ timeout: 10000 });
    
    // Verify actual API response has correct structure
    const response = await page.request.get('/api/v1/players?limit=5');
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data).toHaveProperty('success');
    expect(data).toHaveProperty('data');
    expect(data.data).toHaveProperty('data');
    expect(Array.isArray(data.data.data)).toBeTruthy();
  });

  test('Navigation links work correctly', async ({ page }) => {
    await page.goto('/demo/index.html');
    
    // Test navigation to players page
    await page.click('a[href="players.html"]');
    await expect(page).toHaveURL(/players\.html$/);
    await expect(page.locator('h1')).toContainText('Players API'); // Actual title from HTML
    
    // Go back and test clubs page
    await page.goBack();
    await page.click('a[href="clubs.html"]');
    await expect(page).toHaveURL(/clubs\.html$/);
    await expect(page.locator('h1')).toContainText('Clubs API'); // Actual title from HTML
  });

  test('Error handling displays correctly', async ({ page }) => {
    await page.goto('/demo/index.html');
    
    // Mock a failing API call to test error handling
    await page.route('/health', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Service unavailable' })
      });
    });
    
    await page.click('button:has-text("Refresh Health Check")');
    
    // Verify error is displayed to user (use correct CSS class and check for actual error content)
    await expect(page.locator('.alert.alert-error')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.alert.alert-error')).toContainText('Error');
    await expect(page.locator('.alert.alert-error')).toContainText('Service unavailable');
  });
});
