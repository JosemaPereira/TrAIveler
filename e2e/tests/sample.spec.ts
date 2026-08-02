/**
 * Placeholder verifying the Playwright setup itself works, not app behavior.
 * Replace with real E2E scenarios as features under e2e/tests/{accessibility,auth,collaboration} ship.
 */

import { test, expect } from '@playwright/test';

test.describe('Playwright Setup Verification', () => {
  test('should load example.com successfully', async ({ page }) => {
    await page.goto('https://example.com');

    await expect(page).toHaveTitle(/Example Domain/);

    const heading = page.getByRole('heading', { name: 'Example Domain' });
    await expect(heading).toBeVisible();
  });

  test('should take screenshot on example.com', async ({ page }) => {
    await page.goto('https://example.com');

    const screenshot = await page.screenshot();
    expect(screenshot).toBeTruthy();
    expect(screenshot.length).toBeGreaterThan(0);
  });
});
