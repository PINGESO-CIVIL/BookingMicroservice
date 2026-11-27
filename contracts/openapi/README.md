# Contrato HTTP de reservas

Actualización 0.3.0: el usuario confirmó bloques de duración fija configurables por recurso. `ResourcePolicy` incluye `slotDurationSeconds` (minutos enteros expresados en segundos); disponibilidad incluye esa duración y `availableSlots`, manteniendo `availableWindows` como ventanas continuas. La interfaz selecciona bloques completos y no solicita escribir horas. El frontend calcula los bloques sobre configuración y ocupación de ejemplo, sin persistencia; el backend real sigue pendiente.

El [contrato OpenAPI 3.1](../openapi.yaml) es la fuente de las catorce operaciones, sus tipos y los ejemplos del mock. Desde la raíz, instalar las dependencias con `npm ci --prefix apps/web` y ejecutar `pnpm codegen` para regenerar los tipos TypeScript y los fixtures. Para validar el YAML y sus ejemplos, usar `npm ci --prefix contracts` y `npm run validate --prefix contracts`.

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

## Validación reproducible

Desde la raíz del repositorio:

```powershell
cd contracts
npm ci --ignore-scripts
npm run validate
```

Requisitos: Node >= 22.12.0 y npm >= 10. Redocly CLI está fijado en `package.json` y `package-lock.json`. Se validan estructura OpenAPI, referencias, identificadores únicos de operaciones y ejemplos de cuerpos y parámetros. El archivo de configuración únicamente exceptúa el aviso de servidor localhost por ser una demostración local y la licencia, que aún no se ha definido; no se silencian errores de esquemas ni ejemplos.

Validar el contrato no ejecuta comparaciones temporales ni demuestra idempotencia, autorización, colisiones o persistencia. La documentación se publica como se indica abajo; el servidor de mocks se inicia con `./scripts/demo.ps1 Start` desde la raíz.

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

La API incluye el YAML y los archivos estáticos de Swagger en su imagen. Después de editar y validar el YAML, ejecutar `docker compose up -d --build backend` y recargar el navegador. `API_ENV=dev` habilita los archivos de documentación. El valor predeterminado de Compose es `dev`.

La validación se realiza localmente con Redocly; Swagger UI no envía el contrato al validador público de swagger.io. La dirección local servida fue verificada contra el archivo original y sus catorce operaciones se comprobaron en el navegador.

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

Se ha seleccionado OpenAPI 3.1 y Redocly CLI 2.75.0 para esta entrega. Los tipos TypeScript se generan desde OpenAPI y Vue consume el mock HTTP Go. La publicación es local; un despliegue remoto debe acordarse. La lógica real corresponde al Hito 2.

Referencia: [especificación oficial OpenAPI](https://spec.openapis.org/oas/latest.html).
