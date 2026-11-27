import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useFieldArray, useForm } from 'react-hook-form'
import { queryKey, type AgendaContext } from '../../api/agenda'
import { toInstant } from '../../dates'
import type { ResourceDetail, Schemas } from '../../api/types'
import { validatePolicy, validateSchedule } from '../../validation'
import { addClosure, create, removeClosure, resource, update, updateSchedule } from './editor-recurso.api'

type Fields = Omit<Schemas['CreateResourceRequest'], 'weeklySchedule'> & { weeklySchedule: NonNullable<Schemas['CreateResourceRequest']['weeklySchedule']> }
const defaults: Fields = {
  name: '', enabled: true, timeZone: 'America/Santiago', weeklySchedule: [],
  policy: { cancellationNoticeSeconds: 7200, minDurationSeconds: 3600, maxDurationSeconds: 3600, slotDurationSeconds: 3600 },
}
function values(row: ResourceDetail | undefined): Fields {
  return row ? { name: row.name, enabled: row.enabled, timeZone: row.timeZone, policy: row.policy, weeklySchedule: row.weeklySchedule } : defaults
}
export function useEditorRecurso(context: AgendaContext, id?: string) {
  const client = useQueryClient()
  const key = queryKey(context, 'recurso', id)
  const query = useQuery({ queryKey: key, enabled: !!id, queryFn: () => resource(context, id!) })
  const form = useForm<Fields>({ values: values(query.data) })
  const rules = useFieldArray({ control: form.control, name: 'weeklySchedule' })
  const accept = (row: ResourceDetail) => {
    if (id) client.setQueryData(key, row)
    form.reset(values(row))
  }
  const save = useMutation({ mutationFn: ({ fields, schedule }: { fields: Fields; schedule: boolean }) => {
    const problem = validatePolicy(fields.policy) || ((schedule || !id) && validateSchedule(fields.weeklySchedule))
    if (problem) throw new Error(problem)
    const { weeklySchedule, ...body } = fields
    if (!body.name.trim()) throw new Error('El nombre es obligatorio.')
    try { new Intl.DateTimeFormat('es', { timeZone: body.timeZone }).format() } catch { throw new Error('Ingresa una zona horaria IANA válida.') }
    if (!id) return create(context, { ...body, weeklySchedule })
    const expectedVersion = query.data!.version
    if (schedule) return updateSchedule(context, id, { expectedVersion, policy: body.policy, weeklySchedule })
    return update(context, id, { ...body, expectedVersion })
  }, onSuccess: accept })
  const closures = useMutation({ mutationFn: async (input: { id: string } | { start: string; end: string; reason: string }) => {
    const current = query.data!
    if ('id' in input) {
      await removeClosure(context, current.id, input.id)
      return { ...current, closures: current.closures.filter(c => c.id !== input.id) }
    }
    const instant = (value: string) => toInstant(value.slice(0, 10), value.slice(11), current.timeZone)
    const startAt = instant(input.start)
    const endAt = instant(input.end)
    if (Date.parse(startAt) >= Date.parse(endAt)) throw new Error('El inicio debe ser anterior al término.')
    const created = await addClosure(context, current.id, { startAt, endAt, reason: input.reason })
    return { ...current, closures: [...current.closures.filter(c => c.id !== created.id), created] }
  }, onSuccess: accept })
  return { query, form, rules, save, closures }
}
