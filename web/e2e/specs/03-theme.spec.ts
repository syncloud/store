import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'

test.describe('theme switcher', () => {
  test('toggles between light and dark', async ({ page }, testInfo) => {
    await page.emulateMedia({ colorScheme: 'light' })
    await page.goto('/')

    const switcher = page.getByTestId('theme-switcher')
    await expect(switcher).toBeVisible()
    await expect(page.getByTestId('app-card-item').first()).toBeVisible()

    await expect(page.locator('html')).toHaveAttribute('data-theme', 'light')
    await shoot(page, testInfo, 'theme-light')

    await switcher.click()
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark')
    await shoot(page, testInfo, 'theme-dark')
  })
})
