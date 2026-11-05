<script setup lang="ts">
import { computed, useId } from 'vue'
const props = defineProps<{ id?: string; label: string; hint?: string; optional?: boolean }>()
const generatedId = useId()
const controlId = computed(() => props.id ?? generatedId)
const hintId = computed(() => props.hint ? `${controlId.value}-hint` : undefined)
</script>
<template><div class="field"><label :for="controlId" class="field-label">{{ label }} <span v-if="optional">Opcional</span></label><slot :id="controlId" :describedby="hintId" /><span v-if="hint" :id="hintId" class="field-hint">{{ hint }}</span></div></template>
