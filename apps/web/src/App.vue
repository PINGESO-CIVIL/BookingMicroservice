<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Icon from './components/Icon.vue'
import { resetScenarios } from './demo'
const route = useRoute(), router = useRouter()
const mobileOpen = ref(false)
const admin = computed(() => !!route.meta.admin)
function switchRole(isAdmin: boolean) { resetScenarios(); mobileOpen.value = false; router.push(isAdmin ? '/recursos' : '/reservas') }
</script>
<template>
  <a class="skip-link" href="#main-content">Ir al contenido</a>
  <div class="app-shell">
    <aside :class="['sidebar', { 'sidebar-open': mobileOpen }]" aria-label="Navegación y demostración">
      <div class="brand"><div class="brand-mark"><Icon name="calendar" :size="21" /></div><span>BookingMS</span><button class="mobile-close" aria-label="Cerrar navegación" @click="mobileOpen = false"><Icon name="x" /></button></div>
      <div class="demo-label"><span />Modo demostración</div>
      <nav class="side-nav" aria-label="Navegación principal"><p>{{ admin ? 'Administración' : 'Reservas' }}</p>
        <template v-if="!admin"><RouterLink to="/reservas" :class="{ active: route.path === '/reservas' || (route.params.id && !route.path.endsWith('/editar')) }" @click="mobileOpen = false"><Icon name="calendar" />Mis reservas</RouterLink><RouterLink to="/reservas/nueva" :class="{ active: route.path === '/reservas/nueva' || route.path.endsWith('/editar') }" @click="mobileOpen = false"><Icon name="plus" />Nueva reserva</RouterLink></template>
        <template v-else><RouterLink to="/recursos" :class="{ active: route.path.startsWith('/recursos') }" @click="mobileOpen = false"><Icon name="grid" />Recursos</RouterLink><RouterLink to="/administracion/reservas" :class="{ active: route.path === '/administracion/reservas' }" @click="mobileOpen = false"><Icon name="calendar" />Reservas realizadas</RouterLink></template>
      </nav>
      <div class="demo-tool"><div class="demo-tool-title"><Icon name="user" /><div><strong>Herramienta de demostración</strong><span>Cambiar rol activo</span></div></div><div class="role-toggle"><button :class="{ selected: !admin }" :aria-pressed="!admin" @click="switchRole(false)">Usuario</button><button :class="{ selected: admin }" :aria-pressed="admin" @click="switchRole(true)">Admin.</button></div></div>
      <div class="profile"><div class="avatar">AM</div><div><strong>Andrea Morales</strong><span>{{ admin ? 'Administrador de ámbito' : 'Usuario de demostración' }}</span></div></div>
    </aside>
    <button v-if="mobileOpen" class="scrim" aria-label="Cerrar navegación" @click="mobileOpen = false" />
    <div class="workspace"><header class="topbar"><button class="menu-button" aria-label="Abrir navegación" :aria-expanded="mobileOpen" @click="mobileOpen = true"><Icon name="menu" /></button><div><h1>{{ route.meta.title }}</h1><p>{{ route.meta.subtitle }}</p></div><div class="topbar-meta"><span class="demo-pill">Modo demostración</span><div class="mini-avatar">AM</div></div></header>
      <main id="main-content"><RouterView :key="route.fullPath" /></main>
    </div>
  </div>
</template>
