<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/client'
import type { Reservation } from '../api/types'
import { DEMO_NOW, readScenario, refreshKey } from '../demo'
import { duration, formatDate, localParts, toInstant } from '../dates'
import Btn from '../components/Btn.vue'
import Field from '../components/Field.vue'
import Badge from '../components/Badge.vue'
import Alert from '../components/Alert.vue'
import LoadState from '../components/LoadState.vue'
import Modal from '../components/Modal.vue'
import ReservationInfo from '../components/ReservationInfo.vue'
const props = defineProps<{ admin?: boolean }>()
const router = useRouter()
const selected = ref<Reservation>()
const filters = reactive({ resourceId: '', status: '', from: '', to: '' })
const items = ref<Reservation[]>([]), nextCursor = ref<string | null>(null), loading = ref(false), moreLoading = ref(false), error = ref('')
const filtered = computed(() => Object.values(filters).some(Boolean))
const rangeError = computed(() => !!filters.from !== !!filters.to ? 'Completa ambas fechas para filtrar.' : filters.from && filters.from >= filters.to ? 'El término debe ser posterior al inicio.' : '')
let request = 0
async function load(more = false) {
  const current = ++request
  selected.value = undefined
  if (rangeError.value) return
  error.value = ''; if (more) moreLoading.value = true; else loading.value = true
  try {
    const page = await (props.admin ? api.listAdminReservations : api.listReservations)({ resourceId: filters.resourceId || undefined, status: filters.status || undefined, from: filters.from ? toInstant(filters.from, '00:00', 'America/Santiago') : undefined, to: filters.to ? toInstant(filters.to, '00:00', 'America/Santiago') : undefined, cursor: more ? nextCursor.value ?? undefined : undefined }, readScenario.value)
    if (current !== request) return
    items.value = more ? [...items.value, ...page.items] : page.items; nextCursor.value = page.nextCursor
  } catch (e) { if (current === request) error.value = (e as Error).message }
  finally { if (current === request) { loading.value = false; moreLoading.value = false } }
}
function clear() { Object.assign(filters, { resourceId: '', status: '', from: '', to: '' }) }
watch([() => ({ ...filters }), readScenario, refreshKey], () => { void load() }, { immediate: true })
const zone = (r: Reservation) => api.resourceZone(r.resourceId)
</script>
<template>
  <div class="page"><div class="page-actions"><div class="reference-date">{{ admin ? 'Reservas de todos los usuarios del ámbito administrado' : 'Fecha de referencia: 4 nov 2026' }}</div><Btn v-if="!admin" icon="plus" @click="router.push('/reservas/nueva')">Nueva reserva</Btn></div>
    <section class="filters" aria-label="Filtros de reservas"><div class="filter-row">
      <Field v-slot="{ id, describedby }" label="Recurso"><select :id="id" :aria-describedby="describedby" v-model="filters.resourceId"><option value="">Todos los recursos</option><option v-for="id in ['resource-001', 'resource-002', 'resource-003']" :key="id" :value="id">{{ api.resourceName(id) }}</option></select></Field>
      <Field v-slot="{ id, describedby }" label="Estado"><select :id="id" :aria-describedby="describedby" v-model="filters.status"><option value="">Todas</option><option value="confirmed">Confirmadas</option><option value="cancelled">Canceladas</option></select></Field>
      <Field v-slot="{ id, describedby }" label="Fecha de inicio"><input :id="id" :aria-describedby="describedby" v-model="filters.from" type="date"></Field><Field v-slot="{ id, describedby }" label="Fecha de término" hint="Término exclusivo · America/Santiago"><input :id="id" :aria-describedby="describedby" v-model="filters.to" type="date"></Field>
      <button class="clear-filters" @click="clear">Limpiar filtros</button>
    </div></section>
    <Alert v-if="rangeError" tone="warning">{{ rangeError }}</Alert>
    <LoadState v-else :loading="loading" :error="error" :empty="!items.length" :title="filtered ? 'No encontramos coincidencias' : admin ? 'No hay reservas realizadas' : 'Aún no tienes reservas'" :description="filtered ? 'Prueba modificando los filtros aplicados.' : admin ? 'Aquí aparecerán las reservas de los usuarios de tu ámbito.' : 'Cuando crees una reserva, podrás consultarla aquí.'" @retry="load()">
      <template #empty><Btn v-if="filtered" variant="secondary" @click="clear">Limpiar filtros</Btn><Btn v-else-if="!admin" @click="router.push('/reservas/nueva')">Crear mi primera reserva</Btn></template>
      <section class="data-card desktop-table"><table><thead><tr><th>Recurso</th><th v-if="admin">Responsable</th><th>Fecha</th><th>Horario</th><th>Estado</th><th><span class="sr-only">Acciones</span></th></tr></thead><tbody>
        <tr v-for="r in items" :key="r.id"><td><strong>{{ api.resourceName(r.resourceId) }}</strong><span>{{ r.id }}</span></td><td v-if="admin"><strong>{{ r.responsible.subject }}</strong><span>{{ r.responsible.issuer }}</span></td><td><strong>{{ formatDate(r.startAt, zone(r)) }}</strong><span v-if="r.startAt < DEMO_NOW">Fecha pasada</span></td><td><strong>{{ localParts(r.startAt, zone(r)).time }} – {{ localParts(r.endAt, zone(r)).time }}</strong><span>{{ duration(r.startAt, r.endAt) }} · {{ zone(r) }}</span></td><td><Badge :status="r.status" /></td><td><button v-if="admin" class="text-link" :aria-label="`Ver detalle de ${r.id}`" @click="selected = r">Ver detalle →</button><RouterLink v-else class="text-link" :to="`/reservas/${r.id}`" :aria-label="`Ver detalle de ${r.id}`">Ver detalle →</RouterLink></td></tr>
      </tbody></table></section>
      <div class="mobile-cards"><article v-for="r in items" :key="r.id" class="reservation-card"><div class="card-head"><div><span class="eyebrow">{{ r.id }}</span><h3>{{ api.resourceName(r.resourceId) }}</h3></div><Badge :status="r.status" /></div><p v-if="admin" class="responsible-info">Responsable: {{ r.responsible.subject }} · {{ r.responsible.issuer }}</p><dl><div><dt>Fecha</dt><dd>{{ formatDate(r.startAt, zone(r)) }}</dd></div><div><dt>Horario</dt><dd>{{ localParts(r.startAt, zone(r)).time }} – {{ localParts(r.endAt, zone(r)).time }}</dd></div></dl><p v-if="r.startAt < DEMO_NOW" class="past-label">Fecha pasada</p><Btn v-if="admin" variant="secondary" @click="selected = r">Ver detalle</Btn><RouterLink v-else class="btn btn-secondary" :to="`/reservas/${r.id}`">Ver detalle</RouterLink></article></div>
      <div v-if="nextCursor" class="load-more"><Btn variant="secondary" :loading="moreLoading" @click="load(true)">Cargar más</Btn></div>
    </LoadState>
    <Modal v-if="admin && selected" title="Detalle de reserva" @close="selected = undefined"><div class="detail-head"><strong>{{ selected.id }}</strong><Badge :status="selected.status" /></div><p class="responsible-info">Responsable: {{ selected.responsible.subject }} · {{ selected.responsible.issuer }}</p><ReservationInfo :reservation="selected" /><div class="dialog-actions"><Btn variant="secondary" @click="selected = undefined">Cerrar</Btn></div></Modal>
  </div>
</template>
