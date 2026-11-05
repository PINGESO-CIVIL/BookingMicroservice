<script setup lang="ts">
import type { Policy } from '../api/types'
import Field from './Field.vue'
const policy = defineModel<Policy>({ required: true })
function set(key: keyof Policy, value: string) {
  const seconds = value === '' ? Number.NaN : Number(value) * 60
  policy.value = key === 'slotDurationSeconds' ? { ...policy.value, slotDurationSeconds: seconds, minDurationSeconds: seconds, maxDurationSeconds: seconds } : { ...policy.value, [key]: seconds }
}
</script>
<template><div class="two-cols"><Field v-slot="{ id, describedby }" label="Duración de cada reserva" hint="Minutos enteros · bloques de duración fija"><input :id="id" :aria-describedby="describedby" :value="policy.slotDurationSeconds / 60" type="number" min="1" step="1" required @input="set('slotDurationSeconds', ($event.target as HTMLInputElement).value)"></Field><Field v-slot="{ id, describedby }" label="Anticipación para cambios" hint="Minutos para cancelar o reprogramar"><input :id="id" :aria-describedby="describedby" :value="policy.cancellationNoticeSeconds / 60" type="number" min="0" step="any" required @input="set('cancellationNoticeSeconds', ($event.target as HTMLInputElement).value)"></Field></div><p class="field-hint policy-help">Los horarios se dividen en bloques de esta duración desde la apertura de cada franja. Solo se ofrecen bloques completos y libres.</p></template>
