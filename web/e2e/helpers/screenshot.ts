import { Page, TestInfo } from '@playwright/test'
import * as path from 'node:path'
import * as fs from 'node:fs'

const artifactRoot = process.env.PLAYWRIGHT_ARTIFACT_DIR ?? 'artifact'

export async function shoot(page: Page, testInfo: TestInfo, name: string) {
  await page.evaluate(() =>
    Promise.all(document.getAnimations().map(a => a.finished.catch(() => {})))
  )
  const view = testInfo.project.name
  const dir = path.join(artifactRoot, 'playwright', view, 'screenshot')
  fs.mkdirSync(dir, { recursive: true })
  await page.screenshot({ path: path.join(dir, `${name}-${view}.png`), fullPage: true })
  fs.writeFileSync(path.join(dir, `${name}-${view}.html`), await page.content())
}
