<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { api } from '../api/client'
import type { Availability, Reservation } from '../api/types'
import { availabilityScenario, DEMO_NOW, refreshKey } from '../demo'
import { duration, formatDate, localParts, toInstant } from '../dates'
import Field from './Field.vue'
import Alert from './Alert.vue'
import LoadState from './LoadState.vue'
const props = defineProps<{ resourceId: string; zone: string; original?: Reservation }>()
const date = defineModel<string>('date', { required: true }), start = defineModel<string>('start', { required: true }), end = defineModel<string>('end', { required: true })
const emit = defineEmits<{ availability: [value: Availability | undefined]; changeResource: [] }>()
const month = ref(date.value.slice(0, 7))
const monthName = computed(() => new Intl.DateTimeFormat('es-CL', { month: 'long', year: 'numeric', timeZone: 'UTC' }).format(new Date(`${month.value}-01T12:00:00Z`)))
const days = computed(() => new Date(Number(month.value.slice(0, 4)), Number(month.value.slice(5, 7)), 0).getDate())
const blanks = computed(() => (new Date(`${month.value}-01T12:00:00Z`).getUTCDay() + 6) % 7)
const windows = ref<Availability>(), loading = ref(false), error = ref('')
const selectedDateLabel = computed(() => date.value ? formatDate(`${date.value}T12:00:00Z`, 'UTC') : 'Elige una fecha')
const current = computed(() => props.original && props.original.resourceId === props.resourceId && localParts(props.original.startAt, props.zone).date === date.value ? props.original : undefined)
function isSelected(w: { startAt: string; endAt: string }) { return localParts(w.startAt, props.zone).time === start.value && localParts(w.endAt, props.zone).time === end.value }
function select(w: { startAt: string; endAt: string }) { start.value = localParts(w.startAt, props.zone).time; end.value = localParts(w.endAt, props.zone).time }
function moveMonth(delta: number) { const d = new Date(`${month.value}-01T12:00:00Z`); d.setUTCMonth(d.getUTCMonth() + delta); month.value = d.toISOString().slice(0, 7); date.value = `${month.value}-01` }
function dayString(day: number) { return `${month.value}-${String(day).padStart(2, '0')}` }
let request = 0
async function load() {
  const id = ++request; loading.value = true; error.value = ''; windows.value = undefined; emit('availability', undefined)
  try {
    const d = new Date(`${date.value}T12:00:00Z`); d.setUTCDate(d.getUTCDate() + 1)
    const value = await api.availability(props.resourceId, toInstant(date.value, '00:00', props.zone), toInstant(d.toISOString().slice(0, 10), '00:00', props.zone), availabilityScenario.value)
    if (id === request) {
      windows.value = value
      if (!(current.value && isSelected(current.value)) && !value.availableSlots.some(isSelected)) { start.value = ''; end.value = '' }
      emit('availability', value)
    }
  } catch(e) { if (id === request) error.value = (e as Error).message }
  finally { if (id === request) loading.value = false }
}
const selectionDuration = computed(() => { try { return duration(toInstant(date.value, start.value, props.zone), toInstant(date.value, end.value, props.zone)) } catch { return '—' } })
watch([date, () => props.resourceId, () => props.zone, availabilityScenario, refreshKey], (_next, previous) => {
  if (previous?.length && (_next[0] !== previous[0] || _next[1] !== previous[1])) { start.value = ''; end.value = '' }
  void load()
}, { immediate: true })
onUnmounted(() => { request++ })
</script>
<template><div class="section-heading"><div><span class="eyebrow">Fecha y horario</span><h2>Elige fecha y horario</h2><p><strong>{{ api.resourceName(resourceId) }}</strong> · Horario de {{ zone }}</p></div><button class="change-link" @click="emit('changeResource')">Cambiar recurso</button></div>
  <div class="date-time-grid"><div class="calendar-card"><div class="calendar-head"><button aria-label="Mes anterior" @click="moveMonth(-1)">❯</button><strong>{{ monthName }}</strong><button aria-label="Mes siguiente" @click="moveMonth(1)">❯</button></div><div class="weekdays"><span v-for="d in ['Lu', 'Ma', 'Mi', 'Ju', 'Vi', 'Sá', 'Do']" :key="d">{{ d }}</span></div><div class="days"><i v-for="n in blanks" :key="`blank-${n}`" /><button v-for="day in days" :key="day" :class="{ selected: date === dayString(day), muted: dayString(day) < localParts(DEMO_NOW, zone).date }" :disabled="dayString(day) < localParts(DEMO_NOW, zone).date" :aria-label="dayString(day)" :aria-pressed="date === dayString(day)" @click="date = dayString(day)">{{ day }}</button></div><div class="calendar-foot"><span>Consulta diaria · sin límite de anticipación</span></div><Field v-slot="{ id, describedby }" label="Fecha seleccionada"><input :id="id" :aria-describedby="describedby" v-model="date" type="date" :min="localParts(DEMO_NOW, zone).date" @change="month = date.slice(0, 7)"></Field></div>
    <div class="availability-card"><h3 class="subheading">{{ selectedDateLabel }}</h3><LoadState :loading="loading" :error="error" @retry="load"><template v-if="windows"><p class="field-hint">Cada reserva dura {{ duration('1970-01-01T00:00:00Z', new Date(windows.slotDurationSeconds * 1000).toISOString()) }}.</p><Alert v-if="!windows.enabled" tone="warning">El recurso está deshabilitado.</Alert><Alert v-else-if="!windows.availableSlots.length">No hay horarios disponibles para esta fecha.</Alert><template v-else><div class="availability-head"><strong>Horarios disponibles</strong><span class="available-key"><i />Disponible</span></div><div class="slot-grid" role="group" aria-label="Seleccionar horario disponible"><button v-for="w in windows.availableSlots" :key="w.startAt" type="button" :class="['slot-option', { selected: isSelected(w) }]" :aria-pressed="isSelected(w)" @click="select(w)">{{ localParts(w.startAt, zone).time }} – {{ localParts(w.endAt, zone).time }}<span v-if="isSelected(w)">✓ Seleccionado</span></button></div></template></template></LoadState>
      <div v-if="current" class="current-slot"><span class="field-hint">Horario actual de tu reserva</span><button :class="['slot-option', { selected: isSelected(current) }]" :aria-pressed="isSelected(current)" @click="select(current)">Conservar {{ localParts(current.startAt, zone).time }} – {{ localParts(current.endAt, zone).time }}</button></div><div v-if="start && end" class="duration-row" role="status"><span>Horario seleccionado<br><strong>{{ start }} – {{ end }}</strong></span><span>Duración<br><strong>{{ selectionDuration }}</strong></span></div><p v-else class="field-hint">Selecciona un horario para continuar.</p>
    </div>
  </div><Alert tone="warning">La disponibilidad puede cambiar antes de confirmar.</Alert>
</template>
