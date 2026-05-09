import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'

test.describe('search filter', () => {
  test('filters down to a single match', async ({ page }, testInfo) => {
    await page.goto('/')

    const search = page.getByTestId('search')
    await expect(search).toBeVisible()
    await search.fill('App 1')

    const cards = page.getByTestId('app-card-item')
    await expect(cards).toHaveCount(1)
    await expect(cards.first().getByTestId('app-name')).toHaveText('Test App 1')

    await shoot(page, testInfo, 'search-app1')
  })

  test('shows empty state for unknown query', async ({ page }, testInfo) => {
    await page.goto('/')
    await page.getByTestId('search').fill('zzz-no-such-app-zzz')
    await expect(page.getByTestId('empty')).toBeVisible()
    await shoot(page, testInfo, 'search-empty')
  })

  test('clearing the query restores the full list', async ({ page }) => {
    await page.goto('/')
    const search = page.getByTestId('search')
    const cards = page.getByTestId('app-card-item')

    await expect(cards.first()).toBeVisible()
    const total = await cards.count()
    expect(total).toBeGreaterThanOrEqual(2)

    await search.fill('App 1')
    await expect(cards).toHaveCount(1)

    await search.fill('')
    await expect(cards).toHaveCount(total)
  })
})
