// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { createMemoryHistory, createRouter, type Router } from 'vue-router'
import App from '../src/App.vue'
import { router as applicationRouter } from '../src/router'
import { adminReservations, cancelled, confirmed, expired, past, reservations, resourceDetail, resources } from '../src/api/fixtures'
import { configuredAvailability } from '../src/api/availability'
import { DEMO_NOW, refreshKey, resetScenarios } from '../src/demo'

// Exercise the real views, controls, router and HTTP client; replace only transport.
let wrapper: VueWrapper, router: Router
let failure: { path: string; method?: string; code?: string; remaining: number } | undefined
let empty = false, paginate = false
let calls: { path: string; method: string; body: any; headers: Record<string, string> }[]
const clone = <T,>(value: T): T => JSON.parse(JSON.stringify(value))
const response = (value: unknown, status = 200) => new Response(status === 204 ? null : JSON.stringify(value), { status })
Object.defineProperty(HTMLDialogElement.prototype, 'showModal', { configurable: true, value: function (this: HTMLDialogElement) { this.open = true } })
async function settle() { await flushPromises(); await flushPromises() }
async function click(text: string, root: ParentNode = document) {
  const button = [...root.querySelectorAll<HTMLButtonElement>('button')].find(b => b.textContent?.trim() === text)
  expect(button, `Botón ${text}`).toBeTruthy()
  expect(button!.disabled, `Botón habilitado ${text}`).toBe(false)
  button!.click(); await settle()
}
async function input(label: string, value: string) {
  const control = [...document.querySelectorAll('label')].find(l => l.textContent?.trim().startsWith(label))?.control
  expect(control, `Control asociado a ${label}`).toBeTruthy()
  const field = wrapper.find(`[id="${control!.id}"]`)
  await field.setValue(value); await settle()
}
async function submit() { await wrapper.get('form').trigger('submit'); await settle() }
async function open(path: string) {
  router = createRouter({ history: createMemoryHistory(), routes: applicationRouter.options.routes })
  await router.push(path); await router.isReady()
  wrapper = mount(App, { attachTo: document.body, global: { plugins: [router] } })
  await settle()
}
beforeEach(() => {
  resetScenarios(); calls = []; failure = undefined; empty = false; paginate = false
  vi.stubGlobal('fetch', vi.fn(async (path: string, options: RequestInit) => {
    const url = new URL(path, 'http://localhost'), method = options.method ?? 'GET'
    const body = options.body ? JSON.parse(String(options.body)) : undefined
    calls.push({ path: url.pathname + url.search, method, body, headers: options.headers as Record<string, string> })
    if (failure?.remaining && url.pathname.endsWith(failure.path) && (!failure.method || method === failure.method)) {
      failure.remaining--
      return response({ error: { code: failure.code ?? 'SERVICE_UNAVAILABLE', message: 'Error de prueba; vuelve a intentarlo.' } }, 409)
    }
    if (url.pathname.endsWith('/disponibilidades')) {
      const id = url.pathname.split('/')[4]!
      return response(configuredAvailability(resourceDetail(id), reservations, url.searchParams.get('from')!, url.searchParams.get('to')!, DEMO_NOW))
    }
    if (url.pathname.endsWith('/bloqueos') && method === 'POST') return response({ ...resourceDetail('00000000-0000-4000-8000-100000000001').closures[0], ...body }, 201)
    if (method === 'DELETE') return response(null, 204)
    if (url.pathname.endsWith('/horarios')) return response({ ...resourceDetail('00000000-0000-4000-8000-100000000001'), ...body, version: 2 })
    if (url.pathname.endsWith('/recursos')) {
      if (method === 'POST') return response({ ...resourceDetail('00000000-0000-4000-8000-100000000001'), ...body }, 201)
      const enabled = url.searchParams.get('enabled')
      let items = resources.filter(r => enabled === null || r.enabled === (enabled === 'true'))
      if (empty) items = []
      return response({ items: paginate ? items.slice(url.searchParams.has('cursor') ? 1 : 0, url.searchParams.has('cursor') ? 2 : 1) : items, nextCursor: paginate && !url.searchParams.has('cursor') ? 'page-2' : null })
    }
    if (url.pathname.includes('/recursos/')) return response({ ...resourceDetail(url.pathname.split('/')[4]!), ...body })
    if (url.pathname.endsWith('/cancelaciones')) return response({ ...cancelled, id: confirmed.id })
    if (url.pathname.endsWith('/reservas')) {
      if (method === 'POST') return response({ ...confirmed, ...body }, 201)
      let items = url.pathname.includes('/admin/') ? adminReservations : reservations
      if (url.searchParams.get('status')) items = items.filter(r => r.status === url.searchParams.get('status'))
      if (url.searchParams.get('resourceId')) items = items.filter(r => r.resourceId === url.searchParams.get('resourceId'))
      if (empty) items = []
      return response({ items: paginate ? items.slice(url.searchParams.has('cursor') ? 1 : 0, url.searchParams.has('cursor') ? 2 : 1) : items, nextCursor: paginate && !url.searchParams.has('cursor') ? 'page-2' : null })
    }
    const id = url.pathname.split('/')[4]
    return response({ ...clone(reservations.find(r => r.id === id) ?? confirmed), ...body })
  }))
})
afterEach(() => { wrapper?.unmount(); document.body.innerHTML = ''; vi.restoreAllMocks(); vi.unstubAllGlobals() })

describe('reservas y navegación', () => {
  it('muestra el historial personal y cambia de rol sin ofrecer creación al administrador', async () => {
    await open('/reservas')
    expect(wrapper.text()).toContain(past.id)
    await click('Admin.')
    expect(router.currentRoute.value.path).toBe('/recursos')
    await wrapper.get('a[href="/administracion/reservas"]').trigger('click'); await settle()
    expect(wrapper.text()).toContain('user-002')
    expect(wrapper.text()).toContain('user-003')
    expect(document.body.textContent).not.toContain('Nueva reserva')
    await wrapper.get('button[aria-label="Ver detalle de 00000000-0000-4000-8000-200000000005"]').trigger('click'); await settle()
    expect(document.querySelector('dialog')?.textContent).toContain('user-002')
    expect(document.querySelector('dialog')?.textContent).not.toContain('Modificar reserva')
    await click('Cerrar', document.querySelector('dialog')!)
    expect(document.querySelector('dialog')).toBeNull()
    expect(calls.find(c => c.path.includes('/admin/reservas'))?.headers.Authorization).toBe('Bearer demo-admin')
    await click('Usuario'); expect(router.currentRoute.value.path).toBe('/reservas')
  })
  it('filtra, valida fechas, limpia filtros y carga otra página', async () => {
    paginate = true; await open('/reservas')
    await click('Cargar más'); expect(wrapper.text()).toContain(expired.id)
    expect(calls.some(c => c.path.includes('cursor=page-2'))).toBe(true)
    await input('Estado', 'cancelled'); expect(wrapper.text()).toContain(cancelled.id)
    await input('Recurso', '00000000-0000-4000-8000-100000000001'); expect(wrapper.text()).toContain('No encontramos coincidencias')
    await click('Limpiar filtros')
    await input('Fecha de inicio', '2026-11-10'); expect(wrapper.text()).toContain('Completa ambas fechas')
    await input('Fecha de término', '2026-11-09'); expect(wrapper.text()).toContain('El término debe ser posterior')
    await input('Fecha de término', '2026-11-12')
    expect(calls.some(c => c.path.includes('from=2026-11-10T03'))).toBe(true)
    await click('Limpiar filtros'); expect(wrapper.text()).not.toContain('El término debe ser posterior')
  })
  it('recupera una consulta fallida y permite comenzar desde el historial vacío', async () => {
    failure = { path: '/reservas', remaining: 1 }; await open('/reservas')
    expect(wrapper.text()).toContain('Error de prueba')
    empty = true; await click('Reintentar'); expect(wrapper.text()).toContain('Aún no tienes reservas')
    await click('Crear mi primera reserva'); expect(router.currentRoute.value.path).toBe('/reservas/nueva')
  })
  it('el administrador sin reservas no tiene acciones de creación', async () => {
    empty = true; await open('/administracion/reservas')
    expect(wrapper.text()).toContain('No hay reservas realizadas')
    expect(wrapper.text()).not.toContain('Crear mi primera reserva')
  })
  it('abre y cierra la navegación móvil', async () => {
    await open('/')
    await wrapper.get('[aria-label="Abrir navegación"]').trigger('click')
    expect(wrapper.get('.sidebar').classes()).toContain('sidebar-open')
    await wrapper.get('.scrim').trigger('click')
    expect(wrapper.get('.sidebar').classes()).not.toContain('sidebar-open')
    await wrapper.get('[aria-label="Abrir navegación"]').trigger('click')
    await wrapper.get('.mobile-close').trigger('click')
    expect(wrapper.get('.sidebar').classes()).not.toContain('sidebar-open')
  })
})

describe('crear y reprogramar', () => {
  async function chooseSlot() {
    await wrapper.get('.resource-option').trigger('click'); await click('Continuar')
    expect(wrapper.text()).toContain('Horarios disponibles')
    await wrapper.get('.slot-option').trigger('click'); await click('Continuar')
  }
  it('reserva un bloque configurado y envía responsable implícito e idempotencia', async () => {
    await open('/reservas/nueva')
    await wrapper.get('[aria-label="Buscar recurso"]').setValue('inexistente')
    expect(wrapper.text()).toContain('No hay recursos que coincidan')
    await wrapper.get('[aria-label="Buscar recurso"]').setValue('')
    await chooseSlot()
    expect(wrapper.text()).toContain('Revisa tu reserva')
    await click('Confirmar reserva')
    expect(wrapper.text()).toContain('Reserva confirmada')
    const sent = calls.find(c => c.method === 'POST' && c.path.endsWith('/reservas'))!
    expect(Date.parse(sent.body.endAt) - Date.parse(sent.body.startAt)).toBe(3600000)
    expect(sent.body).not.toHaveProperty('responsible')
    expect(sent.headers['Idempotency-Key']).toBeTruthy()
    await click('Ver reserva'); expect(router.currentRoute.value.path).toContain(confirmed.id)
  })
  it('conserva la clave al reintentar un fallo y permite volver a cambiar horario', async () => {
    await open('/reservas/nueva'); await chooseSlot()
    failure = { path: '/reservas', method: 'POST', remaining: 1 }
    await click('Confirmar reserva'); expect(wrapper.text()).toContain('Error de prueba')
    await click('Confirmar reserva')
    const attempts = calls.filter(c => c.method === 'POST')
    expect(attempts[0]!.headers['Idempotency-Key']).toBe(attempts[1]!.headers['Idempotency-Key'])
    await click('Volver a mis reservas'); expect(router.currentRoute.value.path).toBe('/reservas')
  })
  it('descarta el horario al cambiar de fecha y permite cambiar el recurso', async () => {
    await open('/reservas/nueva'); await wrapper.get('.resource-option').trigger('click'); await click('Continuar')
    await wrapper.get('.slot-option').trigger('click')
    await wrapper.get('[aria-label="Mes siguiente"]').trigger('click'); await settle()
    expect(wrapper.find('.duration-row').exists()).toBe(false)
    await wrapper.get('[aria-label="Mes anterior"]').trigger('click'); await settle()
    await input('Fecha seleccionada', '2026-11-11')
    expect(wrapper.text()).toContain('No hay horarios disponibles')
    await wrapper.get('[aria-label="2026-11-10"]').trigger('click'); await settle()
    expect(wrapper.findAll('.slot-option').length).toBeGreaterThan(0)
    await click('Cambiar recurso'); expect(wrapper.text()).toContain('Elige un recurso')
    await click('Cancelar'); expect(router.currentRoute.value.path).toBe('/reservas')
  })
  it('envía versión y solo los campos cambiados al reprogramar', async () => {
    await open(`/reservas/${confirmed.id}/editar`)
    expect(wrapper.text()).toContain('Conservar')
    await wrapper.get('.slot-option').trigger('click'); await click('Continuar'); await click('Guardar cambios')
    expect(wrapper.text()).toContain('Reserva actualizada')
    const sent = calls.find(c => c.method === 'PATCH')!
    expect(sent.body.expectedVersion).toBe(confirmed.version)
    expect(sent.body).not.toHaveProperty('resourceId')
  })
  it('recarga después de un conflicto de versión y no permite sobrescribirlo', async () => {
    await open(`/reservas/${confirmed.id}/editar`)
    await wrapper.get('.slot-option').trigger('click'); await click('Continuar')
    failure = { path: `/reservas/${confirmed.id}`, method: 'PATCH', remaining: 1, code: 'RESERVATION_VERSION_CONFLICT' }
    await click('Guardar cambios')
    expect([...document.querySelectorAll<HTMLButtonElement>('button')].find(b => b.textContent === 'Guardar cambios')?.disabled).toBe(true)
    await click('Actualizar información'); expect(wrapper.text()).toContain('Elige fecha y horario')
    await click('Descartar cambios'); expect(router.currentRoute.value.path).toBe(`/reservas/${confirmed.id}`)
  })
  it('bloquea la reprogramación de una reserva cancelada', async () => {
    await open(`/reservas/${cancelled.id}/editar`)
    expect(wrapper.text()).toContain('Esta reserva no admite cambios')
    await click('Volver a la reserva'); expect(router.currentRoute.value.path).toBe(`/reservas/${cancelled.id}`)
  })
})

describe('detalle y cancelación', () => {
  it.each([[cancelled.id, 'no puede reactivarse'], [past.id, 'La reserva ya comenzó'], [expired.id, 'El plazo para cancelar']])('protege %s de cambios', async (id, message) => {
    await open(`/reservas/${id}`); expect(wrapper.text()).toContain(message)
    expect(wrapper.text()).not.toContain('Modificar reserva')
  })
  it('cancela únicamente después de confirmar el diálogo', async () => {
    await open(`/reservas/${confirmed.id}`); await click('Cancelar reserva')
    expect(calls.some(c => c.path.endsWith('/cancelaciones'))).toBe(false)
    await click('Volver', document.querySelector('dialog')!)
    await click('Cancelar reserva'); await click('Sí, cancelar reserva', document.querySelector('dialog')!)
    expect(wrapper.text()).toContain('Reserva cancelada')
    expect(calls.find(c => c.path.endsWith('/cancelaciones'))?.body).toEqual({ expectedVersion: confirmed.version })
    expect(document.querySelector('dialog')).toBeNull()
  })
  it('recupera un conflicto de cancelación consultando la versión vigente', async () => {
    await open(`/reservas/${confirmed.id}`); await click('Cancelar reserva')
    failure = { path: '/cancelaciones', remaining: 1, code: 'RESERVATION_VERSION_CONFLICT' }
    await click('Sí, cancelar reserva', document.querySelector('dialog')!)
    expect(document.querySelector('dialog')?.textContent).toContain('Error de prueba')
    await click('Actualizar información', document.querySelector('dialog')!)
    expect(document.querySelector('dialog')).toBeNull()
    await click('Modificar reserva'); expect(router.currentRoute.value.path).toContain('/editar')
  })
})

describe('recursos y configuración', () => {
  it('filtra recursos por estado y navega a crear y administrar', async () => {
    paginate = true; await open('/recursos'); await click('Cargar más')
    expect(wrapper.text()).toContain(resources[1]!.name)
    await click('Deshabilitados'); expect(wrapper.text()).toContain('Recurso 3')
    await wrapper.get('button[aria-label="Administrar Recurso 3"]').trigger('click'); await settle()
    expect(router.currentRoute.value.path).toBe('/recursos/00000000-0000-4000-8000-100000000003')
    await router.push('/recursos'); await settle(); await click('Crear recurso')
    expect(router.currentRoute.value.path).toBe('/recursos/nuevo')
  })
  it('crea un recurso con bloque fijo y horario opcional', async () => {
    await open('/recursos/nuevo')
    await input('Nombre', 'Cancha nueva'); await input('Duración de cada reserva', '30')
    await wrapper.findAll('input[type="checkbox"]')[1]!.setValue(true)
    await wrapper.get('.day-toggle input').setValue(true)
    await submit()
    expect(wrapper.text()).toContain('Recurso creado')
    const body = calls.find(c => c.method === 'POST')!.body
    expect(body.policy).toMatchObject({ slotDurationSeconds: 1800, minDurationSeconds: 1800, maxDurationSeconds: 1800 })
    expect(body.weeklySchedule).toHaveLength(1)
    expect(body.attributes).toEqual({})
  })
  it('rechaza nombre, zona horaria y política inválidos sin enviar cambios', async () => {
    await open('/recursos/nuevo'); await submit(); expect(wrapper.text()).toContain('El nombre es obligatorio')
    await input('Nombre', 'Sala'); await input('Zona horaria IANA', 'zona/invalida'); await submit()
    expect(wrapper.text()).toContain('zona horaria IANA válida')
    await input('Zona horaria IANA', 'UTC'); await input('Duración de cada reserva', ''); await submit()
    expect(wrapper.text()).toContain('número entero de minutos')
    await input('Duración de cada reserva', '60'); await input('Anticipación para cambios', '-1'); await submit()
    expect(wrapper.text()).toContain('anticipación no negativa')
    expect(calls.some(c => c.method === 'POST')).toBe(false)
    await click('Cancelar'); expect(router.currentRoute.value.path).toBe('/recursos')
  })
  it('descarta y guarda información general sin alterar otros recursos', async () => {
    await open('/recursos/00000000-0000-4000-8000-100000000001')
    await input('Nombre', 'Otro nombre'); await click('Descartar cambios')
    expect(wrapper.get('input[placeholder="Ej. Recurso Norte"]').element.value).toBe(resources[0]!.name)
    await input('Nombre', 'Nombre guardado'); await submit()
    expect(wrapper.text()).toContain('Cambios guardados')
    expect(calls.find(c => c.method === 'PATCH')?.body).toMatchObject({ expectedVersion: 1, name: 'Nombre guardado' })
  })
  it('requiere recargar un conflicto de versión del recurso', async () => {
    await open('/recursos/00000000-0000-4000-8000-100000000001'); failure = { path: '/recursos/00000000-0000-4000-8000-100000000001', method: 'PATCH', remaining: 1, code: 'RESOURCE_VERSION_CONFLICT' }
    await submit(); expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
    await click('Actualizar información'); expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
  })
  it('edita franjas, valida solapamiento y guarda el horario con su versión', async () => {
    await open('/recursos/00000000-0000-4000-8000-100000000001'); await click('Horarios')
    const monday = wrapper.get('.schedule-day')
    await monday.get('.add-range').trigger('click'); await submit()
    expect(wrapper.text()).toContain('no pueden solaparse')
    await monday.get('[aria-label="Eliminar franja Lunes 2"]').trigger('click')
    await monday.get('[aria-label="Apertura Lunes 1"]').setValue('09:00')
    await submit(); expect(wrapper.text()).toContain('Horario actualizado')
    expect(calls.find(c => c.method === 'PUT')?.body.weeklySchedule[0].startLocalTime).toBe('09:00')
    await monday.get('input[type="checkbox"]').setValue(false)
    expect(monday.text()).toContain('Cerrado')
  })
  it('valida y crea un bloqueo temporal, y confirma su eliminación', async () => {
    await open('/recursos/00000000-0000-4000-8000-100000000001'); await click('Bloqueos'); await click('Crear bloqueo')
    await submit(); expect(wrapper.text()).toContain('El motivo es obligatorio')
    await input('Inicio', '2026-11-13T12:00'); await submit(); expect(wrapper.text()).toContain('anterior al término')
    await input('Inicio', '2026-11-12T12:00'); await input('Motivo', 'Mantenimiento'); await submit()
    expect(wrapper.text()).toContain('Bloqueo creado')
    await click('Eliminar bloqueo'); await click('Volver', document.querySelector('dialog')!)
    await click('Eliminar bloqueo'); await click('Sí, eliminar bloqueo', document.querySelector('dialog')!)
    expect(wrapper.text()).toContain('Bloqueo eliminado')
    expect(calls.some(c => c.method === 'DELETE')).toBe(true)
  })
  it('conserva el bloqueo si falla la eliminación y permite reintentar', async () => {
    await open('/recursos/00000000-0000-4000-8000-100000000001'); await click('Bloqueos'); await click('Eliminar bloqueo')
    failure = { path: '/bloqueos/00000000-0000-4000-8000-300000000001', method: 'DELETE', remaining: 1 }
    await click('Sí, eliminar bloqueo', document.querySelector('dialog')!)
    expect(document.querySelector('dialog')?.textContent).toContain('Error de prueba')
    await click('Sí, eliminar bloqueo', document.querySelector('dialog')!)
    expect(document.querySelector('dialog')).toBeNull()
  })
  it('recupera fallos de carga de recursos y detalle', async () => {
    failure = { path: '/recursos', remaining: 1 }; await open('/recursos')
    await click('Reintentar'); expect(wrapper.text()).toContain(resources[0]!.name)
    failure = { path: '/recursos/00000000-0000-4000-8000-100000000001', remaining: 1 }
    await router.push('/recursos/00000000-0000-4000-8000-100000000001'); await settle()
    expect(wrapper.text()).toContain('Error de prueba'); await click('Reintentar')
    expect(wrapper.text()).toContain('Información general')
    refreshKey.value++; await settle(); expect(wrapper.text()).toContain('Información general')
  })
})
