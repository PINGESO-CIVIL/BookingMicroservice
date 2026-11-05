import examples from './examples.json'
import type { Reservation, Resource, ResourceDetail } from './types'
export const confirmed: Reservation = { ...examples.ConfirmedReservation, status: 'confirmed' }
export const cancelled: Reservation = { ...examples.CancelledReservation, status: 'cancelled', id: 'reservation-002', resourceId: 'resource-002' }
export const expired: Reservation = { ...confirmed, id: 'reservation-003', startAt: '2026-11-04T17:00:00Z', endAt: '2026-11-04T18:00:00Z', cancellationDeadline: '2026-11-04T15:00:00Z' }
export const past: Reservation = { ...confirmed, id: 'reservation-004', startAt: '2026-11-03T22:00:00Z', endAt: '2026-11-03T23:00:00Z', cancellationDeadline: '2026-11-03T20:00:00Z', createdAt: '2026-11-01T12:00:00Z', updatedAt: '2026-11-01T12:00:00Z' }
export const reservations: Reservation[] = [past, expired, confirmed, cancelled].sort((a, b) => a.startAt.localeCompare(b.startAt) || a.id.localeCompare(b.id))
export const adminReservations: Reservation[] = [...reservations,
  { ...confirmed, id: 'reservation-005', resourceId: 'resource-002', responsible: { issuer: 'demo', subject: 'user-002' }, startAt: '2026-11-09T13:00:00Z', endAt: '2026-11-09T13:30:00Z', cancellationDeadline: '2026-11-09T11:00:00Z' },
  { ...cancelled, id: 'reservation-006', resourceId: 'resource-001', responsible: { issuer: 'demo', subject: 'user-003' }, startAt: '2026-11-16T15:00:00Z', endAt: '2026-11-16T16:00:00Z', cancellationDeadline: '2026-11-16T13:00:00Z' },
].sort((a, b) => a.startAt.localeCompare(b.startAt) || a.id.localeCompare(b.id))
export const resources: Resource[] = [...examples.ResourcePage.items, { id: 'resource-003', name: 'Recurso 3', enabled: false, timeZone: 'America/Santiago', attributes: {} }]
export function resourceDetail(id: string): ResourceDetail {
  const r = resources.find(r => r.id === id)
  if (!r) throw new Error('Recurso no encontrado.')
  return structuredClone({ ...examples.ResourceDetail, ...r, policy: id === 'resource-002' ? { ...examples.ResourceDetail.policy, slotDurationSeconds: 1800, minDurationSeconds: 1800, maxDurationSeconds: 1800 } : examples.ResourceDetail.policy, closures: id === 'resource-001' ? examples.ResourceDetail.closures : [] })
}
