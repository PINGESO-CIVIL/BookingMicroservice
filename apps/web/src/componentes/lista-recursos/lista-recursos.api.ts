import { agenda, result, type AgendaContext } from '../../api/agenda'
export function resourceList(context: AgendaContext, enabled: boolean, cursor?: string) {
  return result(agenda(context).GET('/api/v1/recursos', { params: { query: { enabled, cursor } } }))
}
