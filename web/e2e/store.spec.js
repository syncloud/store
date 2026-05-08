const { test, expect } = require('./fixtures')

test.describe('store web', () => {
  test('lists all apps', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByTestId('brand')).toBeVisible()
    await expect(page.getByTestId('app-list')).toBeVisible()

    const cards = page.getByTestId('app-card-item')
    await expect(cards.first()).toBeVisible()
    const total = await cards.count()
    expect(total).toBeGreaterThan(1)

    await expect(page.getByTestId('results-count')).toContainText(`${total} of ${total} apps`)
  })

  test('filters apps via search', async ({ page }) => {
    await page.goto('/')

    const search = page.getByTestId('search')
    await expect(search).toBeVisible()

    await search.fill('jelly')

    const cards = page.getByTestId('app-card-item')
    await expect(cards).toHaveCount(1)
    await expect(cards.first().getByTestId('app-name')).toHaveText('Jellyfin')

    await search.fill('')
    const after = await cards.count()
    expect(after).toBeGreaterThan(1)
  })

  test('shows empty state for no matches', async ({ page }) => {
    await page.goto('/')
    await page.getByTestId('search').fill('zzz-no-such-app-zzz')
    await expect(page.getByTestId('empty')).toBeVisible()
  })

  test('toggles dark and light theme', async ({ page }) => {
    await page.goto('/')

    const switcher = page.getByTestId('theme-switcher')
    await expect(switcher).toBeVisible()

    const initial = await page.evaluate(() => document.documentElement.getAttribute('data-theme'))
    await switcher.click()
    const next = await page.evaluate(() => document.documentElement.getAttribute('data-theme'))
    expect(next).not.toBe(initial)
    expect(['light', 'dark']).toContain(next)
  })
})
