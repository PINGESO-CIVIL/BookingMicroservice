import { agenda, result, writeHeaders, type AgendaContext } from '../../api/agenda'
export const getReservation = (context: AgendaContext, id: string) => result(agenda(context).GET('/api/v1/reservas/{id}', { params: { path: { id } } }))
export const cancelReservation = (context: AgendaContext, id: string, expectedVersion: number) => result(agenda(context).POST('/api/v1/reservas/{id}/cancelaciones', { params: { path: { id } }, body: { expectedVersion }, headers: writeHeaders(context) }))
