import { Page, TestInfo } from '@playwright/test'
import * as path from 'node:path'
import * as fs from 'node:fs'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const artifactRoot = path.resolve(here, '../../../artifact')

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
