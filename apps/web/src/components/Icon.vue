<script setup lang="ts">
withDefaults(defineProps<{ name: string; size?: number }>(), { size: 18 })
const paths: Record<string, string> = {
  calendar: 'M3 10h18M8 3v4M16 3v4M5 5h14a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2',
  plus: 'M12 5v14M5 12h14', grid: 'M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z',
  chevron: 'm9 18 6-6-6-6', clock: 'M12 7v5l3 2M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0',
  search: 'm20 20-4-4M18 11a7 7 0 1 1-14 0 7 7 0 0 1 14 0',
  x: 'm6 6 12 12M18 6 6 18', check: 'm5 12 4 4L19 6', alert: 'M12 3 2.8 20h18.4L12 3ZM12 9v4M12 17h.01',
  arrow: 'm15 18-6-6 6-6', edit: 'm4 20 4.5-1 10-10a2.1 2.1 0 0 0-3-3l-10 10L4 20Zm10-13 3 3',
  trash: 'M4 7h16M9 7V4h6v3M7 7l1 13h8l1-13M10 11v5M14 11v5', menu: 'M4 7h16M4 12h16M4 17h16',
  building: 'M4 21V5h16v16M8 9h2M14 9h2M8 13h2M14 13h2M9 21v-4h6v4',
  user: 'M16 8a4 4 0 1 1-8 0 4 4 0 0 1 8 0M4 21a8 8 0 0 1 16 0',
  info: 'M12 11v5M12 8h.01M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0',
}
</script>
<template><svg aria-hidden="true" :width="size" :height="size" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path :d="paths[name] ?? paths.info" /></svg></template>
