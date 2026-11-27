import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { setupServer } from 'msw/node'
import { handlers } from '../src/features/demo/mocks'
import { confirmed, resources } from '../src/api/fixtures'

const server = setupServer(...handlers)
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterAll(() => server.close())
const resource = resources[0]!.id
const api = 'http://agenda.test/api/v1'
async function call(path: string, method = 'GET', scenario = 'normal', body?: object) {
  return fetch(api + path, {
    method,
    headers: { 'X-Mock-Scenario': scenario },
    ...(body ? { body: JSON.stringify(body) } : {}),
  })
}
describe('mock de desarrollo', () => {
  it('ofrece los listados, el detalle y la disponibilidad de Agenda', async () => {
    for (const route of [
      '/reservas',
      '/admin/reservas',
      '/recursos',
      '/recursos/' + resource,
      '/reservas/' + confirmed.id,
    ]) {
      expect((await call(route)).status).toBe(200)
    }
    const filtered = await call(
      '/reservas?status=confirmed&resourceId=' +
        resource +
        '&from=2026-11-01&to=2026-12-01',
      'GET',
      'paginated',
    )
    expect((await filtered.json()).items.length).toBeLessThanOrEqual(1)
    expect(
      (await (await call('/recursos?enabled=false', 'GET', 'empty')).json()).items,
    ).toEqual([])
    expect((await call('/reservas/missing')).status).toBe(404)
    expect((await call('/recursos/missing')).status).toBe(404)
    const availability = '/recursos/' + resource + '/disponibilidades'
    expect((await call(availability)).status).toBe(400)
    const slot = await call(
      availability + '?from=2026-11-10T00:00:00Z&to=2026-11-11T00:00:00Z',
      'GET',
      'empty',
    )
    expect((await slot.json()).availableSlots).toEqual([])
    expect(
      (await (await call('/reservas/' + confirmed.id, 'GET', 'cancelled')).json()).status,
    ).toBe('cancelled')
  })
  it('simula las escrituras y los errores que usan los microcomponentes', async () => {
    for (const [route, method, status] of [
      ['/reservas', 'POST', 201],
      ['/reservas/' + confirmed.id, 'PATCH', 200],
      ['/reservas/' + confirmed.id + '/cancelaciones', 'POST', 200],
      ['/recursos', 'POST', 201],
      ['/recursos/' + resource, 'PATCH', 200],
      ['/recursos/' + resource + '/horarios', 'PUT', 200],
      ['/recursos/' + resource + '/bloqueos', 'POST', 201],
    ] as const)
      expect((await call(route, method, 'normal', { expectedVersion: 1 })).status).toBe(
        status,
      )
    expect((await call('/recursos/' + resource + '/bloqueos/any', 'DELETE')).status).toBe(
      204,
    )
    for (const [scenario, status] of [
      ['error', 503],
      ['forbidden', 403],
      ['unauthorized', 401],
      ['conflict', 409],
      ['version', 409],
      ['deadline', 422],
      ['policy', 422],
    ] as const) {
      expect((await call('/reservas', 'POST', scenario, {})).status).toBe(status)
    }
  })
})
