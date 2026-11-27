import { useInfiniteQuery, useMutation, useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { permits, queryKey, type AgendaContext } from '../../api/agenda'
import { canChange } from '../../validation'
import { localParts, toInstant } from '../../dates'
import { availability, createReservation, original, resources, updateReservation } from './formulario-reserva.api'

type Fields = { resourceId: string; date: string; startAt: string; endAt: string }
export function useFormularioReserva(context: AgendaContext, id?: string) {
  const form = useForm<Fields>({ defaultValues: { resourceId: '', date: '', startAt: '', endAt: '' } })
  const [attempt, setAttempt] = useState<{ body: string; key: string }>()
  const list = useInfiniteQuery({
    queryKey: queryKey(context, 'reservables'), initialPageParam: undefined as string | undefined,
    queryFn: ({ pageParam }) => resources(context, pageParam), getNextPageParam: page => page.nextCursor ?? undefined,
  })
  const existing = useQuery({ queryKey: queryKey(context, 'original', id), enabled: !!id, queryFn: () => original(context, id!) })
  const row = existing.data
  const items = list.data?.pages.flatMap(page => page.items) ?? []
  const values = form.watch()
  const resourceId = values.resourceId || row?.resourceId || ''
  const zone = items.find(r => r.id === resourceId)?.timeZone ?? 'America/Santiago'
  const now = context.now ?? new Date().toISOString()
  const date = values.date || localParts(row?.startAt ?? now, zone).date
  const blankSlot = { startAt: '', endAt: '' }
  const slot = values.startAt ? values : blankSlot
  const slots = useQuery({
    queryKey: queryKey(context, 'disponibilidad', resourceId, date, zone),
    enabled: !!resourceId,
    queryFn: () => {
      const tomorrow = new Date(Date.parse(date + 'T12:00:00Z') + 86_400_000).toISOString().slice(0, 10)
      return availability(context, resourceId, toInstant(date, '00:00', zone), toInstant(tomorrow, '00:00', zone))
    },
  })
  const selectedSlots = slots.data?.availableSlots ?? []
  const currentSlot = row && resourceId === row.resourceId && localParts(row.startAt, zone).date === date
    ? { startAt: row.startAt, endAt: row.endAt }
    : undefined
  const options = currentSlot && !selectedSlots.some(value => Date.parse(value.startAt) === Date.parse(currentSlot.startAt))
    ? [...selectedSlots, currentSlot]
    : selectedSlots
  const changed = !row || resourceId !== row.resourceId || Date.parse(slot.startAt) !== Date.parse(row.startAt) || Date.parse(slot.endAt) !== Date.parse(row.endAt)
  const save = useMutation({ mutationFn: () => {
    const body = { resourceId, startAt: slot.startAt, endAt: slot.endAt }
    if (row) {
      const changes = { expectedVersion: row.version,
        ...(body.resourceId !== row.resourceId ? { resourceId: body.resourceId } : {}),
        ...(Date.parse(body.startAt) !== Date.parse(row.startAt) ? { startAt: body.startAt } : {}),
        ...(Date.parse(body.endAt) !== Date.parse(row.endAt) ? { endAt: body.endAt } : {}) }
      if (Object.keys(changes).length === 1) throw new Error('Selecciona un horario diferente.')
      return updateReservation(context, row.id, changes)
    }
    const signature = JSON.stringify(body)
    const key = attempt?.body === signature ? attempt.key : crypto.randomUUID()
    setAttempt({ body: signature, key })
    return createReservation(context, { ...body, attributes: {} }, key)
  } })
  const locked = !permits(context, 'reservas:operar') || (!!row && !canChange(row, now))
  return { form, list, existing, resourceId, date, zone, slot, slots, items, save, locked, options, changed }
}
