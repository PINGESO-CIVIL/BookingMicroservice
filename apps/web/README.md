# Frontend BookingMS

Las vistas y los componentes propios de cada flujo viven en `src/features/reservas/` y `src/features/recursos/`. Los componentes visuales reutilizables viven en `src/components/`.

Migración a **Vue 3 + TypeScript + Vite** del mockup React descargado de Figma Make, conservando la barra lateral, colores, formularios, tarjetas, tablas y adaptación móvil. No se utilizan React, React DOM ni JSX. El CSS original se adaptó para funcionar sin Tailwind y sin descargar fuentes externas.

## Iniciar

Desde `apps/web`, con Node 22.22.2+ (rama 22), 24.15+ (rama 24) o 26+:

```powershell
npm ci
npm run dev
```

Abrir http://127.0.0.1:5173. Las rutas usan hash, de manera que recargar un detalle funciona también con alojamiento estático. Cambiar entre Usuario y Admin. mediante el selector de demostración.

Para levantar toda la demo, usar `docker compose up -d --build` desde la raíz y abrir http://127.0.0.1:5190. Nginx reenvía `/api/` al servicio `backend` de Compose, que se publica en el puerto 5191. En desarrollo local, Vite usa el puerto 5173 y reenvía las peticiones al mock Go en 4010. Ver [la demo del Hito 1](../../README.md).

```powershell
npm run contracts
npm run build
npm test
npm run test:coverage
npm run test:e2e
npm run verify:http
```

`contracts` regenera los tipos públicos y los ejemplos desde `../../contracts/openapi.yaml`. Los archivos generados se incluyen en el repositorio. `build` comprueba los componentes Vue y genera `dist/`. Las pruebas de navegador utilizan Microsoft Edge instalado y un servidor exclusivo en el puerto 5180; no interrumpen el servidor de desarrollo. Para otro sistema, sustituir `channel: 'msedge'` en `playwright.config.ts` por el navegador disponible.

`test:coverage` genera LCOV en `../../reports/frontend/lcov.info`, incluyendo TypeScript y componentes Vue, salvo declaraciones de tipos. `scripts/quality.ps1 Sonar` ejecuta esta cobertura y normaliza sus rutas para que SonarQube la importe desde la raíz del repositorio. Los componentes sin pruebas se mantienen en la medición; no se excluyen para elevar porcentajes.

Vitest 5 ejecuta además pruebas de los flujos Vue con el router y cliente HTTP reales, sustituyendo únicamente el transporte: reservas personales/administrativas, creación, reprogramación, cancelación, conflictos de versión, recursos, horarios y bloqueos. Los umbrales locales de líneas, sentencias, funciones y ramas son 80%. Estas pruebas son independientes del Quality Gate de SonarQube.

## Alcance de esta entrega

Es un **prototipo Vue conectado por HTTP al mock Go** del Hito 1. Las 14 operaciones tienen su equivalente en `src/api/client.ts`, con solicitudes y respuestas tipadas mediante OpenAPI. Vite reenvía `/api` al puerto 4010. Si el servidor no está disponible se muestra un error: no existe fallback a respuestas locales. La variable opcional `VITE_API_BASE_URL` permite configurar otra base HTTP, incluyendo `/api/v1`; requiere un despliegue con CORS/proxy apropiado. Una publicación estática necesita conservar ese acceso a la API.

Las escrituras devuelven ejemplos del contrato, no reflejan necesariamente el formulario enviado y **no persisten ni modifican las consultas posteriores**. La pantalla de resultado muestra la respuesta recibida y lo advierte. No hay almacenamiento local, base de datos, ocupación compartida, autenticación real, envío de correos ni llamadas a G10. El cambio de rol solo permite navegar por la demostración y no implementa autorización.

Los filtros y la paginación consultan fixtures inmutables a través del servidor Go. Los ejemplos incluyen reserva cancelada, plazo vencido y fecha pasada. La disponibilidad se prepara al generar los fixtures a partir del horario semanal, duración fija, reservas y cierres de ejemplo; Go sirve esos resultados sin calcular reglas de negocio durante las solicitudes. Los conflictos de escritura y fallos de servicio son escenarios seleccionables, sin garantías de exclusividad ni validación de un backend productivo.

Fecha fija de referencia: **4 de noviembre de 2026 a las 13:00, America/Santiago** (`2026-11-04T16:00:00Z`). La configuración de ejemplo abre lunes y martes de 08:00 a 22:00. Recurso 1 ofrece bloques de 60 minutos y Recurso 2, de 30 minutos. El 10 de noviembre Recurso 1 tiene una reserva de 19:00 a 20:00: ese bloque no se ofrece. Los días cerrados no muestran opciones. No se escriben manualmente las horas; el usuario selecciona un bloque disponible.

En administración, “Duración de cada reserva” configura los bloques en minutos enteros mediante `policy.slotDurationSeconds`; los límites mínimo y máximo se mantienen iguales a esa duración al editarla. Como las escrituras de este hito son respuestas prefijadas, guardar una configuración todavía no cambia el catálogo consultado después. Los datos HTTP cubren noviembre de 2026 a enero de 2027; fuera de ese período pueden devolverse listas vacías, sin establecer un horizonte de negocio.

## Revisión funcional

La revisión detallada se encuentra en [docs/revision-mockup-vue.md](../docs/revision-mockup-vue.md).

- Reservas personales: listado, filtros combinados y de solapamiento, detalle por identificador, creación, modificación y cancelación.
- Recursos administrativos: listado por estado, alta, detalle, edición, políticas, múltiples franjas semanales y bloqueos.
- Escenarios: vacío, carga, fallo de lectura, varias páginas, indisponibilidad, versión desactualizada, plazo vencido y permisos simulados.
- Presentación temporal por zona IANA del recurso, entrada en hora local y conversión a UTC. Las horas inexistentes o ambiguas en cambios de horario requieren otra selección.
- Diálogos nativos con foco contenido, cierre por Escape y devolución del foco; navegación por teclado, etiquetas de campos y estados accesibles.

## Pendientes del contrato y del siguiente hito

La disponibilidad pública ahora incluye `slotDurationSeconds` y `availableSlots`, para que la selección respete la duración del recurso sin consultar su detalle administrativo. `availableWindows` conserva las ventanas continuas como información separada. La política de cancelación antes de confirmar sigue pendiente de exposición pública; la fecha límite se muestra en resultado y detalle.

`attributes` solo admite `{}` en el contrato genérico. No se agregaron notas, participantes, actividades ni formularios JSON. El administrador dispone de “Reservas realizadas” para consultar reservas de todos los usuarios de su ámbito, con responsable, filtros y detalle de lectura. No hay botón para crear reservas en esa vista ni acciones para modificar o cancelar reservas de otros usuarios. El panel de escenarios se retiró de la interfaz; permanece la selección de escenarios en Swagger.

Las reglas de dominio deben volver a validarse en el futuro backend: las comprobaciones visuales no sustituyen permisos, validación contextual, concurrencia o atomicidad. Al editar, se conserva la selección existente aunque la consulta pública la incluya como ocupada; no se inventa un parámetro para excluir reservas.

## Procedencia visual

Fuente: exportación del usuario `Prototipo de Reservas BookingMS` (Figma Make). Se reutilizaron el CSS adaptado y el favicon; las vistas, componentes, navegación y adaptador se reescribieron en Vue. Se amplió el contrato a la versión 0.3.0 para incluir bloques configurables, por decisión posterior del usuario. La descarga original se conserva.
