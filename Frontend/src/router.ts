import { createRouter, createWebHashHistory } from 'vue-router'
import Reservations from './views/Reservations.vue'
import BookingForm from './views/BookingForm.vue'
import BookingDetail from './views/BookingDetail.vue'
import Resources from './views/Resources.vue'
import ResourceEditor from './views/ResourceEditor.vue'

export const router = createRouter({
  history: createWebHashHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    { path: '/', redirect: '/reservas' },
    { path: '/reservas', component: Reservations, meta: { title: 'Mis reservas', subtitle: 'Consulta y administra tus reservas' } },
    { path: '/reservas/nueva', component: BookingForm, meta: { title: 'Nueva reserva', subtitle: 'Elige un recurso y encuentra un horario disponible' } },
    { path: '/reservas/:id', component: BookingDetail, meta: { title: 'Detalle de reserva', subtitle: 'Información y acciones de tu reserva' } },
    { path: '/reservas/:id/editar', component: BookingForm, meta: { title: 'Modificar reserva', subtitle: 'Ajusta el recurso o el horario de tu reserva' } },
    { path: '/recursos', component: Resources, meta: { title: 'Recursos', subtitle: 'Configura los recursos disponibles para reservar', admin: true } },
    { path: '/administracion/reservas', component: Reservations, props: { admin: true }, meta: { title: 'Reservas realizadas', subtitle: 'Consulta las reservas de los recursos de tu ámbito', admin: true } },
    { path: '/recursos/nuevo', component: ResourceEditor, meta: { title: 'Crear recurso', subtitle: 'Define su configuración y políticas iniciales', admin: true } },
    { path: '/recursos/:id', component: ResourceEditor, meta: { title: 'Administrar recurso', subtitle: 'Configuración, horarios y bloqueos', admin: true } },
    { path: '/:pathMatch(.*)*', redirect: '/reservas' },
  ],
})
