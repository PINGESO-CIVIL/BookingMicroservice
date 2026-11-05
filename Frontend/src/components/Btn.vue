<script setup lang="ts">
import Icon from './Icon.vue'
withDefaults(defineProps<{ variant?: string; icon?: string; loading?: boolean; disabled?: boolean; type?: 'button' | 'submit' }>(), { variant: 'primary', type: 'button' })
</script>
<template><button :type="type" :class="['btn', `btn-${variant}`]" :disabled="disabled || loading" :aria-busy="loading || undefined"><span v-if="loading" class="spinner" /><Icon v-else-if="icon" :name="icon" /><slot /></button></template>
