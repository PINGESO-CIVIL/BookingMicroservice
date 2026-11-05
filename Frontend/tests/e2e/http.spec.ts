import { expect, test } from '@playwright/test'
test('la pantalla consume una respuesta HTTP del mock Go', async ({ page }) => {
  const responsePromise = page.waitForResponse(r => r.url().includes('/api/v1/reservations') && r.request().method() === 'GET')
  await page.goto('/#/reservas')
  const response = await responsePromise
  expect(response.status()).toBe(200)
  expect(response.headers()['x-mock-response']).toBe('true')
  expect((await response.json()).items).toHaveLength(4)
  await expect(page.locator('.desktop-table tbody tr')).toHaveCount(4)
})
test('crear y cancelar usan HTTP y las escrituras no persisten', async ({ request }) => {
  const headers = { Authorization: 'Bearer demo-user', 'Idempotency-Key': 'http-demo-key' }
  const before = await request.get('/api/v1/reservations', { headers })
  const creation = await request.post('/api/v1/reservations', { headers, data: { resourceId: 'resource-001', startAt: '2026-11-10T21:00:00Z', endAt: '2026-11-10T22:00:00Z', attributes: {} } })
  expect(creation.status()).toBe(201)
  const cancellation = await request.post('/api/v1/reservations/reservation-001/cancel', { headers, data: { expectedVersion: 1 } })
  expect((await cancellation.json()).status).toBe('cancelled')
  const after = await request.get('/api/v1/reservations', { headers })
  expect(await after.json()).toEqual(await before.json())
})
