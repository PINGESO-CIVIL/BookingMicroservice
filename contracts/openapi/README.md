# Contrato HTTP de reservas

Este directorio alojará la especificación OpenAPI compartida por el backend Go y el frontend Vue. La [definición funcional de creación](crear-reserva.md) fue aprobada por el usuario el 2026-11-04, incluyendo `attributes` y creación personal. Los detalles técnicos pendientes están señalados en ese documento. Todavía no contiene una especificación OpenAPI validada ni endpoints implementados.

La [definición funcional de modificación](modificar-reserva.md) fue aprobada por el usuario el 2026-11-04, incluyendo `expectedVersion` y reemplazo completo de `attributes` cuando se envía.

## Inventario de contratos

| Operación | Método y ruta | Documento | Estado funcional |
|---|---|---|---|
| Crear reserva | `POST /api/v1/reservations` | [Creación](crear-reserva.md) | Aprobado; detalles técnicos pendientes señalados |
| Modificar reserva | `PATCH /api/v1/reservations/{id}` | [Modificación](modificar-reserva.md) | Aprobado; detalles técnicos pendientes señalados |
| Cancelar reserva | `POST /api/v1/reservations/{id}/cancel` | [Cancelación](cancelar-reserva.md) | Para revisión |
| Listar reservas propias | `GET /api/v1/reservations` | [Listado](listar-reservas.md) | Para revisión |
| Consultar reserva propia | `GET /api/v1/reservations/{id}` | [Detalle](consultar-reserva.md) | Para revisión |
| Listar recursos visibles | `GET /api/v1/resources` | [Recursos](listar-recursos.md) | Para revisión |
| Consultar disponibilidad | `GET /api/v1/resources/{id}/availability` | [Disponibilidad](consultar-disponibilidad.md) | Para revisión |

Las [convenciones comunes](convenciones.md) consolidan identidad, respuestas, fechas, errores y paginación. El usuario confirmó que todas las operaciones de reservas se limitan a sus propias reservas dentro de su ámbito. El 2026-11-04 aprobó los límites de página (20 por defecto, 100 máximo) y disponibilidad (31 días por consulta). Estos límites controlan la carga de datos por solicitud, no la cantidad de reservas ni la anticipación permitida para reservar.

Estos son contratos funcionales en Markdown. La siguiente etapa consiste en cerrar pendientes del alcance mock, trasladar a OpenAPI, validar y luego implementar mocks y Swagger UI. Los contratos de administración de recursos y bloqueos quedan para una etapa separada.

## Criterios acordados

- Definir el contrato antes de implementar handlers y clientes que lo consuman.
- Especificar solicitudes, respuestas, errores, autenticación, fechas, paginación, control de versión e idempotencia según las reglas acordadas.
- Separar esquemas de creación, modificación y lectura: el cliente no debe poder sustituir al responsable ni el plazo original de cancelación mediante una actualización.
- En creación, el responsable es siempre el usuario autenticado; no se recibe en el cuerpo ni se admite reservar para terceros inicialmente.
- Listado, detalle, modificación y cancelación se limitan a reservas propias en el ámbito autenticado. Las reservas ajenas no se exponen por sus identificadores.
- Incluir `attributes` en creación y respuesta para datos específicos de la reserva, con validación explícita por contexto. En el caso genérico puede omitirse en la solicitud o enviarse vacío.
- Representar explícitamente instantes y unidades de duración; no exponer directamente `time.Duration` ni tipos BSON como contrato público.
- Mantener el contrato versionado y comprobar su validez y compatibilidad en integración continua.
- Mapear los tipos de transporte a los modelos internos; los documentos MongoDB pertenecen al repositorio.

La versión OpenAPI, las herramientas de validación y la eventual generación de tipos o clientes están pendientes. No se incorporan archivos OpenAPI vacíos que aparenten describir una API existente.

Referencia: [especificación oficial OpenAPI](https://spec.openapis.org/oas/latest.html).
