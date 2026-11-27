<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { Closure, ResourceDetail } from '../../api/types'
import { api } from '../../api/client'
import { writeScenario } from '../../demo'
import { formatDateTime, toInstant } from '../../dates'
import Btn from '../../components/Btn.vue'
import Field from '../../components/Field.vue'
import Alert from '../../components/Alert.vue'
import Modal from '../../components/Modal.vue'
const props = defineProps<{ resource: ResourceDetail }>()
const showForm = ref(false), removing = ref<Closure>(), processing = ref(false), error = ref(''), notice = ref(''), receipt = ref<Closure>()
const form = reactive({ start: '2026-11-12T12:00', end: '2026-11-12T15:00', reason: '' })
async function create() {
  if (processing.value) return
  error.value = ''; notice.value = ''; receipt.value = undefined
  try {
    const startAt = toInstant(form.start.slice(0, 10), form.start.slice(11), props.resource.timeZone), endAt = toInstant(form.end.slice(0, 10), form.end.slice(11), props.resource.timeZone)
    if (Date.parse(startAt) >= Date.parse(endAt)) throw new Error('El inicio del bloqueo debe ser anterior al término.')
    if (!form.reason.trim()) throw new Error('El motivo es obligatorio.')
    processing.value = true
    receipt.value = await api.createClosure(props.resource.id, { startAt, endAt, reason: form.reason.trim() }, writeScenario.value)
    showForm.value = false; notice.value = 'Bloqueo creado: respuesta prefijada. El listado y las consultas posteriores conservan los ejemplos originales.'
  } catch(e) { error.value = (e as Error).message }
  finally { processing.value = false }
}
async function remove() {
  if (!removing.value || processing.value) return
  processing.value = true; error.value = ''; notice.value = ''; receipt.value = undefined
  try { await api.deleteClosure(props.resource.id, removing.value.id, writeScenario.value); removing.value = undefined; notice.value = 'Bloqueo eliminado: respuesta simulada sin persistencia. La disponibilidad continúa dependiendo del horario y otras ocupaciones; el listado conserva los ejemplos originales.' }
  catch(e) { error.value = (e as Error).message }
  finally { processing.value = false }
}
</script>
<template><section class="form-card tab-panel"><div class="section-heading"><div><h2>Bloqueos temporales</h2><p>Horarios de {{ resource.timeZone }}.</p></div><Btn icon="plus" @click="showForm = true; error = ''; notice = ''">Crear bloqueo</Btn></div>
  <Alert v-if="notice" tone="success">{{ notice }}</Alert><Alert v-if="error && !removing" tone="error">{{ error }}</Alert>
  <div v-if="receipt" class="dialog-summary"><strong>{{ receipt.reason }}</strong><span>{{ formatDateTime(receipt.startAt, resource.timeZone) }} → {{ formatDateTime(receipt.endAt, resource.timeZone) }}</span></div>
  <form v-if="showForm" class="inline-form" @submit.prevent="create"><div class="two-cols"><Field v-slot="{ id, describedby }" label="Inicio"><input :id="id" :aria-describedby="describedby" v-model="form.start" type="datetime-local" required></Field><Field v-slot="{ id, describedby }" label="Término"><input :id="id" :aria-describedby="describedby" v-model="form.end" type="datetime-local" required></Field></div><Field v-slot="{ id, describedby }" label="Motivo"><input :id="id" :aria-describedby="describedby" v-model="form.reason" required placeholder="Ej. mantenimiento preventivo"></Field><div class="form-actions"><Btn variant="secondary" :disabled="processing" @click="showForm = false">Cancelar</Btn><Btn type="submit" :loading="processing">Crear bloqueo</Btn></div></form>
  <div v-if="resource.closures.length" class="blocks-list"><article v-for="block in resource.closures" :key="block.id"><div class="block-icon">▦</div><div><strong>{{ block.reason }}</strong><span><b>Inicio</b>{{ formatDateTime(block.startAt, resource.timeZone) }}</span><span><b>Término</b>{{ formatDateTime(block.endAt, resource.timeZone) }}</span></div><Btn variant="ghost" @click="removing = block; error = ''">Eliminar bloqueo</Btn></article></div><div v-else class="empty-state compact"><h3>No hay bloqueos</h3><p>Los bloqueos del recurso aparecerán aquí.</p></div>
  <Modal v-if="removing" title="¿Eliminar este bloqueo?" :busy="processing" @close="removing = undefined"><p>Se levantará el cierre excepcional. El horario semanal y otras ocupaciones seguirán determinando la disponibilidad.</p><div class="dialog-summary"><strong>{{ removing.reason }}</strong></div><Alert v-if="error" tone="error">{{ error }}</Alert><div class="dialog-actions"><Btn variant="secondary" :disabled="processing" @click="removing = undefined">Volver</Btn><Btn variant="danger" :loading="processing" @click="remove">Sí, eliminar bloqueo</Btn></div></Modal>
</section></template>
