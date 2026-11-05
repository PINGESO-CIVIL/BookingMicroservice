import type { Availability, Reservation, ResourceDetail } from './types'
import { localParts, overlaps, toInstant } from '../dates'

type Window = { startAt: string; endAt: string }
const iso = (instant: number) => new Date(instant).toISOString()
function subtract(windows: Window[], blocked: Window): Window[] {
  return windows.flatMap(w => {
    if (!overlaps(w.startAt, w.endAt, blocked.startAt, blocked.endAt)) return [w]
    const result: Window[] = []
    if (Date.parse(w.startAt) < Date.parse(blocked.startAt)) result.push({ startAt: w.startAt, endAt: blocked.startAt })
    if (Date.parse(w.endAt) > Date.parse(blocked.endAt)) result.push({ startAt: blocked.endAt, endAt: w.endAt })
    return result
  })
}

// Stateless demo calculation over fixture configuration. No persistence or booking guarantees.
export function configuredAvailability(resource: ResourceDetail, reservations: Reservation[], from: string, to: string, now: string): Availability {
  const result: Availability = { resourceId: resource.id, enabled: resource.enabled, timeZone: resource.timeZone, from, to, slotDurationSeconds: resource.policy.slotDurationSeconds, availableWindows: [], availableSlots: [] }
  if (!resource.enabled) return result
  const step = resource.policy.slotDurationSeconds * 1000
  if (!Number.isInteger(step) || step < 60000 || step % 60000) throw new Error('La duración de cada reserva debe ser un número entero de minutos.')
  const busy: Window[] = [...resource.closures, ...reservations.filter(r => r.resourceId === resource.id && r.status === 'confirmed')]
  const first = localParts(from, resource.timeZone).date, last = localParts(to, resource.timeZone).date
  for (let day = new Date(`${first}T12:00:00Z`); day.toISOString().slice(0, 10) <= last; day.setUTCDate(day.getUTCDate() + 1)) {
    const date = day.toISOString().slice(0, 10), weekday = day.getUTCDay() || 7
    for (const rule of resource.weeklySchedule.filter(r => r.dayOfWeek === weekday)) {
      let opening: number, closing: number
      try { opening = Date.parse(toInstant(date, rule.startLocalTime, resource.timeZone)); closing = Date.parse(toInstant(date, rule.endLocalTime, resource.timeZone)) } catch { continue }
      const start = Math.max(opening, Date.parse(from)), end = Math.min(closing, Date.parse(to))
      if (start >= end) continue
      let windows: Window[] = [{ startAt: iso(start), endAt: iso(end) }]
      for (const blocked of busy) windows = subtract(windows, blocked)
      result.availableWindows.push(...windows)
      // Anchor blocks to each configured opening, never to the first free minute after a collision.
      for (let at = opening; at + step <= closing; at += step) {
        const slot = { startAt: iso(at), endAt: iso(at + step) }
        if (at < start || at + step > end || at <= Date.parse(now) || busy.some(b => overlaps(slot.startAt, slot.endAt, b.startAt, b.endAt))) continue
        // Avoid offering ambiguous clock labels on daylight-saving transitions.
        try {
          const s = localParts(slot.startAt, resource.timeZone), e = localParts(slot.endAt, resource.timeZone)
          if (Date.parse(toInstant(s.date, s.time, resource.timeZone)) !== at || Date.parse(toInstant(e.date, e.time, resource.timeZone)) !== at + step) continue
        } catch { continue }
        result.availableSlots.push(slot)
      }
    }
  }
  result.availableSlots.sort((a, b) => a.startAt.localeCompare(b.startAt))
  const merged: Window[] = []
  for (const w of result.availableWindows.sort((a, b) => a.startAt.localeCompare(b.startAt))) {
    const previous = merged[merged.length - 1]
    if (previous && Date.parse(w.startAt) <= Date.parse(previous.endAt)) previous.endAt = iso(Math.max(Date.parse(previous.endAt), Date.parse(w.endAt)))
    else merged.push({ ...w })
  }
  result.availableWindows = merged
  return result
}
