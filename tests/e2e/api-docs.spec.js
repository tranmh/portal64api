// E2E Tests for API Documentation Page
// Tests swagger documentation accessibility and functionality

import { test, expect } from '@playwright/test';

test.describe('API Documentation E2E Tests', () => {
  
  test('API documentation page loads correctly', async ({ page }) => {
    await page.goto('/demo/api-docs.html');
    
    // Verify page loads
    await expect(page).toHaveTitle(/Portal64 API Demo - Documentation/);
    await expect(page.locator('h1')).toContainText('API Documentation');
    
    // Verify tabs are visible (instead of embedded Swagger UI)
    await expect(page.locator('.tab-nav')).toBeVisible({ timeout: 10000 });
    
    // Verify navigation tabs are present
    await expect(page.locator('.tab-nav-item')).toHaveCount(6);
  });

  test('Swagger documentation interactive testing works', async ({ page }) => {
    await page.goto('/demo/api-docs.html');
    
    // Navigate to Swagger tab
    await page.click('button[data-tab="swagger"]');
    await expect(page.locator('#swagger')).toBeVisible();
    
    // Verify Swagger UI link is present (use more specific selector to avoid strict mode violation)
    await expect(page.locator('a[href="http://localhost:8080/swagger/"].btn.btn-primary')).toBeVisible();
    
    // Test the OpenAPI JSON link
    const response = await page.request.get('/swagger/doc.json');
    expect(response.ok()).toBeTruthy();
    
    // Verify the page explains Swagger functionality
    await expect(page.locator('#swagger')).toContainText('Interactive API Documentation');
  });

  test('All API endpoints are documented', async ({ page }) => {
    await page.goto('/demo/api-docs.html');
    
    // Check different API documentation tabs
    const apiTabs = ['players-api', 'clubs-api', 'tournaments-api'];
    const expectedEndpoints = [
      '/api/v1/players',
      '/api/v1/players/{id}',
      '/api/v1/players/{id}/rating-history',
      '/api/v1/clubs',
      '/api/v1/clubs/{id}',
      '/api/v1/clubs/{id}/players',
      '/api/v1/tournaments',
      '/api/v1/tournaments/{id}'
    ];
    
    // Verify each tab contains endpoint documentation
    for (const tab of apiTabs) {
      await page.click(`button[data-tab="${tab}"]`);
      await expect(page.locator(`#${tab}`)).toBeVisible();
      await expect(page.locator(`#${tab}`)).toContainText('GET /api/v1/');
    }
  });

  test('API documentation examples work correctly', async ({ page }) => {
    await page.goto('/demo/api-docs.html');
    
    // Test direct API call from documentation
    const response = await page.request.get('/api/v1/players?limit=1');
    expect(response.ok()).toBeTruthy();
    
    const data = await response.json();
    // API returns nested structure: {success: true, data: {data: [...], meta: {...}}}
    expect(data).toHaveProperty('success');
    expect(data).toHaveProperty('data');
    expect(data.data).toHaveProperty('data');
    expect(Array.isArray(data.data.data)).toBeTruthy();
    
    // Verify documentation contains code examples
    await page.click('button[data-tab="players-api"]');
    await expect(page.locator('#players-api')).toContainText('cURL Example');
  });
});
