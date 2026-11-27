<script setup lang="ts">
import type { Reservation } from '../../api/types'
import { api } from '../../api/client'
import { duration, formatDate, formatDateTime, localParts } from '../../dates'
defineProps<{ reservation: Reservation }>()
</script>
<template><div class="resource-banner"><div class="resource-symbol">▦</div><div><span>Recurso</span><strong>{{ api.resourceName(reservation.resourceId) }}</strong><small>{{ api.resourceZone(reservation.resourceId) }}</small></div></div><dl class="detail-grid">
  <div><dt>Fecha</dt><dd>{{ formatDate(reservation.startAt, api.resourceZone(reservation.resourceId)) }}</dd></div><div><dt>Horario</dt><dd>{{ localParts(reservation.startAt, api.resourceZone(reservation.resourceId)).time }} – {{ localParts(reservation.endAt, api.resourceZone(reservation.resourceId)).time }}</dd></div>
  <div><dt>Duración</dt><dd>{{ duration(reservation.startAt, reservation.endAt) }}</dd></div><div><dt>Creada el</dt><dd>{{ formatDateTime(reservation.createdAt, api.resourceZone(reservation.resourceId)) }}</dd></div><div class="full"><dt>Fecha límite para cancelar o reprogramar</dt><dd>{{ formatDateTime(reservation.cancellationDeadline, api.resourceZone(reservation.resourceId)) }}</dd></div>
</dl></template>
