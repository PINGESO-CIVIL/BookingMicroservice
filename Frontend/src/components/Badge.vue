<script setup lang="ts">
defineProps<{ status: string }>()
const labels: Record<string, string> = { confirmed: 'Confirmada', cancelled: 'Cancelada', enabled: 'Habilitado', disabled: 'Deshabilitado' }
</script>
<template><span :class="['badge', ['confirmed', 'enabled'].includes(status) ? 'badge-success' : 'badge-neutral']"><span class="badge-dot" />{{ labels[status] ?? status }}</span></template>
