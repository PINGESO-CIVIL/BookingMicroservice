# Contrato de administración de recursos, horarios y bloqueos — borrador funcional

Estado: preparado para el Hito 1 (demostración de vista de administrador y respuestas mock). Cubre el alta, consulta administrativa, modificación de estado y políticas, configuración de horarios recurrentes y gestión de cierres o bloqueos excepcionales sobre un recurso del ámbito autorizado.

## Alcance y reglas de dominio vinculadas

- **Bloques de duración fija**: el usuario confirmó que cada recurso configura `slotDurationSeconds`, en minutos enteros. La interfaz ofrece esa duración junto al plazo de cancelación, y mantiene `minDurationSeconds` y `maxDurationSeconds` iguales al bloque al editarlo. Los horarios seleccionables se generan desde la apertura de cada franja, sin bloques parciales ni solapamientos con ocupación o cierres. Las reservas existentes conservan su horario y su plazo aceptado.

- **Propiedad de ámbito (`REQ-002`, `REQ-005`, `SEC-002`)**: Toda operación administrativa actúa exclusivamente sobre recursos del ámbito derivado del contexto autenticado con rol de administrador de ámbito. Un recurso de otro ámbito responde `404 RESOURCE_NOT_FOUND`.
- **Inmutabilidad del plazo aceptado en reservas previas (`REQ-031`)**: Actualizar `cancellationNoticeSeconds` o las políticas temporales de un recurso rige únicamente para las nuevas reservas que se confirmen a partir de ese cambio; las reservas ya confirmadas conservan su plazo original aceptado.
- **Protección frente a cambios conflictivos de disponibilidad (`REQ-012`, `AC-021`)**: Deshabilitar un recurso (`enabled: false`), recortar su horario recurrente o registrar un bloqueo temporal (`closure`) que impida cumplir reservas en estado `confirmed` se rechaza íntegramente con `409 RESOURCE_SCHEDULE_CONFLICT`, sin cancelar ni invalidar reservas silenciosamente.
- **Control de concurrencia (`REQ-020`)**: Las modificaciones del recurso y de su horario requieren `expectedVersion` para evitar sobrescrituras concurrentes inadvertidas (`409 RESOURCE_VERSION_CONFLICT`).

## Operaciones administrativas propuestas

| Operación | Método y ruta | Propósito |
|---|---|---|
| Crear recurso | `POST /api/v1/resources` | Registrar un nuevo recurso reservable en el ámbito con su zona horaria IANA, política de cancelación y referencia externa opcional. |
| Consultar detalle de recurso | `GET /api/v1/resources/{id}` | Obtener la configuración completa del recurso, su `version`, política vigente, horario semanal recurrente y bloqueos activos. |
| Modificar recurso | `PATCH /api/v1/resources/{id}` | Cambiar nombre, estado `enabled`, zona horaria, referencia externa, plazo de cancelación o `attributes` enviando `expectedVersion`. |
| Configurar horario y límites de tiempo | `PUT /api/v1/resources/{id}/schedule` | Reemplazar las reglas semanales recurrentes de apertura y límites de duración (`minDurationSeconds`, `maxDurationSeconds`) enviando `expectedVersion`. |
| Registrar bloqueo / cierre temporal | `POST /api/v1/resources/{id}/closures` | Crear un intervalo excepcional de cierre `[startAt, endAt)` por mantenimiento o evento administrativo. |
| Eliminar bloqueo / cierre temporal | `DELETE /api/v1/resources/{id}/closures/{closureId}` | Levantar un bloqueo temporal existente del recurso. |

## Representación administrativa del recurso (`ResourceDetail`)

Extiende la vista resumida de `GET /api/v1/resources` con los metadatos de gobierno del recurso:

```json
{
  "id": "resource-001",
  "name": "Recurso 1",
  "enabled": true,
  "timeZone": "America/Santiago",
  "externalReference": {
    "system": "catalogo-institucional",
    "id": "cat-rec-001"
  },
  "policy": {
    "cancellationNoticeSeconds": 7200,
    "minDurationSeconds": 3600,
    "maxDurationSeconds": 7200,
    "slotDurationSeconds": 3600
  },
  "weeklySchedule": [
    {
      "dayOfWeek": 1,
      "startLocalTime": "08:00",
      "endLocalTime": "22:00"
    }
  ],
  "closures": [
    {
      "id": "closure-001",
      "resourceId": "resource-001",
      "startAt": "2026-11-12T15:00:00Z",
      "endAt": "2026-11-12T18:00:00Z",
      "reason": "Mantenimiento programado",
      "createdAt": "2026-11-04T12:00:00Z"
    }
  ],
  "attributes": {},
  "version": 1,
  "createdAt": "2026-11-01T10:00:00Z",
  "updatedAt": "2026-11-04T12:00:00Z"
}
```

## Errores específicos de administración

- `400 INVALID_REQUEST`: cuerpo mal formado, `dayOfWeek` fuera de rango `1..7`, hora local inválida o intervalo de bloqueo invertido.
- `401 UNAUTHENTICATED` / `403 FORBIDDEN`: sin identidad autenticada o sin permisos de administración sobre el ámbito.
- `404 RESOURCE_NOT_FOUND` / `CLOSURE_NOT_FOUND`: recurso o bloqueo inexistente o perteneciente a otro ámbito.
- `409 RESOURCE_VERSION_CONFLICT`: la versión esperada (`expectedVersion`) del recurso no coincide con la versión vigente.
- `409 RESOURCE_SCHEDULE_CONFLICT`: la deshabilitación, el cambio de horario o el bloqueo solicitado entra en conflicto con reservas confirmadas existentes en el intervalo afectado (`REQ-012`).

Referencia: [diseño funcional del núcleo](../../spec/spec-design-nucleo-reservas.md).
