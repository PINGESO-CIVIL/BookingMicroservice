import { useState } from 'react'
import { permits, type AgendaContext } from '../../api/agenda'
import { canChange } from '../../validation'
import { formatDateTime } from '../../dates'
import { Action, AgendaBoundary, Notice } from '../shared'
import { useDetalleReserva } from './use-detalle-reserva'

export interface DetalleReservaProps extends AgendaContext { id: string; onEdit?: (id: string) => void; onBack?: () => void }
function View(props: DetalleReservaProps) {
  const { query, cancel } = useDetalleReserva(props, props.id)
  const [confirm, setConfirm] = useState(false)
  const row = query.data
  const editable = !!row && permits(props, 'reservas:operar') && canChange(row, props.now)
  return <>
    {props.onBack && <Action className="agenda-button-secondary" onClick={props.onBack}><span aria-hidden="true">←</span> Volver a mis reservas</Action>}
    <Notice error={query.error} retry={() => void query.refetch()} />
    {query.isPending && <p role="status">Cargando reserva…</p>}
    {row && <section className="agenda-card">
      <div className="agenda-toolbar-heading">
        <h2>Horario de la reserva</h2>
        <span className="agenda-status" data-status={row.status}>{row.status === 'confirmed' ? 'Confirmada' : 'Cancelada'}</span>
      </div>
      <dl className="agenda-facts">
        <div><dt>Inicio</dt><dd><time dateTime={row.startAt}>{formatDateTime(row.startAt)}</time></dd></div>
        <div><dt>Término</dt><dd><time dateTime={row.endAt}>{formatDateTime(row.endAt)}</time></dd></div>
        <div><dt>Plazo para cancelar</dt><dd><time dateTime={row.cancellationDeadline}>{formatDateTime(row.cancellationDeadline)}</time></dd></div>
      </dl>
      {editable && <div className="agenda-row">
        {props.onEdit && <Action onClick={() => props.onEdit?.(row.id)}>Modificar reserva</Action>}
        <Action className="agenda-button-secondary" onClick={() => setConfirm(true)}>Cancelar reserva</Action>
      </div>}
      {confirm && <section role="dialog" aria-label="Cancelar reserva">
        <p>¿Quieres cancelar esta reserva?</p><Notice error={cancel.error} />
        <Action className="agenda-button-secondary" onClick={() => setConfirm(false)}>Conservar reserva</Action>
        <Action className="agenda-button-danger" disabled={cancel.isPending} onClick={() => cancel.mutate(row.version, { onSuccess: () => setConfirm(false) })}>Confirmar cancelación</Action>
      </section>}
      {cancel.isSuccess && <p className="agenda-confirmation" role="status">Reserva cancelada</p>}
    </section>}
  </>
}
export function DetalleReserva(props: DetalleReservaProps) { return <AgendaBoundary><View {...props} /></AgendaBoundary> }
