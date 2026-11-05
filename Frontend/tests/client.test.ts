import { afterEach, expect, it, vi } from 'vitest'
import { api, ApiError } from '../src/api/client'
afterEach(() => vi.unstubAllGlobals())
it('envía creación por HTTP con clave de idempotencia y sin responsable editable', async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ id: 'reservation-001' }), { status: 201 }))
  vi.stubGlobal('fetch', fetchMock)
  const body = { resourceId: 'resource-001', startAt: '2026-11-10T21:00:00Z', endAt: '2026-11-10T22:00:00Z', attributes: {} }
  await api.createReservation(body, 'same-request-key', 'success')
  const [path, options] = fetchMock.mock.calls[0]!
  expect(path).toBe('/api/v1/reservations')
  expect(options.method).toBe('POST')
  expect(options.headers['Idempotency-Key']).toBe('same-request-key')
  expect(options.headers.Authorization).toBe('Bearer demo-user')
  expect(JSON.parse(options.body)).toEqual(body)
})
it('conserva códigos de error HTTP para recuperar conflictos', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { code: 'RESERVATION_VERSION_CONFLICT', message: 'Recarga la reserva.', requestId: 'request-1' } }), { status: 409 })))
  await expect(api.cancelReservation('reservation-001', { expectedVersion: 1 }, 'version')).rejects.toMatchObject({ code: 'RESERVATION_VERSION_CONFLICT', requestId: 'request-1' })
})
it('maneja DELETE 204 sin intentar leer un cuerpo JSON', async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 })); vi.stubGlobal('fetch', fetchMock)
  await expect(api.deleteClosure('resource-001', 'closure-001')).resolves.toBeUndefined()
  expect(fetchMock.mock.calls[0]![1].headers.Authorization).toBe('Bearer demo-admin')
})
it('un fallo de red se muestra como indisponibilidad y nunca vuelve a datos locales', async () => {
  vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))
  await expect(api.createResource({ name: 'Recurso', enabled: true, timeZone: 'America/Santiago', policy: { minDurationSeconds: 3600, maxDurationSeconds: 3600, slotDurationSeconds: 3600, cancellationNoticeSeconds: 7200 } })).rejects.toBeInstanceOf(ApiError)
})
