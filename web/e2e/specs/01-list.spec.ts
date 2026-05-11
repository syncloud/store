import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'

test.describe('app list', () => {
  test('lists all apps from the store', async ({ page }, testInfo) => {
    await page.goto('/')

    await expect(page.getByTestId('brand')).toBeVisible()
    await expect(page.getByTestId('app-list')).toBeVisible()

    const cards = page.getByTestId('app-card-item')
    await expect(cards.first()).toBeVisible()

    const total = await cards.count()
    expect(total).toBeGreaterThanOrEqual(2)
    await expect(page.getByTestId('results-count')).toContainText(`${total} of ${total} apps`)
    const names = await cards.locator('[data-testid="app-name"]').allTextContents()
    expect(names).toEqual(expect.arrayContaining(['Test App 1', 'Test App 2']))

    await shoot(page, testInfo, 'app-list')
  })
})
