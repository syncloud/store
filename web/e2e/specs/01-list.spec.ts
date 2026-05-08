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
    expect(total).toBeGreaterThan(1)
    await expect(page.getByTestId('results-count')).toContainText(`${total} of ${total} apps`)

    await shoot(page, testInfo, 'app-list')
  })
})
