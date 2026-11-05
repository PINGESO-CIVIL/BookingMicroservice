<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
defineProps<{ title: string; busy?: boolean }>()
const emit = defineEmits<{ close: [] }>()
const dialog = ref<HTMLDialogElement>()
let previous: HTMLElement | null = null
onMounted(() => { previous = document.activeElement as HTMLElement; dialog.value?.showModal() })
onUnmounted(() => previous?.focus())
</script>
<template><Teleport to="body"><dialog ref="dialog" class="dialog native-dialog" aria-labelledby="modal-title" @cancel.prevent="!busy && emit('close')"><h2 id="modal-title">{{ title }}</h2><slot /></dialog></Teleport></template>
