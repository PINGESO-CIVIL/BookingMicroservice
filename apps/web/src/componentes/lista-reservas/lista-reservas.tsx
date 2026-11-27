import { useState } from 'react'
import { permits, type AgendaContext } from '../../api/agenda'
import { formatDate, localParts } from '../../dates'
import { Action, AgendaBoundary, Notice } from '../shared'
import { useListaReservas } from './use-lista-reservas'
import type { ReservationFilters } from './lista-reservas.api'

export interface ListaReservasProps extends AgendaContext {
  admin?: boolean
  onOpen?: (id: string) => void
  onCreate?: () => void
}

function View(props: ListaReservasProps) {
  const [status, setStatus] = useState<ReservationFilters['status']>()
  const [resourceId, setResourceId] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const range = from && to && from <= to ? { from: from + 'T00:00:00Z', to: to + 'T23:59:59Z' } : {}
  const { reservations, resources } = useListaReservas(props, !!props.admin, { status, resourceId: resourceId || undefined, ...range })
  const names = resources.data ?? {}
  const rows = reservations.data?.pages.flatMap(page => page.items) ?? []
  return <>
    <div className="agenda-toolbar">
      <div className="agenda-toolbar-heading">
        <h2>Filtros</h2>
        {!props.admin && permits(props, 'reservas:operar') && props.onCreate && <Action onClick={props.onCreate}><span aria-hidden="true">＋</span> Nueva reserva</Action>}
      </div>
      <div className="agenda-filters">
        <label>Estado<select value={status ?? ''} onChange={e => setStatus((e.target.value || undefined) as typeof status)}><option value="">Todos los estados</option><option value="confirmed">Confirmada</option><option value="cancelled">Cancelada</option></select></label>
        <label>Recurso<select value={resourceId} onChange={e => setResourceId(e.target.value)}><option value="">Todos los recursos</option>{Object.entries(names).map(([id, name]) => <option key={id} value={id}>{name}</option>)}</select></label>
        <label>Desde<input type="date" value={from} onChange={e => setFrom(e.target.value)} /></label>
        <label>Hasta<input type="date" value={to} onChange={e => setTo(e.target.value)} /></label>
        <Action className="agenda-button-secondary" onClick={() => { setStatus(undefined); setResourceId(''); setFrom(''); setTo('') }}>Limpiar filtros</Action>
      </div>
    </div>
    <Notice error={reservations.error} retry={() => void reservations.refetch()} />
    {reservations.isPending && <p className="agenda-message" role="status">Cargando reservas…</p>}
    {reservations.isSuccess && !rows.length && <p className="agenda-message">No hay reservas que coincidan.</p>}
    <div className="agenda-list">
      {rows.map(row => {
        const zone = 'America/Santiago'
        const start = localParts(row.startAt, zone)
        const end = localParts(row.endAt, zone)
        return <article className="agenda-list-row" key={row.id}>
          <div className="agenda-date-tile" aria-hidden="true">
            <span>{new Intl.DateTimeFormat('es-CL', { month: 'short', timeZone: zone }).format(new Date(row.startAt))}</span>
            <strong>{start.date.slice(-2)}</strong>
          </div>
          <div className="agenda-list-body">
            <strong>{names[row.resourceId] ?? 'Recurso'}</strong>
            <p><time dateTime={row.startAt}>{formatDate(row.startAt, zone)} · {start.time}</time>–{end.time}</p>
            {props.admin && <p>{row.responsible.subject}</p>}
          </div>
          <span className="agenda-status" data-status={row.status}>{row.status === 'confirmed' ? 'Confirmada' : 'Cancelada'}</span>
          {props.onOpen && <Action className="agenda-button-secondary agenda-row-action" onClick={() => props.onOpen?.(row.id)}>Ver detalle <span aria-hidden="true">→</span></Action>}
        </article>
      })}
    </div>
    {reservations.hasNextPage && <Action disabled={reservations.isFetchingNextPage} onClick={() => void reservations.fetchNextPage()}>Cargar más</Action>}
  </>
}
export function ListaReservas(props: ListaReservasProps) { return <AgendaBoundary><View {...props} /></AgendaBoundary> }
