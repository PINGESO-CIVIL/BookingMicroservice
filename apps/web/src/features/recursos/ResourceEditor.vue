<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, ApiError } from '../../api/client'
import type { Policy, ResourceDetail, ScheduleRule, Schemas } from '../../api/types'
import { readScenario, refreshKey, writeScenario } from '../../demo'
import { validatePolicy, validateSchedule } from '../../validation'
import Btn from '../../components/Btn.vue'
import Field from '../../components/Field.vue'
import Badge from '../../components/Badge.vue'
import Alert from '../../components/Alert.vue'
import LoadState from '../../components/LoadState.vue'
import PolicyFields from './PolicyFields.vue'
import ScheduleEditor from './ScheduleEditor.vue'
import ClosuresPanel from './ClosuresPanel.vue'
const route = useRoute(), router = useRouter(), creating = !route.params.id
const resource = ref<ResourceDetail>(), loading = ref(false), error = ref(''), actionError = ref(''), errorCode = ref(''), processing = ref(false), notice = ref(''), tab = ref('general'), configureSchedule = ref(false)
const general = reactive({ name: '', enabled: true, timeZone: 'America/Santiago' })
const policy = ref<Policy>({ cancellationNoticeSeconds: 7200, minDurationSeconds: 3600, maxDurationSeconds: 3600, slotDurationSeconds: 3600 })
const rules = ref<ScheduleRule[]>([])
function fill(r: ResourceDetail) { Object.assign(general, { name: r.name, enabled: r.enabled, timeZone: r.timeZone }); policy.value = { ...r.policy }; rules.value = r.weeklySchedule.map(rule => ({ ...rule })) }
function discard() { if (resource.value) fill(resource.value); notice.value = ''; actionError.value = ''; /* Version conflicts require a fresh read. */ }
let request = 0
async function load() {
  if (creating) return
  const id = ++request; loading.value = true; error.value = ''; actionError.value = ''; errorCode.value = ''; notice.value = ''
  try { const r = await api.getResource(String(route.params.id), readScenario.value === 'error' ? 'error' : 'normal'); if (id === request) { resource.value = r; fill(r) } }
  catch(e) { if (id === request) error.value = (e as Error).message }
  finally { if (id === request) loading.value = false }
}
async function save() {
  if (processing.value || errorCode.value.endsWith('VERSION_CONFLICT')) return
  actionError.value = ''; notice.value = ''; errorCode.value = ''
  try {
    const policyError = validatePolicy(policy.value)
    if (policyError) throw new Error(policyError)
    if (tab.value === 'schedule' || (creating && configureSchedule.value)) { const scheduleError = validateSchedule(rules.value); if (scheduleError) throw new Error(scheduleError) }
    if (!general.name.trim()) throw new Error('El nombre es obligatorio.')
    try { new Intl.DateTimeFormat('es', { timeZone: general.timeZone }).format() } catch { throw new Error('Ingresa una zona horaria IANA válida.') }
    processing.value = true
    const body: Schemas['CreateResourceRequest'] = { name: general.name.trim(), enabled: general.enabled, timeZone: general.timeZone, policy: { ...policy.value }, attributes: {} }
    if (creating && configureSchedule.value) body.weeklySchedule = rules.value.map(rule => ({ ...rule }))
    let response: ResourceDetail
    if (creating) response = await api.createResource(body, writeScenario.value)
    else if (tab.value === 'schedule') response = await api.updateSchedule(resource.value!.id, { expectedVersion: resource.value!.version, policy: { ...policy.value }, weeklySchedule: rules.value.map(rule => ({ ...rule })) }, writeScenario.value)
    else {
      response = await api.updateResource(resource.value!.id, { expectedVersion: resource.value!.version, ...body }, writeScenario.value)
    }
    resource.value = response; fill(response)
    notice.value = `${creating ? 'Recurso creado' : tab.value === 'schedule' ? 'Horario actualizado' : 'Cambios guardados'}: respuesta prefijada sin persistencia. Los valores mostrados corresponden al ejemplo recibido; las consultas posteriores conservan sus datos originales.`
  } catch(e) { actionError.value = (e as Error).message; errorCode.value = e instanceof ApiError ? e.code : '' }
  finally { processing.value = false }
}
watch([readScenario, refreshKey], load, { immediate: true })
watch(tab, () => { discard() })
</script>
<template><div class="page form-page"><RouterLink class="back-link" to="/recursos">← Volver a recursos</RouterLink><LoadState :loading="loading" :error="error" @retry="load">
  <template v-if="!creating && resource"><section class="resource-context"><div class="resource-symbol">▦</div><div><h2>{{ resource.name }}</h2><p><Badge :status="resource.enabled ? 'enabled' : 'disabled'" /><span>{{ resource.timeZone }}</span></p></div></section><div class="tabs" role="tablist" aria-label="Configuración del recurso"><button v-for="t in [{ id: 'general', label: 'General' }, { id: 'schedule', label: 'Horarios' }, { id: 'closures', label: 'Bloqueos' }]" :key="t.id" role="tab" :aria-selected="tab === t.id" :class="{ active: tab === t.id }" @click="tab = t.id">{{ t.label }}</button></div></template>
  <ClosuresPanel v-if="tab === 'closures' && resource" :resource="resource" />
  <form v-else class="form-card tab-panel" @submit.prevent="save">
    <Alert v-if="notice" tone="success">{{ notice }}</Alert><Alert v-if="actionError" tone="error">{{ actionError }}<template v-if="errorCode.endsWith('VERSION_CONFLICT')" #action><Btn variant="secondary" @click="load">Actualizar información</Btn></template></Alert>
    <template v-if="tab === 'general'"><div class="section-heading"><div><h2>Información general</h2><p>Los datos que identifican al recurso.</p></div></div><div class="two-cols"><Field v-slot="{ id, describedby }" label="Nombre"><input :id="id" :aria-describedby="describedby" v-model="general.name" required placeholder="Ej. Recurso Norte"></Field><Field v-slot="{ id, describedby }" label="Zona horaria IANA"><input :id="id" :aria-describedby="describedby" v-model="general.timeZone" required list="timezones"><datalist id="timezones"><option>America/Santiago</option><option>America/Bogota</option><option>Europe/Madrid</option><option>UTC</option></datalist></Field></div><label class="switch-row"><div><strong>Recurso habilitado</strong><span>Visible para nuevas reservas.</span></div><input v-model="general.enabled" type="checkbox"><i /></label><div class="section-divider" /></template>
    <template v-if="tab === 'schedule'"><div class="section-heading"><div><h2>Horario semanal</h2><p>Los horarios se interpretan en la zona horaria del recurso: {{ resource?.timeZone }}.</p></div></div><ScheduleEditor v-model="rules" /><div class="section-divider" /></template>
    <div class="section-heading"><div><h2>Políticas de reserva</h2><p>Define las duraciones y el plazo para cancelar o reprogramar.</p></div></div><PolicyFields v-model="policy" />
    <Alert v-if="!creating">Los cambios de política se aplican a nuevas reservas. Las reservas existentes conservan el plazo de cancelación que aceptaron.</Alert>
    <template v-if="creating"><label class="checkbox-row"><input v-model="configureSchedule" type="checkbox"><span><strong>Configurar horario semanal ahora</strong><small>También podrás hacerlo desde el detalle del recurso.</small></span></label><ScheduleEditor v-if="configureSchedule" v-model="rules" /></template>
    <div class="form-actions"><Btn variant="secondary" :disabled="processing" @click="creating ? router.push('/recursos') : discard()">{{ creating ? 'Cancelar' : 'Descartar cambios' }}</Btn><Btn type="submit" :loading="processing" :disabled="errorCode.endsWith('VERSION_CONFLICT')">{{ creating ? 'Crear recurso' : 'Guardar cambios' }}</Btn></div>
  </form>
</LoadState></div></template>
