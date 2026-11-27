<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, ApiError, canChange } from '../../api/client'
import type { Reservation } from '../../api/types'
import { DEMO_NOW, detailScenario, readScenario, refreshKey, writeScenario } from '../../demo'
import { formatDateTime } from '../../dates'
import Btn from '../../components/Btn.vue'
import Badge from '../../components/Badge.vue'
import Alert from '../../components/Alert.vue'
import LoadState from '../../components/LoadState.vue'
import Modal from '../../components/Modal.vue'
import ReservationInfo from './ReservationInfo.vue'
const route = useRoute(), router = useRouter()
const reservation = ref<Reservation>(), loading = ref(false), error = ref(''), actionError = ref(''), errorCode = ref(''), processing = ref(false), dialog = ref(false), cancelled = ref(false)
const editable = computed(() => !!reservation.value && canChange(reservation.value))
let request = 0
async function load() {
  const id = ++request; loading.value = true; error.value = ''; actionError.value = ''; cancelled.value = false
  try { const r = await api.getReservation(String(route.params.id), readScenario.value === 'error' ? 'error' : detailScenario.value); if (id === request) reservation.value = r }
  catch(e) { if (id === request) error.value = (e as Error).message }
  finally { if (id === request) loading.value = false }
}
async function cancel() {
  if (processing.value || !reservation.value || !editable.value) return
  processing.value = true; actionError.value = ''; errorCode.value = ''
  try { reservation.value = await api.cancelReservation(reservation.value.id, { expectedVersion: reservation.value.version }, writeScenario.value); cancelled.value = true; dialog.value = false }
  catch(e) { actionError.value = (e as Error).message; errorCode.value = e instanceof ApiError ? e.code : '' }
  finally { processing.value = false }
}
watch([detailScenario, readScenario, refreshKey], load, { immediate: true })
</script>
<template><div class="page narrow-page"><RouterLink class="back-link" to="/reservas">← Volver a mis reservas</RouterLink>
  <LoadState :loading="loading" :error="error" @retry="load"><section v-if="reservation" class="detail-card"><Alert v-if="cancelled" tone="success" title="Reserva cancelada">Respuesta de demostración. La cancelación no se guarda y el historial volverá a mostrar los datos prefijados.</Alert><div class="detail-head"><Badge :status="reservation.status" /></div><ReservationInfo :reservation="reservation" />
    <Alert v-if="reservation.status === 'cancelled'">Esta reserva se conserva en tu historial y no puede reactivarse.</Alert>
    <Alert v-else-if="reservation.startAt <= DEMO_NOW" tone="warning">La reserva ya comenzó o su fecha pasó. No se puede modificar ni cancelar.</Alert>
    <Alert v-else-if="!editable" tone="warning">El plazo para cancelar o reprogramar esta reserva venció. La reserva sigue confirmada.</Alert>
    <template v-else><Alert>Puedes cancelar o reprogramar hasta {{ formatDateTime(reservation.cancellationDeadline, api.resourceZone(reservation.resourceId)) }}.</Alert><div class="detail-actions"><Btn icon="edit" @click="router.push(`/reservas/${route.params.id}/editar`)">Modificar reserva</Btn><Btn variant="danger" @click="dialog = true; actionError = ''">Cancelar reserva</Btn></div></template>
  </section></LoadState>
  <Modal v-if="dialog && reservation" title="¿Cancelar esta reserva?" :busy="processing" @close="dialog = false"><p>El horario quedará disponible y la reserva se conservará en tu historial.</p><div class="dialog-summary"><strong>{{ api.resourceName(reservation.resourceId) }}</strong><span>{{ formatDateTime(reservation.startAt, api.resourceZone(reservation.resourceId)) }}</span></div><Alert v-if="actionError" tone="error">{{ actionError }}</Alert><div class="dialog-actions"><Btn variant="secondary" :disabled="processing" @click="dialog = false">Volver</Btn><Btn v-if="errorCode === 'RESERVATION_VERSION_CONFLICT'" @click="dialog = false; load()">Actualizar información</Btn><Btn v-else variant="danger" :loading="processing" @click="cancel">Sí, cancelar reserva</Btn></div></Modal>
</div></template>
