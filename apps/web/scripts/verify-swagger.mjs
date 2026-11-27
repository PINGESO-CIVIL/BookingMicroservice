import { chromium } from '@playwright/test'
import { strict as assert } from 'node:assert'
import { readFileSync, mkdirSync, writeFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
const apiBase = process.env.MOCK_BASE_URL ?? 'http://127.0.0.1:5191'
const frontendBase = process.env.FRONTEND_BASE_URL ?? 'http://127.0.0.1:5190'
function isReservationsRead(response, base) {
  const url = new URL(response.url())
  return url.origin === new URL(base).origin && url.pathname === '/api/v1/reservas' &&
    response.request().method() === 'GET'
}
const redirect = await fetch(new URL('/api/docs', apiBase), { redirect: 'manual' })
assert.equal(redirect.status, 307)
assert.equal(redirect.headers.get('location'), '/api/docs/')
const url = new URL('/api/docs/', apiBase).href
const served = await fetch(`${url}openapi.yaml`)
assert.equal(served.status, 200)
assert.equal(await served.text(), readFileSync(new URL('../../../contracts/openapi.yaml', import.meta.url), 'utf8'), 'Swagger debe publicar el contrato actual.')
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
  const responsePromise = page.waitForResponse(r => isReservationsRead(r, apiBase))
  await operation.getByRole('button', { name: 'Execute', exact: true }).click()
  const response = await responsePromise
  assert.equal(response.status(), 200, `${response.url()} → ${await response.text()}`)
  assert.equal(response.headers()['x-mock-response'], 'true')
  assert.equal((await response.json()).items.length, 4)
  const destination = new URL('../../../reports/demo/', import.meta.url)
  mkdirSync(destination, { recursive: true })
  await page.screenshot({ path: fileURLToPath(new URL('swagger.png', destination)), fullPage: true })
  const frontend = await browser.newPage()
  const proxiedPromise = frontend.waitForResponse(r => isReservationsRead(r, frontendBase))
  await frontend.goto(new URL('/#/reservas', frontendBase).href)
  const proxied = await proxiedPromise
  assert.equal(proxied.status(), 200)
  assert.equal(proxied.headers()['x-mock-response'], 'true')
  await frontend.getByRole('article').first().waitFor()
  assert.equal(await frontend.getByRole('article').count(), 4)
  writeFileSync(new URL('swagger.json', destination), JSON.stringify({ checkedAt: new Date().toISOString(), operations: 14, servedContractMatches: true, tryItOutStatus: response.status(), mockResponse: true, frontendProxyStatus: proxied.status() }, null, 2) + '\n')
  console.log('Swagger ejecuta Try it out y React consulta el mismo backend por su proxy. Ambas peticiones devuelven HTTP 200.')
} finally { await browser.close() }
