import { reactive } from 'vue'
import type { Availability, Closure, Page, Reservation, Resource, ResourceDetail, Schemas } from './types'
import { DEMO_NOW } from '../demo'

export class ApiError extends Error {
  constructor(public code: string, message: string, public requestId = '') { super(message) }
}
const catalog = reactive(new Map<string, Resource>())
const baseURL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'
async function request<T>(path: string, options: { method?: string; body?: unknown; scenario?: string; admin?: boolean; key?: string; query?: Record<string, unknown> } = {}): Promise<T> {
  const query = new URLSearchParams()
  for (const [key, value] of Object.entries(options.query ?? {})) if (value !== undefined && value !== '') query.set(key, String(value))
  const headers: Record<string, string> = { Authorization: `Bearer ${options.admin ? 'demo-admin' : 'demo-user'}`, 'X-Mock-Scenario': options.scenario ?? 'normal' }
  if (options.body !== undefined) headers['Content-Type'] = 'application/json'
  if (options.key) headers['Idempotency-Key'] = options.key
  let response: Response
  try { response = await fetch(`${baseURL}${path}${query.size ? `?${query}` : ''}`, { method: options.method ?? 'GET', headers, body: options.body === undefined ? undefined : JSON.stringify(options.body), signal: AbortSignal.timeout(10000) }) }
  catch { throw new ApiError('SERVICE_UNAVAILABLE', 'No pudimos conectar con el servidor de demostración. Comprueba que esté iniciado.') }
  if (!response.ok) {
    const data = await response.json().catch(() => ({}))
    throw new ApiError(data.error?.code ?? 'SERVICE_UNAVAILABLE', data.error?.message ?? 'No pudimos completar la solicitud.', data.error?.requestId)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}
function remember<T extends Resource>(resource: T): T { catalog.set(resource.id, resource); return resource }
let catalogLoad: Promise<void> | undefined
async function loadCatalog() {
  catalogLoad ??= Promise.all([true, false].map(async enabled => { const p = await request<Page<Resource>>('/recursos', { query: { enabled } }); p.items.forEach(remember) })).then(() => undefined).catch(e => { catalogLoad = undefined; throw e })
  return catalogLoad
}
export function canChange(r: Reservation, now = DEMO_NOW) { return r.status === 'confirmed' && Date.parse(now) < Date.parse(r.startAt) && Date.parse(now) <= Date.parse(r.cancellationDeadline) }
export const api = {
  async listAdminReservations(query: { resourceId?: string; status?: string; from?: string; to?: string; cursor?: string; limit?: number } = {}, scenario = 'normal') {
    const [page] = await Promise.all([request<Page<Reservation>>('/admin/reservas', { query, scenario, admin: true }), loadCatalog()]); return page
  },
  async listReservations(query: { resourceId?: string; status?: string; from?: string; to?: string; cursor?: string; limit?: number } = {}, scenario = 'normal') {
    const [page] = await Promise.all([request<Page<Reservation>>('/reservas', { query, scenario }), loadCatalog()]); return page
  },
  async listResources(query: { enabled?: boolean; cursor?: string; limit?: number } = {}, scenario = 'normal') {
    const page = await request<Page<Resource>>('/recursos', { query, scenario }); page.items.forEach(remember); return page
  },
  async getReservation(id: string, scenario = 'normal') {
    const [reservation] = await Promise.all([request<Reservation>(`/reservas/${encodeURIComponent(id)}`, { scenario }), loadCatalog()]); return reservation
  },
  async getResource(id: string, scenario = 'normal') { return remember(await request<ResourceDetail>(`/recursos/${encodeURIComponent(id)}`, { admin: true, scenario })) },
  resourceName(id: string) { return catalog.get(id)?.name ?? 'Recurso' },
  resourceOptions() { return [...catalog.values()] },
  resourceZone(id: string) { return catalog.get(id)?.timeZone ?? 'America/Santiago' },
  availability(id: string, from: string, to: string, scenario = 'normal') { return request<Availability>(`/recursos/${encodeURIComponent(id)}/disponibilidades`, { query: { from, to }, scenario }) },
  createReservation(body: Schemas['CreateReservationRequest'], key: string, scenario = 'success') { return request<Reservation>('/reservas', { method: 'POST', body, key, scenario }) },
  updateReservation(id: string, body: Schemas['UpdateReservationRequest'], scenario = 'success') { return request<Reservation>(`/reservas/${encodeURIComponent(id)}`, { method: 'PATCH', body, scenario }) },
  cancelReservation(id: string, body: Schemas['CancelReservationRequest'], scenario = 'success') { return request<Reservation>(`/reservas/${encodeURIComponent(id)}/cancelaciones`, { method: 'POST', body, scenario }) },
  createResource(body: Schemas['CreateResourceRequest'], scenario = 'success') { return request<ResourceDetail>('/recursos', { method: 'POST', body, scenario, admin: true }) },
  updateResource(id: string, body: Schemas['UpdateResourceRequest'], scenario = 'success') { return request<ResourceDetail>(`/recursos/${encodeURIComponent(id)}`, { method: 'PATCH', body, scenario, admin: true }) },
  updateSchedule(id: string, body: Schemas['UpdateResourceScheduleRequest'], scenario = 'success') { return request<ResourceDetail>(`/recursos/${encodeURIComponent(id)}/horarios`, { method: 'PUT', body, scenario, admin: true }) },
  createClosure(id: string, body: Schemas['CreateResourceClosureRequest'], scenario = 'success') { return request<Closure>(`/recursos/${encodeURIComponent(id)}/bloqueos`, { method: 'POST', body, scenario, admin: true }) },
  deleteClosure(id: string, closureId: string, scenario = 'success') { return request<void>(`/recursos/${encodeURIComponent(id)}/bloqueos/${encodeURIComponent(closureId)}`, { method: 'DELETE', scenario, admin: true }) },
}
