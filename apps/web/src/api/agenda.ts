import createClient from 'openapi-fetch'
import type { paths } from './schema'

export interface AgendaContext {
  baseUrl: string
  token: string
  permissions?: readonly string[]
  scenario?: string
  writeScenario?: string
  now?: string
}

export class AgendaError extends Error {
  constructor(message: string, public readonly code?: string, public readonly requestId?: string) { super(message) }
}

export function agenda(context: AgendaContext) {
  const baseUrl = new URL(context.baseUrl || '/', globalThis.location?.origin).href.replace(/\/$/, '')
  const headers = { Authorization: 'Bearer ' + context.token, 'X-Mock-Scenario': context.scenario ?? 'normal' }
  return createClient<paths>({ baseUrl, headers, fetch: request => fetch(request, { signal: request.signal.aborted ? request.signal : AbortSignal.any([request.signal, AbortSignal.timeout(5000)]) }) })
}

export function writeHeaders(context: AgendaContext) {
  return { 'X-Mock-Scenario': context.writeScenario ?? 'normal' }
}

export async function result<T>(call: Promise<{ data?: T; error?: unknown; response: Response }>): Promise<T> {
  const { data, error, response } = await call.catch(() => { throw new AgendaError('No pudimos conectar con Agenda.') })
  if (!response.ok) {
    const detail = (error as { error?: { message?: string; code?: string; requestId?: string } })?.error
    throw new AgendaError(detail?.message ?? 'No pudimos completar la solicitud.', detail?.code, detail?.requestId)
  }
  return data as T
}

export function queryKey(context: AgendaContext, ...key: unknown[]) {
  return ['agenda', context.baseUrl, context.token, context.scenario, ...key]
}

export function permits(context: AgendaContext, permission: string) {
  return context.permissions === undefined || context.permissions.includes(permission)
}
