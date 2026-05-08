const { test, expect } = require('./fixtures')

test('mobile layout renders the app grid', async ({ page }) => {
  await page.goto('/')

  await expect(page.getByTestId('brand')).toBeVisible()
  await expect(page.getByTestId('search')).toBeVisible()

  const cards = page.getByTestId('app-card-item')
  await expect(cards.first()).toBeVisible()
  expect(await cards.count()).toBeGreaterThan(1)
})
