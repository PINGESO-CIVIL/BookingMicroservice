import { agenda, result, writeHeaders, type AgendaContext } from '../../api/agenda'
import type { Schemas } from '../../api/types'
export const resources = (context: AgendaContext, cursor?: string) => result(agenda(context).GET('/api/v1/recursos', { params: { query: { cursor } } }))
export const original = (context: AgendaContext, id: string) => result(agenda(context).GET('/api/v1/reservas/{id}', { params: { path: { id } } }))
export const availability = (context: AgendaContext, id: string, from: string, to: string) => result(agenda(context).GET('/api/v1/recursos/{id}/disponibilidades', { params: { path: { id }, query: { from, to } } }))
export const createReservation = (context: AgendaContext, body: Schemas['CreateReservationRequest'], key: string) => result(agenda(context).POST('/api/v1/reservas', { params: { header: { 'Idempotency-Key': key } }, body, headers: writeHeaders(context) }))
export const updateReservation = (context: AgendaContext, id: string, body: Schemas['UpdateReservationRequest']) => result(agenda(context).PATCH('/api/v1/reservas/{id}', { params: { path: { id } }, body, headers: writeHeaders(context) }))
