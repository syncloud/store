import { defineConfig, devices } from '@playwright/test'
import * as path from 'node:path'

const baseURL = process.env.PLAYWRIGHT_BASE_URL
if (!baseURL) throw new Error('PLAYWRIGHT_BASE_URL is required')
const stub = process.env.PLAYWRIGHT_STUB === '1'
const artifactDir = path.resolve(__dirname, '../../artifact')

export default defineConfig({
  testDir: './specs',
  fullyParallel: false,
  workers: 1,
  retries: 0,
  maxFailures: 0,
  reporter: [
    ['list'],
    ['html', { open: 'never', outputFolder: `${artifactDir}/playwright/report` }],
  ],
  outputDir: `${artifactDir}/playwright/test-results`,
  timeout: 60_000,
  expect: { timeout: 10_000 },
  webServer: stub
    ? {
        command: 'npm --prefix .. run preview:stub',
        url: baseURL,
        reuseExistingServer: !process.env.CI,
        timeout: 120_000,
      }
    : undefined,
  use: {
    baseURL,
    ignoreHTTPSErrors: true,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'desktop',
      use: { ...devices['Desktop Chrome'], viewport: { width: 1440, height: 960 } },
    },
    {
      name: 'mobile',
      use: { ...devices['Desktop Chrome'], viewport: { width: 390, height: 844 } },
    },
  ],
  metadata: {
    baseURL,
    artifactDir,
  },
})
