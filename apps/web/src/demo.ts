import { ref } from 'vue'
export const DEMO_NOW = '2026-11-04T16:00:00Z'
export const readScenario = ref('normal')
export const writeScenario = ref('success')
export const detailScenario = ref('normal')
export const availabilityScenario = ref('normal')
export const refreshKey = ref(0)
export function resetScenarios() { readScenario.value = 'normal'; writeScenario.value = 'success'; detailScenario.value = 'normal'; availabilityScenario.value = 'normal'; refreshKey.value++ }
