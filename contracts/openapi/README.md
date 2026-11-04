# Contrato HTTP de reservas

Este directorio contiene el [contrato OpenAPI 3.0.3](openapi.yaml) compartido por el futuro backend Go y el frontend Vue. Incluye las siete operaciones y ejemplos prefijados de solicitudes y respuestas. Se valida con Redocly CLI, incluida la compatibilidad de los ejemplos con sus esquemas. Swagger UI está publicado localmente; todavía no hay servidor mock ni lógica de negocio implementados.

La [definición funcional de creación](crear-reserva.md) fue aprobada por el usuario el 2026-11-04, incluyendo `attributes` y creación personal. Los detalles de lógica real identificados como pendientes se mantienen para el siguiente hito.

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

Los documentos Markdown explican el comportamiento esperado; `openapi.yaml` formaliza las formas públicas y los ejemplos para la demostración. No se adelantan servicios, handlers ni repositorios Go. El servidor de mocks responderá con escenarios prefijados: las escrituras no persistirán ni modificarán las consultas posteriores. Los contratos de administración de recursos y bloqueos quedan para una etapa separada.

## Validación reproducible

Desde la raíz del repositorio:

```powershell
cd contracts
npm ci --ignore-scripts
npm run validate
```

Requisitos: Node >= 22.12.0 y npm >= 10. Redocly CLI está fijado en `package.json` y `package-lock.json`. Se validan estructura OpenAPI, referencias, identificadores únicos de operaciones y ejemplos de cuerpos y parámetros. El archivo de configuración únicamente exceptúa el aviso de servidor localhost por ser una demostración local y la licencia, que aún no se ha definido; no se silencian errores de esquemas ni ejemplos.

Validar el contrato no ejecuta comparaciones temporales ni demuestra idempotencia, autorización, colisiones o persistencia. La documentación se publica como se indica abajo; la siguiente entrega será el servidor de mocks.

## Swagger UI local

- Documentación: [http://localhost:8080/docs/](http://localhost:8080/docs/).
- Contrato servido: [http://localhost:8080/docs/openapi.yaml](http://localhost:8080/docs/openapi.yaml).
- Contiene las siete operaciones, parámetros, cuerpos, ejemplos, códigos de respuesta y esquemas.
- Las llamadas de **Try it out** están desactivadas hasta levantar el servidor de mocks previsto en `localhost:4010`. Consultar ejemplos no ejecuta ni guarda reservas.

Desde la raíz del repositorio:

```powershell
docker compose up -d --wait
docker compose ps
```

El `compose.yaml` de la raíz publica solamente Swagger UI en localhost y utiliza una imagen fijada por versión y digest. No inicia backend, MongoDB ni frontend Vue. La instalación de calidad mantiene su Compose y sus datos independientes.

El contrato se monta como solo lectura y la imagen de Swagger lo copia al directorio público al iniciar. Después de editar y validar el YAML, ejecutar `docker compose restart swagger-ui` y recargar el navegador. Para detener la documentación, `docker compose stop swagger-ui`.

La validación se realiza localmente con Redocly; Swagger UI no envía el contrato al validador público de swagger.io. La dirección local servida fue verificada contra el archivo original y sus siete operaciones se comprobaron en el navegador.

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

Se ha seleccionado OpenAPI 3.0.3 y Redocly CLI 2.75.0 para esta entrega. La generación de clientes y tipos, el servidor de mocks y su publicación se definirán en las siguientes entregas, sin cambiar el alcance a lógica real.

Referencia: [especificación oficial OpenAPI](https://spec.openapis.org/oas/latest.html).
