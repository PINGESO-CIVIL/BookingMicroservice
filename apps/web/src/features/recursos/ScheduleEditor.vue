<script setup lang="ts">
import type { ScheduleRule } from '../../api/types'
import Icon from '../../components/Icon.vue'
const rules = defineModel<ScheduleRule[]>({ required: true })
const days = ['Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado', 'Domingo']
function add(day: number) { rules.value = [...rules.value, { dayOfWeek: day, startLocalTime: '08:00', endLocalTime: '18:00' }] }
function toggle(day: number, checked: boolean) { if (checked) add(day); else rules.value = rules.value.filter(r => r.dayOfWeek !== day) }
function remove(rule: ScheduleRule) { rules.value = rules.value.filter(r => r !== rule) }
</script>
<template><div class="schedule-list"><section v-for="(day, index) in days" :key="day" class="schedule-day"><label class="day-toggle"><input type="checkbox" :checked="rules.some(r => r.dayOfWeek === index + 1)" @change="toggle(index + 1, ($event.target as HTMLInputElement).checked)"><strong>{{ day }}</strong></label><div class="day-ranges"><div v-for="(rule, i) in rules.filter(r => r.dayOfWeek === index + 1)" :key="i" class="schedule-time"><input v-model="rule.startLocalTime" type="time" required :aria-label="`Apertura ${day} ${i + 1}`"><span>a</span><input v-model="rule.endLocalTime" type="time" required :aria-label="`Cierre ${day} ${i + 1}`"><button type="button" class="icon-button" :aria-label="`Eliminar franja ${day} ${i + 1}`" @click="remove(rule)"><Icon name="trash" /></button></div><span v-if="!rules.some(r => r.dayOfWeek === index + 1)" class="closed-label">Cerrado</span><button v-else type="button" class="add-range" @click="add(index + 1)"><Icon name="plus" />Agregar franja</button></div></section></div></template>
