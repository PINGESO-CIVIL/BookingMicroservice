import type { AgendaContext } from '../../api/agenda'
import { AgendaError } from '../../api/agenda'
import { formatDateTime, localParts } from '../../dates'
import { Action, AgendaBoundary, Notice } from '../shared'
import { useFormularioReserva } from './use-formulario-reserva'

export interface FormularioReservaProps extends AgendaContext {
  id?: string
  onSaved?: (id: string) => void
  onBack?: () => void
}
function View(props: FormularioReservaProps) {
  const state = useFormularioReserva(props, props.id)
  const {
    form,
    list,
    existing,
    resourceId,
    date,
    zone,
    slot,
    slots,
    items,
    save,
    locked,
    options,
    changed,
  } = state
  const conflict =
    save.error instanceof AgendaError && save.error.code?.endsWith('VERSION_CONFLICT')
  function selectResource(id: string) {
    form.setValue('resourceId', id)
    form.setValue('startAt', '')
    form.setValue('endAt', '')
    save.reset()
  }
  return (
    <div className="agenda-card">
      {props.onBack && (
        <Action className="agenda-button-secondary" onClick={props.onBack}>
          <span aria-hidden="true">←</span> Volver a mis reservas
        </Action>
      )}
      <Notice
        error={list.error ?? existing.error}
        retry={() => {
          void list.refetch()
          if (props.id) void existing.refetch()
        }}
      />
      {(list.isPending || (!!props.id && existing.isPending)) && (
        <p role="status">Cargando…</p>
      )}
      {locked && <p>Esta reserva ya no admite cambios.</p>}
      {save.data ? (
        <div className="agenda-save-notice">
          <span className="agenda-save-icon" aria-hidden="true">
            ✓
          </span>
          <div className="agenda-save-message" role="status">
            <strong>{props.id ? 'Reserva modificada' : 'Reserva confirmada'}</strong>
            <p>{formatDateTime(save.data.startAt, zone)}</p>
            {!props.id && <p>Se simuló el envío del correo de confirmación.</p>}
          </div>
          {props.onSaved && (
            <Action
              className="agenda-button-secondary"
              onClick={() => props.onSaved?.(save.data.id)}
            >
              Ver reserva
            </Action>
          )}
        </div>
      ) : (
        <form onSubmit={form.handleSubmit(() => save.mutate())}>
          <div className="agenda-fields">
            <label>
              Recurso
              <select
                value={resourceId}
                onChange={(e) => selectResource(e.target.value)}
                disabled={locked}
              >
                <option value="">Selecciona un recurso</option>
                {items.map((r) => (
                  <option key={r.id} value={r.id}>
                    {r.name}
                  </option>
                ))}
              </select>
            </label>
            {list.hasNextPage && (
              <Action
                onClick={() => void list.fetchNextPage()}
                disabled={list.isFetchingNextPage}
              >
                Cargar más recursos
              </Action>
            )}
            <label>
              Fecha
              <input
                type="date"
                value={date}
                disabled={locked}
                onChange={(e) => {
                  form.setValue('date', e.target.value)
                  form.setValue('startAt', '')
                  form.setValue('endAt', '')
                  save.reset()
                }}
              />
            </label>
          </div>
          {!locked && resourceId && (
            <section className="agenda-slots">
              <h2>Horarios disponibles</h2>
              <Notice error={slots.error} retry={() => void slots.refetch()} />
              {slots.isPending && <p role="status">Consultando disponibilidad…</p>}
              {slots.isSuccess && !options.length && (
                <p>No hay horarios disponibles para esta fecha.</p>
              )}
              <div className="agenda-slot-grid">
                {options.map((window) => (
                  <Action
                    className="agenda-button-secondary"
                    key={window.startAt}
                    aria-pressed={slot.startAt === window.startAt}
                    onClick={() => {
                      form.setValue('startAt', window.startAt)
                      form.setValue('endAt', window.endAt)
                      save.reset()
                    }}
                  >
                    {localParts(window.startAt, zone).time} –{' '}
                    {localParts(window.endAt, zone).time}
                  </Action>
                ))}
              </div>
            </section>
          )}
          <Notice
            error={save.error}
            retry={
              conflict
                ? () => {
                    void existing.refetch()
                    save.reset()
                  }
                : undefined
            }
          />
          <div className="agenda-form-footer">
            <Action
              type="submit"
              disabled={
                locked ||
                !changed ||
                !slot.startAt ||
                save.isPending ||
                !!conflict ||
                (!!props.id && !existing.data)
              }
            >
              {props.id ? 'Guardar cambios' : 'Confirmar reserva'}
            </Action>
          </div>
        </form>
      )}
    </div>
  )
}
export function FormularioReserva(props: FormularioReservaProps) {
  return (
    <AgendaBoundary>
      <View key={props.id ?? 'new'} {...props} />
    </AgendaBoundary>
  )
}
