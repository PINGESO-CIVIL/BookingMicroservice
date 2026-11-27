<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, ApiError, canChange } from '../api/client'
import type { Availability, Reservation, Resource, Schemas } from '../api/types'
import { DEMO_NOW, detailScenario, readScenario, refreshKey, writeScenario } from '../demo'
import { duration, formatDate, localParts, toInstant } from '../dates'
import Btn from '../components/Btn.vue'
import Alert from '../components/Alert.vue'
import LoadState from '../components/LoadState.vue'
import Icon from '../components/Icon.vue'
import ReservationInfo from '../components/ReservationInfo.vue'
import TimeSelection from '../components/TimeSelection.vue'
const route = useRoute(), router = useRouter(), editing = !!route.params.id
const step = ref(editing ? 2 : 1), selected = ref(''), search = ref(''), date = ref('2026-11-10'), start = ref(''), end = ref('')
const resources = ref<Resource[]>([]), original = ref<Reservation>(), availability = ref<Availability>(), receipt = ref<Reservation>(), loading = ref(false), processing = ref(false), error = ref(''), actionError = ref(''), errorCode = ref(''), nextCursor = ref<string | null>(null), moreLoading = ref(false)
const zone = computed(() => resources.value.find(r => r.id === selected.value)?.timeZone ?? api.resourceZone(selected.value))
const filtered = computed(() => resources.value.filter(r => r.name.toLocaleLowerCase('es').includes(search.value.toLocaleLowerCase('es'))))
const blocked = computed(() => editing && !!original.value && !canChange(original.value))
const interval = computed(() => { try { return { startAt: toInstant(date.value, start.value, zone.value), endAt: toInstant(date.value, end.value, zone.value) } } catch { return undefined } })
const intervalError = computed(() => {
  if (!selected.value) return 'Elige un recurso.'
  if (!interval.value) return 'Selecciona uno de los horarios disponibles.'
  const { startAt, endAt } = interval.value
  if (Date.parse(startAt) >= Date.parse(endAt)) return 'El inicio debe ser anterior al término.'
  if (Date.parse(startAt) <= Date.parse(DEMO_NOW)) return 'El inicio de la reserva debe ser futuro.'
  if (!availability.value) return 'Espera a que se consulte la disponibilidad.'
  if (!availability.value.enabled) return 'El recurso está deshabilitado.'
  const current = original.value
  const unchanged = current && current.resourceId === selected.value && Date.parse(current.startAt) === Date.parse(startAt) && Date.parse(current.endAt) === Date.parse(endAt)
  if (unchanged) return ''
  if (!availability.value.availableSlots.some(w => Date.parse(startAt) === Date.parse(w.startAt) && Date.parse(endAt) === Date.parse(w.endAt))) return 'Selecciona uno de los horarios disponibles.'
  return ''
})
const changed = computed(() => !editing || (!!original.value && !!interval.value && (selected.value !== original.value.resourceId || Date.parse(interval.value.startAt) !== Date.parse(original.value.startAt) || Date.parse(interval.value.endAt) !== Date.parse(original.value.endAt))))
let key = crypto.randomUUID(), request = 0
watch([selected, date, start, end], () => { key = crypto.randomUUID(); actionError.value = ''; errorCode.value = '' })
async function load() {
  const id = ++request; loading.value = true; error.value = ''; receipt.value = undefined; actionError.value = ''
  try {
    const page = await api.listResources({}, readScenario.value)
    let r: Reservation | undefined
    if (editing) r = await api.getReservation(String(route.params.id), detailScenario.value)
    if (id !== request) return
    resources.value = page.items; nextCursor.value = page.nextCursor; original.value = r
    if (r) { selected.value = r.resourceId; const p = localParts(r.startAt, api.resourceZone(r.resourceId)); date.value = p.date; start.value = p.time; end.value = localParts(r.endAt, api.resourceZone(r.resourceId)).time }
  } catch(e) { if (id === request) error.value = (e as Error).message }
  finally { if (id === request) loading.value = false }
}
async function more() {
  if (!nextCursor.value || moreLoading.value) return
  moreLoading.value = true
  try { const page = await api.listResources({ cursor: nextCursor.value }, readScenario.value); resources.value.push(...page.items); nextCursor.value = page.nextCursor } catch(e) { actionError.value = (e as Error).message } finally { moreLoading.value = false }
}
async function confirm() {
  if (processing.value || intervalError.value || !interval.value || blocked.value || !changed.value) return
  processing.value = true; actionError.value = ''; errorCode.value = ''
  try {
    if (editing && original.value) {
      const body: Schemas['UpdateReservationRequest'] = { expectedVersion: original.value.version }
      if (selected.value !== original.value.resourceId) body.resourceId = selected.value
      if (Date.parse(interval.value.startAt) !== Date.parse(original.value.startAt)) body.startAt = interval.value.startAt
      if (Date.parse(interval.value.endAt) !== Date.parse(original.value.endAt)) body.endAt = interval.value.endAt
      receipt.value = await api.updateReservation(original.value.id, body, writeScenario.value)
    } else {
      receipt.value = await api.createReservation({ resourceId: selected.value, ...interval.value, attributes: {} }, key, writeScenario.value)
    }
  } catch(e) { actionError.value = (e as Error).message; errorCode.value = e instanceof ApiError ? e.code : '' }
  finally { processing.value = false }
}
function leave() { router.push(editing ? `/reservas/${route.params.id}` : '/reservas') }
watch([readScenario, detailScenario, refreshKey], load, { immediate: true })
</script>
<template><div class="page">
  <LoadState :loading="loading" :error="error" @retry="load">
    <section v-if="receipt" class="result-card narrow-page page"><div class="success-mark"><Icon name="check" :size="32" /></div><h2>{{ editing ? 'Reserva actualizada' : 'Reserva confirmada' }}</h2><p>Resultado de demostración</p><ReservationInfo :reservation="receipt" /><Alert>Esta respuesta es prefijada y puede diferir de tu selección. No se guarda ni modifica consultas posteriores.</Alert><div class="result-actions"><Btn @click="router.push(`/reservas/${receipt.id}`)">Ver reserva</Btn><Btn variant="secondary" @click="router.push('/reservas')">Volver a mis reservas</Btn></div></section>
    <template v-else-if="blocked"><Alert tone="warning">Esta reserva no admite cambios de horario o recurso: está cancelada, ya comenzó o venció el plazo.</Alert><Btn variant="secondary" @click="leave">Volver a la reserva</Btn></template>
    <template v-else>
      <Alert v-if="editing" title="Estás modificando tu reserva">Tu selección actual se conserva mientras eliges otro horario.</Alert>
      <ol v-else class="stepper"><li v-for="(label, i) in ['Elegir recurso', 'Fecha y horario', 'Revisar y confirmar']" :key="label" :class="i + 1 < step ? 'done' : i + 1 === step ? 'current' : ''" :aria-current="i + 1 === step ? 'step' : undefined"><span>{{ i + 1 < step ? '✓' : i + 1 }}</span><div><small>Paso {{ i + 1 }}</small><strong>{{ label }}</strong></div><i v-if="i < 2" /></li></ol>
      <div class="booking-layout"><section class="form-card">
        <template v-if="step === 1"><div class="section-heading"><div><span class="eyebrow">Paso 1 de 3</span><h2>Elige un recurso</h2><p>Selecciona el recurso que quieres reservar.</p></div></div><div class="search-box"><Icon name="search" /><input v-model="search" aria-label="Buscar recurso" placeholder="Buscar por nombre"></div><div class="resource-options"><button v-for="r in filtered" :key="r.id" :class="['resource-option', { selected: selected === r.id }]" :aria-pressed="selected === r.id" @click="selected = r.id"><div class="resource-symbol"><Icon name="building" /></div><div><strong>{{ r.name }}</strong><span><Icon name="clock" :size="14" />{{ r.timeZone }}</span></div><span class="radio"><i v-if="selected === r.id" /></span></button></div><p v-if="!filtered.length" class="field-hint">No hay recursos que coincidan con tu búsqueda.</p><div v-if="nextCursor" class="load-more"><Btn variant="secondary" :loading="moreLoading" @click="more">Cargar más</Btn></div></template>
        <TimeSelection v-if="step === 2" v-model:date="date" v-model:start="start" v-model:end="end" :resource-id="selected" :zone="zone" :original="original" @availability="availability = $event" @change-resource="step = 1" />
        <template v-if="step === 3"><div class="section-heading"><div><span class="eyebrow">Paso 3 de 3</span><h2>Revisa tu reserva</h2><p>Confirma que los datos sean correctos antes de continuar.</p></div></div><div class="review-card"><div class="review-icon"><Icon name="building" /></div><div><span>Recurso</span><strong>{{ api.resourceName(selected) }}</strong></div><dl><div><dt>Fecha</dt><dd>{{ formatDate(`${date}T12:00:00Z`, 'UTC') }}</dd></div><div><dt>Horario</dt><dd>{{ start }} – {{ end }}</dd></div><div><dt>Duración</dt><dd>{{ interval ? duration(interval.startAt, interval.endAt) : '—' }}</dd></div><div><dt>Zona horaria</dt><dd>{{ zone }}</dd></div></dl></div><Alert tone="success" title="La reserva quedará a tu nombre">Andrea Morales será la responsable de esta reserva.</Alert></template>
        <Alert v-if="step > 1 && intervalError" tone="warning">{{ intervalError }}</Alert><Alert v-if="actionError" tone="error">{{ actionError }}<template v-if="errorCode.endsWith('VERSION_CONFLICT')" #action><Btn variant="secondary" @click="load(); step = 2">Actualizar información</Btn></template></Alert>
        <div class="form-actions"><Btn variant="secondary" :disabled="processing" @click="step === 1 || editing ? leave() : step--">{{ editing ? 'Descartar cambios' : step === 1 ? 'Cancelar' : 'Volver' }}</Btn><Btn v-if="step < 3" :disabled="!selected || (step === 2 && !!intervalError)" @click="step++">Continuar</Btn><Btn v-else :loading="processing" :disabled="!!intervalError || !changed || errorCode.endsWith('VERSION_CONFLICT')" @click="confirm">{{ editing ? 'Guardar cambios' : 'Confirmar reserva' }}</Btn></div>
      </section><aside class="help-card"><Icon name="info" /><div><strong>Antes de reservar</strong><p>La disponibilidad puede cambiar antes de confirmar.</p><p>Todos los horarios se muestran en la zona horaria del recurso.</p></div></aside></div>
    </template>
  </LoadState>
</div></template>
