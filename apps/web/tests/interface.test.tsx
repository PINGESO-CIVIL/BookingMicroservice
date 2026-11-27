// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, within } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'
import { DetalleReserva, EditorRecurso, FormularioReserva, ListaRecursos, ListaReservas } from '../src/componentes'
import { confirmed, cancelled, reservations, resources, resourceDetail } from '../src/api/fixtures'

const context = { baseUrl: 'http://localhost', token: 'demo-user', now: '2026-11-04T16:00:00Z' }
const json = (data: unknown, status = 200) => new Response(JSON.stringify(data), {status, headers: {'Content-Type': 'application/json'}})
afterEach(() => { cleanup(); vi.unstubAllGlobals() })

it('monta reservas sin router y entrega la selección al host', async () => {
  const open = vi.fn()
  vi.stubGlobal('fetch', vi.fn((req: Request) => Promise.resolve(json({ items: req.url.includes('/recursos') ? resources : reservations, nextCursor: null }))))
  render(<ListaReservas {...context} onOpen={open} />)
  const actions = await screen.findAllByRole('button', { name: 'Ver detalle' })
  expect(actions.length).toBe(reservations.length)
  fireEvent.click(actions[0]!)
  expect(open).toHaveBeenCalledWith(reservations[0]!.id)
})
it('carga el catálogo de recursos y cambia el filtro de habilitación', async () => {
  const fetch = vi.fn((req: Request) => Promise.resolve(json({ items: resources.filter(r => r.enabled === (new URL(req.url).searchParams.get('enabled') === 'true')), nextCursor: null })))
  vi.stubGlobal('fetch', fetch)
  render(<ListaRecursos {...context} />)
  await screen.findByText(resources.find(r => r.enabled)!.name)
  fireEvent.change(screen.getByLabelText('Estado'), { target: { value: 'false' } })
  await screen.findByText(resources.find(r => !r.enabled)!.name)
})
it('confirma la cancelación desde el detalle', async () => {
  vi.stubGlobal('fetch', vi.fn((req: Request) => Promise.resolve(json(req.method === 'POST' ? cancelled : confirmed))))
  render(<DetalleReserva {...context} id={confirmed.id} />)
  fireEvent.click(await screen.findByRole('button', { name: 'Cancelar reserva' }))
  fireEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Confirmar cancelación' }))
  await screen.findByText('Reserva cancelada')
})
it('selecciona un bloque y confirma desde el formulario', async () => {
  const slot = { startAt: '2026-11-05T12:00:00Z', endAt: '2026-11-05T13:00:00Z' }
  vi.stubGlobal('fetch', vi.fn((req: Request) => Promise.resolve(req.method === 'POST'
    ? json({ ...confirmed, ...slot }, 201)
    : json(req.url.includes('disponibilidades') ? { availableSlots: [slot] } : { items: resources, nextCursor: null }))))
  render(<FormularioReserva {...context} />)
  await screen.findByText(resources[0]!.name)
  fireEvent.change(screen.getByLabelText('Recurso'), { target: { value: resources[0]!.id } })
  fireEvent.change(screen.getByLabelText('Fecha'), { target: { value: '2026-11-05' } })
  fireEvent.click(await screen.findByRole('button', { name: /09:00/ }))
  fireEvent.click(screen.getByRole('button', { name: 'Confirmar reserva' }))
  await screen.findByText('Reserva confirmada')
})
it('edita un recurso y entrega el resultado', async () => {
  const row = resourceDetail(resources[0]!.id)
  const fetch = vi.fn((req: Request) => Promise.resolve(json(req.method === 'PATCH' ? { ...row, name: 'Cancha' } : row)))
  vi.stubGlobal('fetch', fetch)
  render(<EditorRecurso {...context} id={row.id} />)
  fireEvent.change(await screen.findByLabelText('Nombre'), { target: { value: 'Cancha' } })
  fireEvent.click(screen.getByRole('button', { name: 'Guardar cambios' }))
  await screen.findByText('Cambios guardados')
  const patch = fetch.mock.calls.find(([req]) => req.method === 'PATCH')?.[0]
  expect(await patch?.json()).toMatchObject({ name: 'Cancha', expectedVersion: row.version })
})
 
it('aplica filtros, pagina y permite reintentar el listado', async () => {
  const fetch = vi.fn((req: Request) => {
    const url = new URL(req.url)
    const cursor = url.searchParams.get('cursor')
    if (url.pathname.endsWith('/recursos')) return Promise.resolve(json({ items: resources, nextCursor: null }))
    if (url.searchParams.get('status') && !cursor) return Promise.resolve(json({ error: { message: 'Error de lista' } }, 503))
    return Promise.resolve(json({ items: cursor ? [] : [confirmed], nextCursor: cursor ? null : 'page-2' }))
  })
  vi.stubGlobal('fetch', fetch)
  const create = vi.fn()
  render(<ListaReservas {...context} onCreate={create} />)
  fireEvent.click(screen.getByRole('button', { name: 'Nueva reserva' }))
  expect(create).toHaveBeenCalledOnce()
  fireEvent.click(await screen.findByRole('button', { name: 'Cargar más' }))
  await vi.waitFor(() => expect(fetch.mock.calls.some(([req]) => req.url.includes('cursor='))).toBe(true))
  fireEvent.change(screen.getByLabelText('Estado'), { target: { value: 'cancelled' } })
  await screen.findByText('Error de lista')
  fireEvent.click(screen.getByRole('button', { name: 'Reintentar' }))
  fireEvent.change(screen.getByLabelText('Recurso'), { target: { value: resources[0]!.id } })
  fireEvent.change(screen.getByLabelText('Desde'), { target: { value: '2026-11-01' } })
  fireEvent.change(screen.getByLabelText('Hasta'), { target: { value: '2026-11-30' } })
  fireEvent.click(screen.getByRole('button', { name: 'Limpiar filtros' }))
  expect(screen.getByLabelText<HTMLInputElement>('Desde').value).toBe('')
})

it('emite navegación y limita la edición en el detalle', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(json(confirmed)))
  const edit = vi.fn(), back = vi.fn()
  const { rerender } = render(<DetalleReserva {...context} id={confirmed.id} onEdit={edit} onBack={back} />)
  fireEvent.click(await screen.findByRole('button', { name: 'Modificar reserva' }))
  expect(edit).toHaveBeenCalledWith(confirmed.id)
  fireEvent.click(screen.getByRole('button', { name: 'Volver a mis reservas' }))
  expect(back).toHaveBeenCalledOnce()
  fireEvent.click(screen.getByRole('button', { name: 'Cancelar reserva' }))
  fireEvent.click(screen.getByRole('button', { name: 'Conservar reserva' }))
  expect(screen.queryByRole('dialog')).toBeNull()
  rerender(<DetalleReserva {...context} id={confirmed.id} permissions={[]} onEdit={edit} />)
  expect(screen.queryByRole('button', { name: 'Modificar reserva' })).toBeNull()
})

it('edita un bloque y conserva el control de versión', async () => {
  const slot = { startAt: '2026-11-10T23:00:00Z', endAt: '2026-11-11T00:00:00Z' }
  const fetch = vi.fn((req: Request) => {
    const path = new URL(req.url).pathname
    const answer = path.endsWith('/disponibilidades') ? { availableSlots: [slot] }
      : path.endsWith('/recursos') ? { items: resources, nextCursor: null } : confirmed
    return Promise.resolve(json(req.method === 'PATCH' ? { ...confirmed, ...slot } : answer))
  })
  vi.stubGlobal('fetch', fetch)
  const saved = vi.fn(), back = vi.fn()
  render(<FormularioReserva {...context} id={confirmed.id} onSaved={saved} onBack={back} />)
  fireEvent.click(await screen.findByRole('button', { name: '20:00 – 21:00' }))
  fireEvent.click(screen.getByRole('button', { name: 'Guardar cambios' }))
  await screen.findByText('Reserva modificada')
  fireEvent.click(screen.getByRole('button', { name: 'Ver reserva' }))
  expect(saved).toHaveBeenCalledWith(confirmed.id)
  fireEvent.click(screen.getByRole('button', { name: 'Volver a mis reservas' }))
  expect(back).toHaveBeenCalledOnce()
  const patch = fetch.mock.calls.find(([req]) => req.method === 'PATCH')?.[0]
  expect(await patch?.json()).toMatchObject({ expectedVersion: confirmed.version, startAt: slot.startAt })
})

it('actualiza horarios y gestiona bloqueos del recurso', async () => {
  const row = resourceDetail(resources[0]!.id)
  const closure = { ...row.closures[0]!, id: 'nuevo', reason: 'Prueba de bloqueo' }
  const fetch = vi.fn((req: Request) => Promise.resolve(req.method === 'DELETE' ? new Response(null, { status: 204 })
    : json(req.method === 'POST' ? closure : row)))
  vi.stubGlobal('fetch', fetch)
  render(<EditorRecurso {...context} id={row.id} />)
  fireEvent.click(await screen.findByRole('tab', { name: 'Horarios' }))
  fireEvent.click(screen.getByRole('button', { name: 'Agregar franja' }))
  fireEvent.click(screen.getAllByRole('button', { name: 'Eliminar franja' }).at(-1)!)
  fireEvent.click(screen.getByRole('button', { name: 'Guardar cambios' }))
  await screen.findByText('Cambios guardados')
  fireEvent.click(screen.getByRole('tab', { name: 'Bloqueos' }))
  fireEvent.change(screen.getByLabelText('Inicio'), { target: { value: '2026-11-10T11:00' } })
  fireEvent.change(screen.getByLabelText('Término'), { target: { value: '2026-11-10T12:00' } })
  fireEvent.change(screen.getByLabelText('Motivo'), { target: { value: closure.reason } })
  fireEvent.click(screen.getByRole('button', { name: 'Agregar bloqueo' }))
  const created = await screen.findByText(/Prueba de bloqueo/)
  fireEvent.click(within(created.closest('article')!).getByRole('button', { name: 'Eliminar bloqueo' }))
  fireEvent.click(screen.getByRole('button', { name: 'Confirmar eliminación' }))
  await vi.waitFor(() => expect(screen.queryByText(/Prueba de bloqueo/)).toBeNull())
  expect(fetch.mock.calls.map(([req]) => req.method)).toEqual(expect.arrayContaining(['PUT', 'POST', 'DELETE']))
})

it('crea recursos con horario opcional y permite salir', async () => {
  const row = resourceDetail(resources[0]!.id), goBack = vi.fn(), open = vi.fn()
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(json(row, 201)))
  render(<EditorRecurso {...context} onSaved={open} onBack={goBack} />)
  fireEvent.change(screen.getByLabelText('Nombre'), { target: { value: 'Cancha' } })
  fireEvent.click(screen.getByLabelText('Configurar horario semanal ahora'))
  fireEvent.click(screen.getByRole('button', { name: 'Agregar franja' }))
  fireEvent.click(screen.getByRole('button', { name: 'Crear recurso' }))
  fireEvent.click(await screen.findByRole('button', { name: 'Ver recurso' }))
  expect(open).toHaveBeenCalledWith(row.id)
  fireEvent.click(screen.getByRole('button', { name: 'Volver a recursos' }))
  expect(goBack).toHaveBeenCalledOnce()
})
