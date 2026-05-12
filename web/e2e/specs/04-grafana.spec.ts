import { test } from '@playwright/test'
import { shoot } from '../helpers/screenshot'

const STORE = process.env.PLAYWRIGHT_BASE_URL!
const GRAFANA = 'http://grafana:3000'

test.describe('grafana popularity dashboard', () => {
  test.skip(
    !STORE.includes('api.store.test'),
    'requires the Drone vm + grafana services'
  )

  test('renders panels with seeded popularity data', async ({ page, request }, testInfo) => {
    const mobile = testInfo.project.name === 'mobile'
    await page.setViewportSize({ width: mobile ? 390 : 1440, height: mobile ? 2600 : 1400 })

    const refresh = async (snap: string, snapId: string, device: string) => {
      await request.post(`${STORE}/v2/snaps/refresh`, {
        headers: {
          'Content-Type': 'application/json',
          'Syncloud-Architecture': 'amd64',
          'Syncloud-Device-Id': device,
        },
        data: {
          actions: [
            { action: 'refresh', 'instance-key': 'k', name: snap, 'snap-id': snapId, channel: 'stable' },
          ],
        },
      })
    }

    for (let i = 0; i < 8; i++) await refresh('testapp1', 'testapp1.1', `e2e-app1-${i}`)
    for (let i = 0; i < 3; i++) await refresh('testapp2', 'testapp2.1', `e2e-app2-${i}`)

    // wait one VM scrape interval (5s) + slack so the latest popularity is in VM
    await page.waitForTimeout(8000)

    await page.goto(
      `${GRAFANA}/d/popularity/store-popularity?orgId=1&from=now-5m&to=now`,
      { waitUntil: 'networkidle' }
    )

    await page.waitForSelector('[data-testid="data-testid Panel header Unique devices"]', { timeout: 30_000 })
    await page.waitForSelector('[data-testid="data-testid VizLegend series testapp1"]', { timeout: 30_000 })
    await page.waitForSelector('[data-testid="data-testid Bar gauge value"]', { timeout: 30_000 })
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3_000)

    await shoot(page, testInfo, 'grafana-popularity', { fullPage: false })
  })
})
