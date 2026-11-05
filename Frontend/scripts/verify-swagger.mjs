import { chromium } from '@playwright/test'
import { strict as assert } from 'node:assert'
import { readFileSync, mkdirSync, writeFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
const url = 'http://127.0.0.1:8080/docs/'
const served = await fetch(`${url}openapi.yaml`)
assert.equal(served.status, 200)
assert.equal(await served.text(), readFileSync(new URL('../../contracts/openapi/openapi.yaml', import.meta.url), 'utf8'), 'Swagger debe publicar el contrato actual.')
const browser = await chromium.launch({ channel: 'msedge', headless: true })
try {
  const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } })
  await page.goto(url)
  await page.locator('.opblock').first().waitFor()
  assert.equal(await page.locator('.opblock').count(), 14)
  await page.getByRole('button', { name: 'Authorize', exact: false }).first().click()
  const dialog = page.locator('.dialog-ux')
  await dialog.locator('input').fill('demo-user')
  await dialog.getByRole('button', { name: 'Apply credentials', exact: true }).click()
  await dialog.getByRole('button', { name: 'Close', exact: true }).click()
  const operation = page.locator('#operations-Reservas-listReservations')
  await operation.locator('.opblock-summary').click()
  await operation.getByRole('button', { name: 'Try it out', exact: true }).click()
  const responsePromise = page.waitForResponse(r => r.url().includes(':4010/api/v1/reservations') && r.request().method() === 'GET')
  await operation.getByRole('button', { name: 'Execute', exact: true }).click()
  const response = await responsePromise
  assert.equal(response.status(), 200, `${response.url()} → ${await response.text()}`)
  assert.equal(response.headers()['x-mock-response'], 'true')
  assert.equal((await response.json()).items.length, 4)
  const destination = new URL('../../reports/demo/', import.meta.url)
  mkdirSync(destination, { recursive: true })
  await page.screenshot({ path: fileURLToPath(new URL('swagger.png', destination)), fullPage: true })
  writeFileSync(new URL('swagger.json', destination), JSON.stringify({ checkedAt: new Date().toISOString(), operations: 14, servedContractMatches: true, tryItOutStatus: response.status(), mockResponse: true }, null, 2) + '\n')
  console.log('Swagger publica el contrato actual y ejecuta Try it out contra Go: HTTP 200.')
} finally { await browser.close() }
