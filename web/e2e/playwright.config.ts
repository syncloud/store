import { defineConfig, devices } from '@playwright/test'

const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? 'http://127.0.0.1:4173'
const artifactDir = process.env.PLAYWRIGHT_ARTIFACT_DIR ?? 'artifact'

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
  webServer: process.env.PLAYWRIGHT_BASE_URL
    ? undefined
    : {
        command: 'npm --prefix .. run preview:stub',
        url: 'http://127.0.0.1:4173',
        reuseExistingServer: !process.env.CI,
        timeout: 120_000,
      },
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
