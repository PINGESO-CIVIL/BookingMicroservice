import { describe, expect, it } from 'vitest'
import { configuredAvailability } from '../src/api/availability'
import examples from '../src/api/examples.json'
import type { Reservation, ResourceDetail } from '../src/api/types'
import { localParts } from '../src/dates'
import { validatePolicy } from '../src/validation'
const now = '2026-11-04T16:00:00Z', from = '2026-11-10T03:00:00Z', to = '2026-11-11T03:00:00Z'
const fixture = (): ResourceDetail => structuredClone(examples.ResourceDetail)
const booking: Reservation = { ...examples.ConfirmedReservation, status: 'confirmed' }
const times = (value: ReturnType<typeof configuredAvailability>) => value.availableSlots.map(s => `${localParts(s.startAt, value.timeZone).time}–${localParts(s.endAt, value.timeZone).time}`)
describe('bloques según configuración del recurso', () => {
  it('ofrece bloques completos alineados a la apertura y excluye reservas confirmadas', () => {
    const result = configuredAvailability(fixture(), [booking], from, to, now)
    expect(result.availableSlots).toHaveLength(13)
    expect(times(result)).toContain('08:00–09:00')
    expect(times(result)).not.toContain('19:00–20:00')
    expect(times(result)).toContain('20:00–21:00')
    expect(result.availableSlots.every(s => Date.parse(s.endAt) - Date.parse(s.startAt) === 3600000)).toBe(true)
  })
  it('cambiar la duración cambia las opciones, sin producir restos parciales', () => {
    const r = fixture(); r.policy.slotDurationSeconds = 5400
    const result = configuredAvailability(r, [], from, to, now)
    expect(times(result)[0]).toBe('08:00–09:30')
    expect(times(result).at(-1)).toBe('20:00–21:30')
    expect(result.availableSlots).toHaveLength(9)
  })
  it('un bloqueo parcial elimina el bloque entero y no desplaza su alineación', () => {
    const r = fixture()
    r.closures.push({ id: 'closure-test', resourceId: r.id, reason: 'Mantenimiento', startAt: '2026-11-10T13:15:00Z', endAt: '2026-11-10T13:45:00Z', createdAt: now })
    const result = configuredAvailability(r, [], from, to, now)
    expect(times(result)).not.toContain('10:00–11:00')
    expect(times(result)).toContain('11:00–12:00')
    expect(times(result).some(t => t.startsWith('10:45'))).toBe(false)
  })
  it('reservas canceladas y de otros recursos no ocupan bloques', () => {
    expect(times(configuredAvailability(fixture(), [{ ...booking, status: 'cancelled' }, { ...booking, resourceId: 'otro' }], from, to, now))).toContain('19:00–20:00')
  })
  it('no ofrece días cerrados, recursos deshabilitados ni bloques ya iniciados', () => {
    const r = fixture()
    expect(configuredAvailability(r, [], '2026-11-11T03:00:00Z', '2026-11-12T03:00:00Z', now).availableSlots).toEqual([])
    expect(configuredAvailability({ ...r, enabled: false }, [], from, to, now).availableSlots).toEqual([])
    const result = configuredAvailability(r, [], from, to, '2026-11-10T12:30:00Z')
    expect(times(result)[0]).toBe('10:00–11:00')
  })
  it('respeta varias franjas y nunca cruza el cierre intermedio', () => {
    const r = fixture(); r.weeklySchedule = [{ dayOfWeek: 2, startLocalTime: '08:00', endLocalTime: '10:30' }, { dayOfWeek: 2, startLocalTime: '14:30', endLocalTime: '16:00' }]
    expect(times(configuredAvailability(r, [], from, to, now))).toEqual(['08:00–09:00', '09:00–10:00', '14:30–15:30'])
  })
  it('no ofrece fragmentos al consultar un rango que empieza a mitad de un bloque', () => {
    const result = configuredAvailability(fixture(), [], '2026-11-10T11:30:00Z', '2026-11-10T13:30:00Z', now)
    expect(times(result)).toEqual(['09:00–10:00'])
  })
  it('rechaza duraciones fraccionarias y duraciones fijas fuera de los límites', () => {
    const r = fixture(); r.policy.slotDurationSeconds = 90
    expect(validatePolicy(r.policy)).toContain('minutos')
    r.policy.slotDurationSeconds = 1800
    expect(validatePolicy(r.policy)).toContain('límites')
  })
})
