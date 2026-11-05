import type { Policy, ScheduleRule } from './api/types'
export function validatePolicy(p: Policy) {
  if (!Number.isInteger(p.slotDurationSeconds) || p.slotDurationSeconds < 60 || p.slotDurationSeconds % 60 !== 0) return 'La duración de cada reserva debe ser un número entero de minutos, mayor o igual a uno.'
  if (!Object.values(p).every(Number.isInteger)) return 'Las duraciones deben representar segundos enteros.'
  if (p.minDurationSeconds < 60 || p.maxDurationSeconds < p.minDurationSeconds || p.cancellationNoticeSeconds < 0) return 'Revisa las duraciones: mínimo un minuto, máximo igual o superior al mínimo y anticipación no negativa.'
  if (p.slotDurationSeconds < p.minDurationSeconds || p.slotDurationSeconds > p.maxDurationSeconds) return 'La duración fija debe estar dentro de los límites del recurso.'
  return ''
}
export function validateSchedule(rules: ScheduleRule[]) {
  const pattern = /^(?:[01]\d|2[0-3]):[0-5]\d$/
  for (const r of rules) {
    if (!pattern.test(r.startLocalTime) || !pattern.test(r.endLocalTime) || r.startLocalTime >= r.endLocalTime || !Number.isInteger(r.dayOfWeek) || r.dayOfWeek < 1 || r.dayOfWeek > 7) return 'Cada franja necesita un inicio anterior al término dentro del mismo día.'
  }
  for (let i = 0; i < rules.length; i++) for (let j = i + 1; j < rules.length; j++) {
    const a = rules[i]!, b = rules[j]!
    if (a.dayOfWeek === b.dayOfWeek && a.startLocalTime < b.endLocalTime && a.endLocalTime > b.startLocalTime) return 'Las franjas del mismo día no pueden solaparse.'
  }
  return ''
}
