<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/client'
import type { Resource } from '../api/types'
import { readScenario, refreshKey } from '../demo'
import Btn from '../components/Btn.vue'
import Badge from '../components/Badge.vue'
import LoadState from '../components/LoadState.vue'
const router = useRouter(), enabled = ref(true), items = ref<Resource[]>([]), loading = ref(false), moreLoading = ref(false), error = ref(''), nextCursor = ref<string | null>(null)
let request = 0
async function load(more = false) {
  const id = ++request; error.value = ''; if (more) moreLoading.value = true; else loading.value = true
  try { const page = await api.listResources({ enabled: enabled.value, cursor: more ? nextCursor.value ?? undefined : undefined }, readScenario.value); if (id === request) { items.value = more ? [...items.value, ...page.items] : page.items; nextCursor.value = page.nextCursor } }
  catch(e) { if (id === request) error.value = (e as Error).message }
  finally { if (id === request) { loading.value = false; moreLoading.value = false } }
}
watch([enabled, readScenario, refreshKey], () => load(), { immediate: true })
</script>
<template><div class="page"><div class="page-actions"><div class="segmented" aria-label="Estado de los recursos"><button :class="{ active: enabled }" :aria-pressed="enabled" @click="enabled = true">Habilitados</button><button :class="{ active: !enabled }" :aria-pressed="!enabled" @click="enabled = false">Deshabilitados</button></div><Btn icon="plus" @click="router.push('/recursos/nuevo')">Crear recurso</Btn></div>
  <LoadState :loading="loading" :error="error" :empty="!items.length" title="No hay recursos en este listado" description="Prueba con otro estado o crea un recurso." @retry="load()"><section class="data-card desktop-table"><table><thead><tr><th>Nombre</th><th>Estado</th><th>Zona horaria</th><th><span class="sr-only">Acciones</span></th></tr></thead><tbody><tr v-for="r in items" :key="r.id"><td><strong>{{ r.name }}</strong><span>{{ r.id }}</span></td><td><Badge :status="r.enabled ? 'enabled' : 'disabled'" /></td><td>{{ r.timeZone }}</td><td><Btn variant="secondary" :aria-label="`Administrar ${r.name}`" @click="router.push(`/recursos/${r.id}`)">Administrar</Btn></td></tr></tbody></table></section><div class="mobile-cards"><article v-for="r in items" :key="r.id" class="reservation-card"><div class="card-head"><h3>{{ r.name }}</h3><Badge :status="r.enabled ? 'enabled' : 'disabled'" /></div><p class="inline-meta">{{ r.timeZone }}</p><Btn variant="secondary" @click="router.push(`/recursos/${r.id}`)">Administrar</Btn></article></div><div v-if="nextCursor" class="load-more"><Btn variant="secondary" :loading="moreLoading" @click="load(true)">Cargar más</Btn></div></LoadState>
</div></template>
