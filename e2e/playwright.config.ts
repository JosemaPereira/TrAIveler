/**
 * Playwright Test Configuration
 * 
 * Configures E2E test execution for TrAIveler application.
 * Tests run against local development environment (frontend + backend + database).
 * 
 * Key features:
 * - Parallel execution disabled by default (shared database state)
 * - Chromium-only for speed (CI runs all browsers)
 * - Retry on failure for flaky network/timing issues
 * - Screenshots and traces on failure for debugging
 * - Accessibility scanning via @axe-core/playwright
 */

import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  // Test directory structure
  testDir: './tests',
  
  // Fail fast on first failure during development
  fullyParallel: false,
  
  // Retry failed tests once (handles transient failures)
  retries: process.env.CI ? 2 : 1,
  
  // Single worker to avoid database conflicts
  // (Tests share local PostgreSQL instance)
  workers: 1,
  
  // Reporter configuration
  reporter: [
    ['list'], // Console output during run
    ['html', { outputFolder: 'playwright-report' }], // HTML report for debugging
  ],
  
  // Shared test configuration
  use: {
    // Base URL for frontend application
    baseURL: 'http://localhost:5173',
    
    // Capture screenshots on failure
    screenshot: 'only-on-failure',
    
    // Record trace for failed tests (video-like debugging)
    trace: 'retain-on-failure',
    
    // Default timeout for actions (click, fill, etc.)
    actionTimeout: 10000,
  },
  
  // Browser projects
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    
    // Firefox and WebKit disabled for local dev (faster feedback)
    // Enabled in CI for cross-browser validation
    // {
    //   name: 'firefox',
    //   use: { ...devices['Desktop Firefox'] },
    // },
    // {
    //   name: 'webkit',
    //   use: { ...devices['Desktop Safari'] },
    // },
  ],
  
  // Web server configuration
  // Assumes frontend is already running on port 5173
  // To auto-start: uncomment webServer block below
  // webServer: {
  //   command: 'cd ../frontend && npm run dev',
  //   url: 'http://localhost:5173',
  //   reuseExistingServer: !process.env.CI,
  //   timeout: 120000,
  // },
});
