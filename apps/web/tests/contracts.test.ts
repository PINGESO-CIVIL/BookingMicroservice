import { describe, expect, it } from 'vitest'
import { canChange } from '../src/validation'
import { confirmed } from '../src/api/fixtures'
import { duration, localParts, overlaps, toInstant } from '../src/dates'
import { validatePolicy, validateSchedule } from '../src/validation'
describe('fechas y comprobaciones visuales', () => {
  it('permite el límite inclusivo y rechaza el inicio', () => {
    expect(canChange(confirmed, confirmed.cancellationDeadline)).toBe(true)
    expect(
      canChange(
        confirmed,
        new Date(Date.parse(confirmed.cancellationDeadline) + 1).toISOString(),
      ),
    ).toBe(false)
    expect(
      canChange(
        {
          ...confirmed,
          cancellationNoticeSeconds: 0,
          cancellationDeadline: confirmed.startAt,
        },
        confirmed.startAt,
      ),
    ).toBe(false)
    expect(canChange({ ...confirmed, status: 'cancelled' })).toBe(false)
  })
  it('convierte por zona del recurso independientemente de la zona del equipo', () => {
    expect(toInstant('2026-11-10', '19:00', 'America/Santiago')).toBe(
      '2026-11-10T22:00:00.000Z',
    )
    expect(toInstant('2026-07-10', '19:00', 'America/Santiago')).toBe(
      '2026-07-10T23:00:00.000Z',
    )
    expect(localParts('2026-11-10T22:00:00Z', 'America/Santiago')).toEqual({
      date: '2026-11-10',
      time: '19:00',
    })
    expect(duration('2026-11-10T22:00:00Z', '2026-11-10T23:30:00Z')).toBe('1 h 30 min')
    expect(() => toInstant('2026-02-31', '12:00', 'UTC')).toThrow()
    expect(toInstant('2026-11-10', '12:00', 'UTC')).toBe('2026-11-10T12:00:00.000Z')
    expect(localParts('2026-11-10T22:00:00Z', 'UTC')).toEqual({
      date: '2026-11-10',
      time: '22:00',
    })
    expect(toInstant('2026-11-11', '11:00', 'Pacific/Auckland')).toBe(
      '2026-11-10T22:00:00.000Z',
    )
    expect(localParts('2026-11-10T22:00:00Z', 'Pacific/Auckland')).toEqual({
      date: '2026-11-11',
      time: '11:00',
    })
  })
  it('rechaza horas inexistentes y ambiguas por cambio de horario', () => {
    expect(() => toInstant('2026-03-08', '02:30', 'America/New_York')).toThrow(
      'no existe',
    )
    expect(() => toInstant('2026-11-01', '01:30', 'America/New_York')).toThrow(
      'se repite',
    )
  })
  it('no considera colisión los intervalos consecutivos', () => {
    expect(
      overlaps(
        '2026-11-10T22:00:00Z',
        '2026-11-10T23:00:00Z',
        '2026-11-10T23:00:00Z',
        '2026-11-11T00:00:00Z',
      ),
    ).toBe(false)
  })
  it('valida políticas y múltiples franjas semanales', () => {
    expect(
      validatePolicy({
        minDurationSeconds: 60,
        maxDurationSeconds: 120,
        cancellationNoticeSeconds: 0,
        slotDurationSeconds: 60,
      }),
    ).toBe('')
    expect(
      validatePolicy({
        minDurationSeconds: 120,
        maxDurationSeconds: 60,
        cancellationNoticeSeconds: 0,
        slotDurationSeconds: 60,
      }),
    ).not.toBe('')
    expect(
      validateSchedule([
        { dayOfWeek: 1, startLocalTime: '08:00', endLocalTime: '12:00' },
        { dayOfWeek: 1, startLocalTime: '12:00', endLocalTime: '18:00' },
      ]),
    ).toBe('')
    expect(
      validateSchedule([
        { dayOfWeek: 1, startLocalTime: '08:00', endLocalTime: '13:00' },
        { dayOfWeek: 1, startLocalTime: '12:00', endLocalTime: '18:00' },
      ]),
    ).toContain('solaparse')
  })
})
