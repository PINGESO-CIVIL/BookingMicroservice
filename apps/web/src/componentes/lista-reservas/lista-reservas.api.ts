import { agenda, result, type AgendaContext } from '../../api/agenda'
import type { paths } from '../../api/schema'

export type ReservationFilters = NonNullable<paths['/api/v1/reservas']['get']['parameters']['query']>

export function reservationList(context: AgendaContext, admin: boolean, query: ReservationFilters) {
  const client = agenda(context)
  return result(admin ? client.GET('/api/v1/admin/reservas', { params: { query } }) : client.GET('/api/v1/reservas', { params: { query } }))
}
export async function resourceNames(context: AgendaContext) {
  const client = agenda({ ...context, scenario: 'normal' })
  const pages = await Promise.all([true, false].map(enabled => result(client.GET('/api/v1/recursos', { params: { query: { enabled } } }))))
  return Object.fromEntries(pages.flatMap(page => page.items.map(r => [r.id, r.name])))
}
