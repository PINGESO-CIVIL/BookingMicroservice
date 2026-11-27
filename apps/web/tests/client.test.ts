import { afterEach, expect, it, vi } from 'vitest'
import { AgendaError } from '../src/api/agenda'
import { createReservation } from '../src/componentes/formulario-reserva/formulario-reserva.api'
import { cancelReservation } from '../src/componentes/detalle-reserva/detalle-reserva.api'
import { removeClosure } from '../src/componentes/editor-recurso/editor-recurso.api'

const context = { baseUrl: 'http://localhost', token: 'demo-user' }
const reservation = '00000000-0000-4000-8000-200000000001'
const resource = '00000000-0000-4000-8000-100000000001'
afterEach(() => vi.unstubAllGlobals())
it('envía el cuerpo y la clave de idempotencia de la creación', async () => {
  const fetchMock = vi.fn<(request: Request) => Promise<Response>>()
  fetchMock.mockResolvedValue(new Response(JSON.stringify({ id: reservation }), { status: 201 }))
  vi.stubGlobal('fetch', fetchMock)
  const body = { resourceId: resource, startAt: '2026-11-10T21:00:00Z', endAt: '2026-11-10T22:00:00Z', attributes: {} }
  await createReservation(context, body, 'same-request-key')
  const [request] = fetchMock.mock.calls[0]!
  expect(new URL(request.url).pathname).toBe('/api/v1/reservas')
  expect(request.method).toBe('POST')
  expect(request.headers.get('Idempotency-Key')).toBe('same-request-key')
  expect(request.headers.get('Authorization')).toBe('Bearer demo-user')
  expect(await request.json()).toEqual(body)
})
it('conserva el código y el identificador de un conflicto', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { code: 'RESERVATION_VERSION_CONFLICT', message: 'Recarga la reserva.', requestId: 'request-1' } }), { status: 409 })))
  await expect(cancelReservation(context, reservation, 1)).rejects.toMatchObject({ code: 'RESERVATION_VERSION_CONFLICT', requestId: 'request-1' })
})
it('maneja DELETE 204 y fallos de red', async () => {
  const fetchMock = vi.fn().mockResolvedValueOnce(new Response(null, { status: 204 })).mockRejectedValueOnce(new TypeError('Failed to fetch'))
  vi.stubGlobal('fetch', fetchMock)
  await expect(removeClosure(context, resource, reservation)).resolves.toBeUndefined()
  await expect(cancelReservation(context, reservation, 1)).rejects.toBeInstanceOf(AgendaError)
})
