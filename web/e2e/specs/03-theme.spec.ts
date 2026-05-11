import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'

test.describe('theme switcher', () => {
  test.beforeEach(async ({ context }) => {
    await context.clearCookies()
    await context.addInitScript(() => window.localStorage.clear())
  })

  test('defaults to light when prefers-color-scheme is light', async ({ page }, testInfo) => {
    await page.emulateMedia({ colorScheme: 'light' })
    await page.goto('/')

    await expect(page.getByTestId('theme-switcher')).toBeVisible()
    await expect(page.getByTestId('app-card-item').first()).toBeVisible()
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light')

    await shoot(page, testInfo, 'theme-light')
  })

  test('defaults to dark when prefers-color-scheme is dark', async ({ page }, testInfo) => {
    await page.emulateMedia({ colorScheme: 'dark' })
    await page.goto('/')

    await expect(page.getByTestId('theme-switcher')).toBeVisible()
    await expect(page.getByTestId('app-card-item').first()).toBeVisible()
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark')

    await shoot(page, testInfo, 'theme-dark-default')
  })

  test('toggles light to dark and back', async ({ page }, testInfo) => {
    await page.emulateMedia({ colorScheme: 'light' })
    await page.goto('/')

    const switcher = page.getByTestId('theme-switcher')
    await expect(switcher).toBeVisible()
    await expect(page.getByTestId('app-card-item').first()).toBeVisible()
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light')

    await switcher.click()
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark')
    await shoot(page, testInfo, 'theme-dark')

    await switcher.click()
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light')
    await shoot(page, testInfo, 'theme-light-after-toggle')
  })
})
