import { test, expect } from '@playwright/test'
import { shoot } from '../helpers/screenshot'

const stub = process.env.PLAYWRIGHT_STUB === '1'

test.describe('app card content', () => {
  test.skip(stub, 'asserts against real backend (testapp1/testapp2 from index-v2)')

  test('renders icon and summary for testapp1', async ({ page }, testInfo) => {
    await page.goto('/')

    const card = page.getByTestId('app-card-item').filter({ has: page.getByTestId('app-name').getByText('Test App 1', { exact: true }) })
    await expect(card).toHaveCount(1)

    await expect(card.getByTestId('app-summary')).toHaveText('First test application for the integration suite')

    const icon = card.getByTestId('app-icon')
    await expect(icon).toBeVisible()
    await expect(icon).toHaveAttribute('src', '/api/ui/v1/icons/testapp1.png')
    const naturalWidth = await icon.evaluate((el: HTMLImageElement) => el.naturalWidth)
    expect(naturalWidth).toBeGreaterThan(0)

    await shoot(page, testInfo, 'card-testapp1')
  })

  test('renders icon and summary for testapp2', async ({ page }) => {
    await page.goto('/')

    const card = page.getByTestId('app-card-item').filter({ has: page.getByTestId('app-name').getByText('Test App 2', { exact: true }) })
    await expect(card).toHaveCount(1)

    await expect(card.getByTestId('app-summary')).toHaveText('Second test application for the integration suite')

    const icon = card.getByTestId('app-icon')
    await expect(icon).toHaveAttribute('src', '/api/ui/v1/icons/testapp2.png')
    const naturalWidth = await icon.evaluate((el: HTMLImageElement) => el.naturalWidth)
    expect(naturalWidth).toBeGreaterThan(0)
  })
})
