const formatters = new Map<string, Intl.DateTimeFormat>()
export function localParts(instant: string, zone: string) {
  if (!formatters.has(zone)) formatters.set(zone, new Intl.DateTimeFormat('sv-SE', { timeZone: zone, year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hourCycle: 'h23' }))
  const parts = formatters.get(zone)!.formatToParts(new Date(instant))
  const p = Object.fromEntries(parts.map(x => [x.type, x.value]))
  return { date: `${p.year}-${p.month}-${p.day}`, time: `${p.hour}:${p.minute}` }
}
export function toInstant(date: string, time: string, zone: string): string {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(date) || !/^\d{2}:\d{2}$/.test(time)) throw new Error('Completa una fecha y hora válidas.')
  const wall = Date.parse(`${date}T${time}:00Z`)
  if (!Number.isFinite(wall)) throw new Error('Fecha inválida.')
  const matches: string[] = []
  for (let offset = -14 * 60; offset <= 14 * 60; offset += 15) {
    const candidate = new Date(wall - offset * 60000).toISOString()
    const p = localParts(candidate, zone)
    if (p.date === date && p.time === time) matches.push(candidate)
  }
  if (matches.length !== 1) throw new Error(matches.length ? 'Esta hora se repite por el cambio de horario. Elige otra hora.' : 'Esta hora no existe en la zona horaria del recurso.')
  return matches[0]!
}
export function formatDate(instant: string, zone = 'America/Santiago') {
  return new Intl.DateTimeFormat('es-CL', { timeZone: zone, weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' }).format(new Date(instant))
}
export function formatDateTime(instant: string, zone = 'America/Santiago') { return `${formatDate(instant, zone)}, ${localParts(instant, zone).time}` }
export function duration(start: string, end: string) {
  const minutes = Math.round((Date.parse(end) - Date.parse(start)) / 60000)
  if (!Number.isFinite(minutes) || minutes <= 0) return '—'
  return [Math.floor(minutes / 60) ? `${Math.floor(minutes / 60)} h` : '', minutes % 60 ? `${minutes % 60} min` : ''].filter(Boolean).join(' ')
}
export const overlaps = (start: string, end: string, from: string, to: string) => Date.parse(start) < Date.parse(to) && Date.parse(end) > Date.parse(from)
