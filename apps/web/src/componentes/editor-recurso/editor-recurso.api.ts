import { agenda, result, writeHeaders, type AgendaContext } from '../../api/agenda'
import type { Schemas } from '../../api/types'
export const resource = (context: AgendaContext, id: string) => result(agenda(context).GET('/api/v1/recursos/{id}', { params: { path: { id } } }))
export const create = (context: AgendaContext, body: Schemas['CreateResourceRequest']) => result(agenda(context).POST('/api/v1/recursos', { body, headers: writeHeaders(context) }))
export const update = (context: AgendaContext, id: string, body: Schemas['UpdateResourceRequest']) => result(agenda(context).PATCH('/api/v1/recursos/{id}', { params: { path: { id } }, body, headers: writeHeaders(context) }))
export const updateSchedule = (context: AgendaContext, id: string, body: Schemas['UpdateResourceScheduleRequest']) => result(agenda(context).PUT('/api/v1/recursos/{id}/horarios', { params: { path: { id } }, body, headers: writeHeaders(context) }))
export const addClosure = (context: AgendaContext, id: string, body: Schemas['CreateResourceClosureRequest']) => result(agenda(context).POST('/api/v1/recursos/{id}/bloqueos', { params: { path: { id } }, body, headers: writeHeaders(context) }))
export const removeClosure = (context: AgendaContext, id: string, closureId: string) => result(agenda(context).DELETE('/api/v1/recursos/{id}/bloqueos/{closureId}', { params: { path: { id, closureId } }, headers: writeHeaders(context) }))
