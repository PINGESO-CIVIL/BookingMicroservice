# Contrato HTTP de reservas

Actualización 0.3.0: el usuario confirmó bloques de duración fija configurables por recurso. `ResourcePolicy` incluye `slotDurationSeconds` (minutos enteros expresados en segundos); disponibilidad incluye esa duración y `availableSlots`, manteniendo `availableWindows` como ventanas continuas. La interfaz selecciona bloques completos y no solicita escribir horas. El frontend calcula los bloques sobre configuración y ocupación de ejemplo, sin persistencia; el backend real sigue pendiente.

Huma registra las catorce operaciones a partir de los tipos y rutas Go en `apps/api/internal/api/`. `contract_annotations.yaml` aporta descripciones, restricciones del contrato publicado y ejemplos de la demo. Desde la raíz, con Go y las dependencias de `apps/web` instaladas, ejecutar `pnpm codegen` para generar [OpenAPI 3.1](../../contracts/openapi.yaml), los tipos TypeScript y los fixtures. Versionar estos archivos juntos.

## Transición a 0.4.0

- Las rutas de reservas y recursos se publican en español bajo `/api/v1`. Las tablas de operaciones de este documento recogen las rutas nuevas.
- Los IDs de reservas, recursos y bloqueos usan UUID. Las referencias de identidad y los identificadores de catálogos externos conservan su formato.
- Cada operación declara su permiso con `x-permission`. En la demo, `demo-user` tiene `reservas:leer`, `reservas:operar` y `recursos:leer`. `demo-admin` tiene además `reservas:administrar` y `recursos:administrar`. Estos nombres están pendientes de confirmación con la sección.

Los consumidores deben actualizar rutas e identificadores juntos. La publicación se coordina con los equipos que indique el documento de componentes, compartiendo el enlace al YAML en el commit acordado.

La [definición funcional de creación](crear-reserva.md) fue aprobada por el usuario el 2026-11-04, incluyendo `attributes` y creación personal. Los detalles de lógica real identificados como pendientes se mantienen para el siguiente hito.

La [definición funcional de modificación](modificar-reserva.md) fue aprobada por el usuario el 2026-11-04, incluyendo `expectedVersion` y reemplazo completo de `attributes` cuando se envía.

## Inventario de contratos

| Operación | Método y ruta | Documento | Estado funcional |
|---|---|---|---|
| Listar reservas del ámbito (Admin) | `GET /api/v1/admin/reservas` | [Consulta administrativa](listar-reservas-admin.md) | Solicitado por el usuario para Hito 1 (Mock) |
| Crear reserva | `POST /api/v1/reservas` | [Creación](crear-reserva.md) | Aprobado; detalles técnicos pendientes señalados |
| Modificar reserva | `PATCH /api/v1/reservas/{id}` | [Modificación](modificar-reserva.md) | Aprobado; detalles técnicos pendientes señalados |
| Cancelar reserva | `POST /api/v1/reservas/{id}/cancelaciones` | [Cancelación](cancelar-reserva.md) | Para revisión |
| Listar reservas propias | `GET /api/v1/reservas` | [Listado](listar-reservas.md) | Para revisión |
| Consultar reserva propia | `GET /api/v1/reservas/{id}` | [Detalle](consultar-reserva.md) | Para revisión |
| Listar recursos visibles | `GET /api/v1/recursos` | [Recursos](listar-recursos.md) | Para revisión |
| Consultar disponibilidad | `GET /api/v1/recursos/{id}/disponibilidades` | [Disponibilidad](consultar-disponibilidad.md) | Para revisión |
| Crear recurso (Admin) | `POST /api/v1/recursos` | [Administración de recursos](administrar-recursos.md) | Aprobado para Hito 1 (Mock) |
| Consultar detalle de recurso (Admin) | `GET /api/v1/recursos/{id}` | [Administración de recursos](administrar-recursos.md) | Aprobado para Hito 1 (Mock) |
| Modificar recurso / política (Admin) | `PATCH /api/v1/recursos/{id}` | [Administración de recursos](administrar-recursos.md) | Aprobado para Hito 1 (Mock) |
| Configurar horario recurrente (Admin) | `PUT /api/v1/recursos/{id}/horarios` | [Administración de recursos](administrar-recursos.md) | Aprobado para Hito 1 (Mock) |
| Registrar bloqueo temporal (Admin) | `POST /api/v1/recursos/{id}/bloqueos` | [Administración de recursos](administrar-recursos.md) | Aprobado para Hito 1 (Mock) |
| Eliminar bloqueo temporal (Admin) | `DELETE /api/v1/recursos/{id}/bloqueos/{closureId}` | [Administración de recursos](administrar-recursos.md) | Aprobado para Hito 1 (Mock) |

Las [convenciones comunes](convenciones.md) consolidan identidad, respuestas, fechas, errores y paginación. El usuario confirmó que todas las operaciones de reservas se limitan a sus propias reservas dentro de su ámbito. El 2026-11-04 aprobó los límites de página (20 por defecto, 100 máximo) y disponibilidad (31 días por consulta). Estos límites controlan la carga de datos por solicitud, no la cantidad de reservas ni la anticipación permitida para reservar.

Los documentos Markdown explican el comportamiento esperado; `openapi.yaml` formaliza las formas públicas y los ejemplos para la demostración. No se adelantan servicios, handlers ni repositorios Go. El servidor de mocks responderá con escenarios prefijados: las escrituras no persistirán ni modificarán las consultas posteriores. Las operaciones administrativas también están incluidas como respuestas simuladas del Hito 1.

## Verificación del contrato

En CI, `pnpm codegen` regenera el contrato y los tipos. El pipeline comprueba que coincidan con Git y compara el contrato con el de la rama base usando oasdiff. La verificación HTTP de la demo contrasta las respuestas del backend con los esquemas publicados.

## Swagger UI local

- Documentación: [http://127.0.0.1:5191/api/docs](http://127.0.0.1:5191/api/docs).
- Contrato servido: [http://127.0.0.1:5191/api/docs/openapi.yaml](http://127.0.0.1:5191/api/docs/openapi.yaml).
- Contiene las catorce operaciones, parámetros, cuerpos, ejemplos, códigos de respuesta y esquemas.
- **Try it out** llama al mismo backend que sirve la documentación. En **Authorize**, usar `demo-user` o `demo-admin`. `X-Mock-Scenario` selecciona ejemplos de éxito y error. Las escrituras devuelven fixtures y dejan intactas las consultas.

Desde la raíz del repositorio:

```powershell
docker compose up -d --build
docker compose ps
```

La API incluye el YAML y los archivos estáticos de Swagger en su imagen. Después de regenerar el YAML, ejecutar `docker compose up -d --build backend` y recargar el navegador. `API_ENV=dev` habilita los archivos de documentación. El valor predeterminado de Compose es `dev`.

Swagger UI tiene deshabilitado el validador público de swagger.io. La dirección local servida fue verificada contra el archivo original y sus catorce operaciones se comprobaron en el navegador.

## Criterios acordados

- Definir el contrato antes de implementar handlers y clientes que lo consuman.
- Especificar solicitudes, respuestas, errores, autenticación, fechas, paginación, control de versión e idempotencia según las reglas acordadas.
- Separar esquemas de creación, modificación y lectura: el cliente no debe poder sustituir al responsable ni el plazo original de cancelación mediante una actualización.
- En creación, el responsable es siempre el usuario autenticado; no se recibe en el cuerpo ni se admite reservar para terceros inicialmente.
- Listado, detalle, modificación y cancelación se limitan a reservas propias en el ámbito autenticado. Las reservas ajenas no se exponen por sus identificadores.
- Incluir `attributes` en creación y respuesta para datos específicos de la reserva, con validación explícita por contexto. En el caso genérico puede omitirse en la solicitud o enviarse vacío.
- Representar explícitamente instantes y unidades de duración; no exponer directamente `time.Duration` ni tipos BSON como contrato público.
- Mantener el contrato versionado y comprobar su validez y compatibilidad en integración continua.
- Mapear los tipos de transporte a los modelos internos; los documentos PostgreSQL pertenecen al repositorio.

El contrato se publica en OpenAPI 3.1. Los tipos TypeScript se generan desde OpenAPI y React consume el mock HTTP Go. La publicación es local; un despliegue remoto debe acordarse. La lógica real corresponde al Hito 2.

Referencia: [especificación oficial OpenAPI](https://spec.openapis.org/oas/latest.html).
