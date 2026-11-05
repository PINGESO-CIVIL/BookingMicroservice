<script setup lang="ts">
import Icon from './Icon.vue'
withDefaults(defineProps<{ tone?: string; title?: string }>(), { tone: 'info' })
</script>
<template><div :class="['alert', `alert-${tone}`]" :role="tone === 'error' ? 'alert' : 'status'"><Icon class="alert-icon" :name="tone === 'success' ? 'check' : tone === 'info' ? 'info' : 'alert'" /><div class="alert-copy"><strong v-if="title">{{ title }}</strong><slot /></div><slot name="action" /></div></template>
