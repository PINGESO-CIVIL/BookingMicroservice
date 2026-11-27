import { useState } from 'react'
import { permits, type AgendaContext } from '../../api/agenda'
import { Action, AgendaBoundary, Notice } from '../shared'
import { useListaRecursos } from './use-lista-recursos'

export interface ListaRecursosProps extends AgendaContext {
  onOpen?: (id: string) => void
  onCreate?: () => void
}
function View(props: ListaRecursosProps) {
  const [enabled, setEnabled] = useState(true)
  const query = useListaRecursos(props, enabled)
  return (
    <>
      <div className="agenda-toolbar agenda-toolbar-heading">
        <label>
          Estado
          <select
            value={String(enabled)}
            onChange={(e) => setEnabled(e.target.value === 'true')}
          >
            <option value="true">Habilitados</option>
            <option value="false">Deshabilitados</option>
          </select>
        </label>
        {permits(props, 'recursos:administrar') && props.onCreate && (
          <Action onClick={props.onCreate}>
            <span aria-hidden="true">＋</span> Nuevo recurso
          </Action>
        )}
      </div>
      <Notice error={query.error} retry={() => void query.refetch()} />
      {query.isPending && (
        <p className="agenda-message" role="status">
          Cargando recursos…
        </p>
      )}
      {query.isSuccess && !query.data.pages.some((page) => page.items.length) && (
        <p className="agenda-message">No hay recursos para este estado.</p>
      )}
      <div className="agenda-resource-grid">
        {query.data?.pages
          .flatMap((page) => page.items)
          .map((resource) => (
            <article className="agenda-card agenda-resource-card" key={resource.id}>
              <div className="agenda-row">
                <span className="agenda-resource-icon" aria-hidden="true">
                  <svg
                    width="18"
                    height="18"
                    viewBox="0 0 18 18"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="1.6"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <rect x="2.5" y="2.5" width="5.5" height="5.5" rx="1" />
                    <rect x="10" y="2.5" width="5.5" height="5.5" rx="1" />
                    <rect x="2.5" y="10" width="5.5" height="5.5" rx="1" />
                    <rect x="10" y="10" width="5.5" height="5.5" rx="1" />
                  </svg>
                </span>
                <span
                  className="agenda-status"
                  data-status={resource.enabled ? 'confirmed' : 'cancelled'}
                >
                  {resource.enabled ? 'Habilitado' : 'Deshabilitado'}
                </span>
              </div>
              <h2>{resource.name}</h2>
              <p>{resource.timeZone}</p>
              {permits(props, 'recursos:administrar') && props.onOpen && (
                <Action
                  className="agenda-button-secondary"
                  onClick={() => props.onOpen?.(resource.id)}
                >
                  Administrar recurso <span aria-hidden="true">→</span>
                </Action>
              )}
            </article>
          ))}
      </div>
      {query.hasNextPage && (
        <Action
          onClick={() => void query.fetchNextPage()}
          disabled={query.isFetchingNextPage}
        >
          Cargar más
        </Action>
      )}
    </>
  )
}
export function ListaRecursos(props: ListaRecursosProps) {
  return (
    <AgendaBoundary>
      <View {...props} />
    </AgendaBoundary>
  )
}
