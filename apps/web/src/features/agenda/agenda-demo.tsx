import { useState } from 'react'
import { NavLink, Navigate, Route, Routes, useNavigate, useParams } from 'react-router'
import type { AgendaContext } from '../../api/agenda'
import {
  DetalleReserva,
  EditorRecurso,
  FormularioReserva,
  ListaRecursos,
  ListaReservas,
} from '../../componentes'

const scenarios = ['normal', 'empty', 'paginated', 'error', 'unauthorized', 'forbidden']
const writes = ['normal', 'conflict', 'version', 'deadline', 'policy']
const scenarioLabels: Record<string, string> = {
  normal: 'Normal',
  empty: 'Sin resultados',
  paginated: 'Con paginación',
  error: 'Servicio caído',
  unauthorized: 'Sin sesión',
  forbidden: 'Sin permiso',
  conflict: 'Conflicto',
  version: 'Versión anterior',
  deadline: 'Plazo vencido',
  policy: 'Política inválida',
}

function PageHeading({ title, description }: { title: string; description: string }) {
  return (
    <div className="agenda-page-heading">
      <h1>{title}</h1>
      <p>{description}</p>
    </div>
  )
}

function Detail({ context }: { context: AgendaContext }) {
  const { id } = useParams()
  const navigate = useNavigate()
  return (
    <DetalleReserva
      {...context}
      id={id!}
      onBack={() => navigate('/reservas')}
      onEdit={(value) => navigate('/reservas/' + value + '/editar')}
    />
  )
}
function Form({ context }: { context: AgendaContext }) {
  const { id } = useParams()
  const navigate = useNavigate()
  return (
    <FormularioReserva
      {...context}
      id={id}
      onBack={() => navigate('/reservas')}
      onSaved={(value) => navigate('/reservas/' + value)}
    />
  )
}
function Editor({ context }: { context: AgendaContext }) {
  const { id } = useParams()
  const navigate = useNavigate()
  return (
    <EditorRecurso
      {...context}
      id={id}
      onBack={() => navigate('/recursos')}
      onSaved={(value) => navigate('/recursos/' + value)}
    />
  )
}
export function AgendaDemo() {
  const navigate = useNavigate()
  const [admin, setAdmin] = useState(false)
  const [scenario, setScenario] = useState('normal')
  const [writeScenario, setWriteScenario] = useState('normal')
  const context: AgendaContext = {
    demoRole: admin ? 'admin' : 'user',
    baseUrl: import.meta.env.VITE_API_BASE_URL ?? '/',
    now: '2026-11-04T16:00:00Z',
    scenario,
    writeScenario,
  }
  function changeRole(next: boolean) {
    setAdmin(next)
    navigate(next ? '/recursos' : '/reservas')
  }
  return (
    <div className="agenda-demo">
      <aside className="agenda-sidebar">
        <div className="agenda-brand">
          <span className="agenda-brand-mark" aria-hidden="true">
            <svg
              width="18"
              height="18"
              viewBox="0 0 18 18"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.75"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <rect x="2.25" y="3.25" width="13.5" height="12.5" rx="2" />
              <path d="M2.25 7.25h13.5M6 1.75v3M12 1.75v3" />
            </svg>
          </span>
          <span>Agenda</span>
        </div>
        <nav className="agenda-nav" aria-label="Navegación de Agenda">
          <NavLink to="/reservas">
            <svg
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.6"
              strokeLinecap="round"
              strokeLinejoin="round"
              aria-hidden="true"
            >
              <rect x="2" y="2.75" width="12" height="11.25" rx="1.75" />
              <path d="M2 6.25h12M5.25 1.5v2.5M10.75 1.5v2.5" />
            </svg>
            Mis reservas
          </NavLink>
          {admin && (
            <>
              <NavLink to="/administracion/reservas">
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 16 16"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.6"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  aria-hidden="true"
                >
                  <path d="M2.5 4h11M2.5 8h11M2.5 12h7" />
                </svg>
                Reservas realizadas
              </NavLink>
              <NavLink to="/recursos">
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 16 16"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.6"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  aria-hidden="true"
                >
                  <rect x="2" y="2" width="5" height="5" rx="1" />
                  <rect x="9" y="2" width="5" height="5" rx="1" />
                  <rect x="2" y="9" width="5" height="5" rx="1" />
                  <rect x="9" y="9" width="5" height="5" rx="1" />
                </svg>
                Recursos
              </NavLink>
            </>
          )}
        </nav>
        <section className="agenda-demo-settings" aria-label="Configuración de la demo">
          <h2>Modo de prueba</h2>
          <div className="agenda-role" role="group" aria-label="Rol de demostración">
            <button type="button" aria-pressed={!admin} onClick={() => changeRole(false)}>
              Usuario
            </button>
            <button type="button" aria-pressed={admin} onClick={() => changeRole(true)}>
              Admin.
            </button>
          </div>
          <div className="agenda-scenarios">
            <label>
              Lectura
              <select value={scenario} onChange={(e) => setScenario(e.target.value)}>
                {scenarios.map((value) => (
                  <option key={value} value={value}>
                    {scenarioLabels[value]}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Escritura
              <select
                value={writeScenario}
                onChange={(e) => setWriteScenario(e.target.value)}
              >
                {writes.map((value) => (
                  <option key={value} value={value}>
                    {scenarioLabels[value]}
                  </option>
                ))}
              </select>
            </label>
          </div>
        </section>
      </aside>
      <main className="agenda-main">
        <div className="agenda-content">
          <Routes key={context.demoRole}>
            <Route
              path="/reservas"
              element={
                <>
                  <PageHeading
                    title="Mis reservas"
                    description="Consulta tus reservas y crea una nueva cuando necesites un espacio."
                  />
                  <ListaReservas
                    {...context}
                    onOpen={(id) => navigate('/reservas/' + id)}
                    onCreate={() => navigate('/reservas/nueva')}
                  />
                </>
              }
            />
            <Route
              path="/reservas/nueva"
              element={
                <>
                  <PageHeading
                    title="Nueva reserva"
                    description="Elige un recurso y selecciona un horario disponible."
                  />
                  <Form context={context} />
                </>
              }
            />
            <Route
              path="/reservas/:id"
              element={
                <>
                  <PageHeading
                    title="Detalle de reserva"
                    description="Consulta el estado, el horario y las opciones de tu reserva."
                  />
                  <Detail context={context} />
                </>
              }
            />
            <Route
              path="/reservas/:id/editar"
              element={
                <>
                  <PageHeading
                    title="Modificar reserva"
                    description="Selecciona un nuevo horario o cambia el recurso de la reserva."
                  />
                  <Form context={context} />
                </>
              }
            />
            <Route
              path="/administracion/reservas"
              element={
                admin ? (
                  <>
                    <PageHeading
                      title="Reservas realizadas"
                      description="Consulta las reservas de los recursos que administras."
                    />
                    <ListaReservas {...context} admin />
                  </>
                ) : (
                  <Navigate to="/reservas" replace />
                )
              }
            />
            <Route
              path="/recursos"
              element={
                admin ? (
                  <>
                    <PageHeading
                      title="Recursos"
                      description="Revisa y configura los espacios disponibles para reservar."
                    />
                    <ListaRecursos
                      {...context}
                      onCreate={() => navigate('/recursos/nuevo')}
                      onOpen={(id) => navigate('/recursos/' + id)}
                    />
                  </>
                ) : (
                  <Navigate to="/reservas" replace />
                )
              }
            />
            <Route
              path="/recursos/nuevo"
              element={
                admin ? (
                  <>
                    <PageHeading
                      title="Crear recurso"
                      description="Define los datos, el horario y las políticas del nuevo recurso."
                    />
                    <Editor context={context} />
                  </>
                ) : (
                  <Navigate to="/reservas" replace />
                )
              }
            />
            <Route
              path="/recursos/:id"
              element={
                admin ? (
                  <>
                    <PageHeading
                      title="Administrar recurso"
                      description="Edita la configuración, los horarios y los bloqueos del recurso."
                    />
                    <Editor context={context} />
                  </>
                ) : (
                  <Navigate to="/reservas" replace />
                )
              }
            />
            <Route path="*" element={<Navigate to="/reservas" replace />} />
          </Routes>
        </div>
      </main>
    </div>
  )
}
