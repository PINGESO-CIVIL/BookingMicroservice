import { http, HttpResponse, type HttpResponseResolver } from 'msw'
import examples from '../../api/examples.json'
import { adminReservations, reservations, resourceDetail, resources } from '../../api/fixtures'
import { configuredAvailability } from '../../api/availability'

export const demoNow = '2026-11-04T16:00:00Z'
const errors = { error: examples.ServiceUnavailable, unauthorized: examples.Unauthenticated, forbidden: examples.Forbidden }
const statuses = { error: 503, unauthorized: 401, forbidden: 403 }
function simulate(resolver: HttpResponseResolver): HttpResponseResolver {
  return args => {
    const scenario = args.request.headers.get('X-Mock-Scenario') ?? ''
    if (scenario in errors) {
      const key = scenario as keyof typeof errors
      return HttpResponse.json(errors[key], { status: statuses[key] })
    }
    return resolver(args)
  }
}
const root = '*/api/v1'
function page<T>(rows: T[], request: Request) {
  const url = new URL(request.url)
  const scenario = request.headers.get('X-Mock-Scenario')
  const limit = Number(url.searchParams.get('limit') ?? (scenario === 'paginated' ? '1' : '20'))
  const start = Number(url.searchParams.get('cursor') ?? '0')
  return HttpResponse.json({ items: scenario === 'empty' ? [] : rows.slice(start, start + limit), nextCursor: scenario === 'empty' || start + limit >= rows.length ? null : String(start + limit) })
}
function list(rows: typeof reservations): HttpResponseResolver {
  return simulate(({ request }) => {
    const q = new URL(request.url).searchParams
    return page(rows.filter(r => (!q.get('status') || q.get('status') === r.status) &&
      (!q.get('resourceId') || q.get('resourceId') === r.resourceId) &&
      (!q.get('from') || r.endAt > q.get('from')!) &&
      (!q.get('to') || r.startAt < q.get('to')!)), request)
  })
}
function write<T extends object>(value: T, status = 200): HttpResponseResolver {
  return simulate(async ({ request }) => {
    const scenario = request.headers.get('X-Mock-Scenario')
    if (scenario === 'conflict') return HttpResponse.json(examples.ResourceUnavailable, { status: 409 })
    if (scenario === 'version') return HttpResponse.json(examples.VersionConflict, { status: 409 })
    if (scenario === 'deadline') return HttpResponse.json(examples.CancellationWindowClosed, { status: 422 })
    if (scenario === 'policy') return HttpResponse.json(examples.PolicyViolation, { status: 422 })
    const body = await request.json() as Record<string, unknown>
    const { expectedVersion, ...fields } = body
    return HttpResponse.json({ ...value, ...fields, ...(typeof expectedVersion === 'number' ? { version: expectedVersion + 1 } : {}) }, { status })
  })
}
export const handlers = [
  http.get(root + '/reservas', list(reservations)),
  http.get(root + '/admin/reservas', list(adminReservations)),
  http.post(root + '/reservas', write(examples.ConfirmedReservation, 201)),
  http.get(root + '/reservas/:id', simulate(({ params, request }) => {
    const row = reservations.find(r => r.id === params.id)
    if (!row) return HttpResponse.json(examples.ReservationNotFound, { status: 404 })
    return HttpResponse.json(request.headers.get('X-Mock-Scenario') === 'cancelled' ? { ...examples.CancelledReservation, id: row.id } : row)
  })),
  http.patch(root + '/reservas/:id', write(examples.UpdatedReservation)),
  http.post(root + '/reservas/:id/cancelaciones', write(examples.CancelledReservation)),
  http.get(root + '/recursos', simulate(({ request }) => {
    const enabled = new URL(request.url).searchParams.get('enabled') !== 'false'
    return page(resources.filter(r => r.enabled === enabled), request)
  })),
  http.post(root + '/recursos', write(examples.ResourceDetail, 201)),
  http.get(root + '/recursos/:id', simulate(({ params }) => resources.some(r => r.id === params.id)
    ? HttpResponse.json(resourceDetail(String(params.id))) : HttpResponse.json(examples.ResourceNotFound, { status: 404 }))),
  http.patch(root + '/recursos/:id', write(examples.UpdatedResourceDetail)),
  http.put(root + '/recursos/:id/horarios', write(examples.UpdatedResourceDetail)),
  http.post(root + '/recursos/:id/bloqueos', write(examples.ResourceClosure, 201)),
  http.delete(root + '/recursos/:id/bloqueos/:closureId', simulate(() => new HttpResponse(null, { status: 204 }))),
  http.get(root + '/recursos/:id/disponibilidades', simulate(({ params, request }) => {
    const q = new URL(request.url).searchParams
    const from = q.get('from'), to = q.get('to')
    if (!from || !to) return HttpResponse.json(examples.InvalidRequest, { status: 400 })
    const detail = resourceDetail(String(params.id))
    const data = configuredAvailability(detail, adminReservations, from, to, demoNow)
    return HttpResponse.json(request.headers.get('X-Mock-Scenario') === 'empty' ? { ...data, availableSlots: [] } : data)
  })),
]

export async function startMocks() {
  const { setupWorker } = await import('msw/browser')
  return setupWorker(...handlers).start({ onUnhandledRequest: 'bypass' })
}
