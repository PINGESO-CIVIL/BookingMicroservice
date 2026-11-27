import { useState, type FormEvent } from 'react'
import { AgendaError, permits, type AgendaContext } from '../../api/agenda'
import { formatDateTime } from '../../dates'
import { Action, AgendaBoundary, Notice } from '../shared'
import { useEditorRecurso } from './use-editor-recurso'

export interface EditorRecursoProps extends AgendaContext {
  id?: string
  onBack?: () => void
  onSaved?: (id: string) => void
}
const days = ['Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado', 'Domingo']
function View(props: EditorRecursoProps) {
  const { query, form, rules, save, closures } = useEditorRecurso(props, props.id)
  const [tab, setTab] = useState<'General' | 'Horarios' | 'Bloqueos'>('General')
  const [withSchedule, setWithSchedule] = useState(false)
  const [deleteId, setDeleteId] = useState<string>()
  const row = query.data
  const conflict =
    save.error instanceof AgendaError && save.error.code?.endsWith('VERSION_CONFLICT')
  const mayEdit = permits(props, 'recursos:administrar')
  const schedule = tab === 'Horarios' || (!props.id && withSchedule)
  function submitClosure(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const fields = new FormData(event.currentTarget)
    closures.mutate({
      start: String(fields.get('start')),
      end: String(fields.get('end')),
      reason: String(fields.get('reason')),
    })
  }
  return (
    <>
      {props.onBack && (
        <Action className="agenda-button-secondary" onClick={props.onBack}>
          <span aria-hidden="true">←</span> Volver a recursos
        </Action>
      )}
      <Notice error={query.error} retry={() => void query.refetch()} />
      {props.id && query.isPending && <p role="status">Cargando recurso…</p>}
      {row && (
        <div
          className="agenda-row agenda-editor-tabs"
          role="tablist"
          aria-label="Configuración del recurso"
        >
          {(['General', 'Horarios', 'Bloqueos'] as const).map((value) => (
            <Action
              key={value}
              role="tab"
              aria-selected={tab === value}
              onClick={() => {
                setTab(value)
                save.reset()
                form.reset()
              }}
            >
              {value}
            </Action>
          ))}
        </div>
      )}
      {tab === 'Bloqueos' && row ? (
        <section className="agenda-card">
          <h2>Bloqueos</h2>
          <Notice error={closures.error} />
          <div className="agenda-closure-list">
            {row.closures.map((closure) => (
              <article key={closure.id} className="agenda-closure-item">
                <p>
                  {closure.reason} · {formatDateTime(closure.startAt, row.timeZone)}
                </p>
                <div className="agenda-closure-actions">
                  {deleteId === closure.id ? (
                    <>
                      <Action
                        className="agenda-button-secondary"
                        onClick={() => setDeleteId(undefined)}
                      >
                        Conservar bloqueo
                      </Action>
                      <Action
                        className="agenda-button-danger"
                        disabled={!mayEdit || closures.isPending}
                        onClick={() =>
                          closures.mutate(
                            { id: closure.id },
                            { onSuccess: () => setDeleteId(undefined) },
                          )
                        }
                      >
                        Confirmar eliminación
                      </Action>
                    </>
                  ) : (
                    <Action
                      className="agenda-button-secondary agenda-schedule-remove"
                      onClick={() => setDeleteId(closure.id)}
                    >
                      Eliminar bloqueo
                    </Action>
                  )}
                </div>
              </article>
            ))}
          </div>
          <form onSubmit={submitClosure} className="agenda-row agenda-closure-form">
            <label>
              Inicio
              <input type="datetime-local" name="start" required />
            </label>
            <label>
              Término
              <input type="datetime-local" name="end" required />
            </label>
            <label>
              Motivo
              <input name="reason" required />
            </label>
            <Action type="submit" disabled={!mayEdit || closures.isPending}>
              Agregar bloqueo
            </Action>
          </form>
        </section>
      ) : (
        (!props.id || row) && (
          <form
            className="agenda-card agenda-resource-editor"
            onSubmit={form.handleSubmit((fields) =>
              save.mutate({
                fields:
                  !props.id && !withSchedule ? { ...fields, weeklySchedule: [] } : fields,
                schedule: tab === 'Horarios',
              }),
            )}
          >
            <Notice
              error={save.error}
              retry={
                conflict
                  ? () => {
                      void query.refetch()
                      save.reset()
                    }
                  : undefined
              }
            />
            {tab === 'General' && (
              <>
                <h2>Información general</h2>
                <label>
                  Nombre
                  <input {...form.register('name')} required />
                </label>
                <label>
                  Zona horaria IANA
                  <input {...form.register('timeZone')} required />
                </label>
                <label>
                  Recurso habilitado
                  <input type="checkbox" {...form.register('enabled')} />
                </label>
              </>
            )}
            <h2>Políticas de reserva</h2>
            <div className="agenda-policy-grid">
              {(
                [
                  ['cancellationNoticeSeconds', 'Anticipación de cancelación (segundos)'],
                  ['minDurationSeconds', 'Duración mínima (segundos)'],
                  ['maxDurationSeconds', 'Duración máxima (segundos)'],
                  ['slotDurationSeconds', 'Duración de bloque (segundos)'],
                ] as const
              ).map(([name, label]) => (
                <label key={name}>
                  {label}
                  <input
                    type="number"
                    min="0"
                    required
                    {...form.register(('policy.' + name) as `policy.${typeof name}`, {
                      valueAsNumber: true,
                    })}
                  />
                </label>
              ))}
            </div>
            {!props.id && (
              <label>
                Configurar horario semanal ahora
                <input
                  type="checkbox"
                  checked={withSchedule}
                  onChange={(e) => setWithSchedule(e.target.checked)}
                />
              </label>
            )}
            {schedule && (
              <section className="agenda-weekly-schedule">
                <h2>Horario semanal</h2>
                {rules.fields.map((rule, index) => (
                  <div key={rule.id} className="agenda-schedule-row">
                    <label>
                      Día
                      <select
                        {...form.register(`weeklySchedule.${index}.dayOfWeek`, {
                          valueAsNumber: true,
                        })}
                      >
                        {days.map((day, i) => (
                          <option key={day} value={i + 1}>
                            {day}
                          </option>
                        ))}
                      </select>
                    </label>
                    <label>
                      Apertura
                      <input
                        type="time"
                        required
                        {...form.register(`weeklySchedule.${index}.startLocalTime`)}
                      />
                    </label>
                    <label>
                      Cierre
                      <input
                        type="time"
                        required
                        {...form.register(`weeklySchedule.${index}.endLocalTime`)}
                      />
                    </label>
                    <Action
                      className="agenda-button-secondary agenda-schedule-remove"
                      onClick={() => rules.remove(index)}
                    >
                      Eliminar franja
                    </Action>
                  </div>
                ))}
                <Action
                  className="agenda-button-secondary agenda-schedule-add"
                  onClick={() =>
                    rules.append({
                      dayOfWeek: 1,
                      startLocalTime: '08:00',
                      endLocalTime: '18:00',
                    })
                  }
                >
                  <span aria-hidden="true">＋</span> Agregar franja
                </Action>
              </section>
            )}
            {save.data && (
              <div className="agenda-save-notice">
                <span className="agenda-save-icon" aria-hidden="true">
                  ✓
                </span>
                <div className="agenda-save-message" role="status">
                  <strong>
                    {!props.id
                      ? 'Recurso creado'
                      : save.variables?.schedule
                        ? 'Horario actualizado'
                        : 'Cambios guardados'}
                  </strong>
                  <p>
                    {props.id && props.onBack
                      ? 'Puedes volver al listado para seguir administrando tus recursos.'
                      : 'La operación se completó correctamente.'}
                  </p>
                </div>
                {props.id
                  ? props.onBack && (
                      <Action className="agenda-button-secondary" onClick={props.onBack}>
                        Ir a Recursos <span aria-hidden="true">→</span>
                      </Action>
                    )
                  : props.onSaved && (
                      <Action
                        className="agenda-button-secondary"
                        onClick={() => props.onSaved?.(save.data.id)}
                      >
                        Ver recurso <span aria-hidden="true">→</span>
                      </Action>
                    )}
              </div>
            )}
            <div className="agenda-editor-actions">
              <Action
                className="agenda-button-secondary"
                onClick={() => {
                  form.reset()
                  save.reset()
                }}
              >
                Descartar cambios
              </Action>
              <Action type="submit" disabled={!mayEdit || save.isPending || !!conflict}>
                {props.id ? 'Guardar cambios' : 'Crear recurso'}
              </Action>
            </div>
          </form>
        )
      )}
    </>
  )
}
export function EditorRecurso(props: EditorRecursoProps) {
  return (
    <AgendaBoundary>
      <View key={props.id ?? 'new'} {...props} />
    </AgendaBoundary>
  )
}
