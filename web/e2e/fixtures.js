const base = require('@playwright/test')

const test = base.test.extend({})

test.afterEach(async ({ page }, testInfo) => {
  if (testInfo.status !== testInfo.expectedStatus) {
    await page.screenshot({
      path: testInfo.outputPath('failure-full-page.png'),
      fullPage: true
    })
  }
})

module.exports = {
  test,
  expect: base.expect
}
