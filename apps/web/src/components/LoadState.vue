<script setup lang="ts">
import Alert from './Alert.vue'
import Btn from './Btn.vue'
import Icon from './Icon.vue'
defineProps<{ loading?: boolean; error?: string; empty?: boolean; title?: string; description?: string }>()
defineEmits<{ retry: [] }>()
</script>
<template>
  <div v-if="loading" class="loading-state" role="status" aria-label="Cargando información"><div v-for="n in 3" :key="n" class="skeleton" />Cargando información…</div>
  <Alert v-else-if="error" tone="error" title="No pudimos completar la consulta">{{ error }}<template #action><Btn variant="secondary" @click="$emit('retry')">Reintentar</Btn></template></Alert>
  <div v-else-if="empty" class="empty-state"><div class="empty-icon"><Icon name="calendar" :size="28" /></div><h3>{{ title ?? 'Sin resultados' }}</h3><p>{{ description }}</p><slot name="empty" /></div>
  <slot v-else />
</template>
