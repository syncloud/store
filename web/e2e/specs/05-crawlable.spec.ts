import { test, expect } from '@playwright/test'

test.describe('crawlability', () => {
  test('serves robots.txt as text, not the app shell', async ({ request }) => {
    const res = await request.get('/robots.txt')

    expect(res.status()).toBe(200)
    expect(res.headers()['content-type']).toContain('text/plain')

    const body = await res.text()
    expect(body).not.toContain('<!doctype html>')
    expect(body).toContain('User-agent: *')
    expect(body).toContain('Sitemap: https://store.syncloud.org/sitemap.xml')
  })

  test('serves sitemap.xml as xml, not the app shell', async ({ request }) => {
    const res = await request.get('/sitemap.xml')

    expect(res.status()).toBe(200)
    expect(res.headers()['content-type']).toContain('xml')

    const body = await res.text()
    expect(body).not.toContain('<!doctype html>')
    expect(body).toContain('<loc>https://store.syncloud.org/</loc>')
  })

  test('still falls back to the app for an unknown path', async ({ request }) => {
    const res = await request.get('/not-a-real-path')

    expect(res.status()).toBe(200)
    expect(await res.text()).toContain('<div id="app">')
  })
})

test.describe('links back to the product pages', () => {
  test('a card is clickable exactly when it has a product page', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByTestId('app-list')).toBeVisible()

    const cards = page.getByTestId('app-card')
    const total = await cards.count()
    expect(total).toBeGreaterThan(0)

    for (let i = 0; i < total; i++) {
      const card = cards.nth(i)
      const href = await card.getAttribute('href')
      const more = await card.getByTestId('app-more').count()
      if (href === null) {
        expect(more, 'a card with no product page must not offer More').toBe(0)
        continue
      }
      expect(href).toMatch(/^https:\/\/syncloud\.org\/en\/[a-z-]+$/)
      expect(more, 'a clickable card must say where it goes').toBe(1)
    }
  })
})
