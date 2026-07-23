import { defineConfig } from '@playwright/test';

// Runs against the real Go binary serving the embedded frontend, not the
// Vite dev server, so the test covers what actually ships.
export default defineConfig({
  testDir: './e2e',
  timeout: 90_000,
  expect: { timeout: 15_000 },
  fullyParallel: false,
  workers: 1,
  reporter: [['list']],
  use: {
    baseURL: process.env.BASE_URL ?? 'http://localhost:8080',
    trace: 'retain-on-failure',
  },
});
