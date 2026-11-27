# Convenciones comunes de los contratos — propuesta funcional

Estado: complemento funcional del contrato `openapi.yaml`. Este Markdown no es código de mocks; las reglas descritas orientan la futura implementación real, mientras la fase actual solo define contratos y ejemplos.

## Alcance e identidad confirmados

Ampliación posterior del usuario: `GET /api/v1/admin/reservas` permite al administrador consultar reservas de todos los responsables de su ámbito. Es una operación separada de lectura; los endpoints personales siguen limitados a reservas propias. No se agregan escrituras administrativas sobre reservas ajenas.

- El usuario crea, lista, consulta, modifica y cancela únicamente sus propias reservas dentro de su ámbito autorizado.
- La identidad y el ámbito efectivos proceden del contexto autenticado; no se aceptan `responsible` ni `scopeId` como filtros o valores enviados para ampliar el acceso.
- Un identificador de reserva ajena o de otro ámbito responde `404`, igual que uno inexistente. La autorización se verifica antes de devolver estado, versión o datos de la reserva.
- El catálogo de recursos y la disponibilidad se consultan según los permisos del ámbito; la disponibilidad no revela datos de otros responsables.
- El mecanismo de autenticación real sigue pendiente. Los futuros mocks deberán identificar explícitamente su carácter simulado y no presentarse como autenticación de producción.

## Formatos propuestos

- Cuerpos JSON con campos en `camelCase`. Los campos de entrada desconocidos se rechazan.
- Identificadores opacos, no vacíos; no se impone UUID ni ObjectID al contrato público.
- Fechas de entrada en RFC 3339 con desplazamiento explícito o `Z`. Se comparan instantes, no cadenas. Las respuestas se normalizan a UTC con `Z`.
- Intervalos semiabiertos `[startAt, endAt)`, con inicio estrictamente anterior al término.
- Duraciones de la API expresadas en segundos enteros, no negativos. La precisión temporal admitida se cerrará al traducir a OpenAPI y PostgreSQL; no se permitirá truncamiento silencioso que altere una reserva.
- `attributes` siempre representa un objeto validado según el contexto, nunca `null`. El esquema genérico admite `{}`; las claves específicas se habilitan por esquema explícito.
- Las consultas GET no reciben cuerpo.

## Representación común de reserva

Creación, modificación, cancelación, detalle y elementos del listado utilizan esta representación. Es una propuesta de consolidación de los contratos ya aprobados, no una modificación de sus reglas.

| Campo | Tipo / significado |
|---|---|
| `id` | Identificador de la reserva. |
| `resourceId` | Único recurso reservado. |
| `responsible` | Objeto `{ "issuer": "...", "subject": "..." }` de la identidad externa responsable. |
| `startAt`, `endAt` | Intervalo vigente. |
| `status` | `confirmed` o `cancelled`. |
| `attributes` | Atributos validados; `{}` cuando no existen datos adicionales. |
| `cancellationNoticeSeconds` | Duración original aceptada, inmutable. |
| `cancellationDeadline` | Inicio vigente menos duración original. El instante exacto del inicio nunca admite cancelación, incluso si la duración es cero. |
| `version` | Entero positivo de control de concurrencia. |
| `createdAt` | Instante de creación, inmutable. |
| `updatedAt` | Instante del último cambio efectivo; en creación coincide con `createdAt`. |

En una reserva cancelada, `cancellationDeadline` conserva su significado histórico y no habilita una nueva transición. La fecha límite tampoco sustituye comprobar estado, permisos y versión al escribir.

## Paginación propuesta

Los listados devuelven `items` y `nextCursor`; `nextCursor` es `null` cuando no hay más resultados. Una consulta válida sin coincidencias devuelve `200` con `items: []`, no `404`.

Entradas: `limit` como entero positivo y `cursor` como cadena opaca opcional. El cursor solo sirve para continuar la misma consulta, con los mismos filtros y orden, y no concede autorización. Si es inválido o incompatible responde `400`. La identidad y el ámbito se vuelven a evaluar en cada petición.

Límites aprobados por el usuario el 2026-11-04: `limit` por defecto 20 y máximo 100. Se admiten enteros de 1 a 100; valores fuera de ese rango se rechazan con `400 INVALID_REQUEST`, sin recortarlos silenciosamente. Estos límites acotan los registros de una página, no el número de reservas que puede tener un usuario.

La primera versión no promete un total global ni una fotografía inmutable entre páginas cuando hay cambios concurrentes. El orden se define por endpoint; el cliente no interpreta ni construye el cursor.

## Errores comunes propuestos

```json
{
  "error": {
    "code": "RESERVATION_VERSION_CONFLICT",
    "message": "La reserva cambió; consulta su versión actual antes de continuar.",
    "details": [],
    "requestId": "request-001"
  }
}
```

- `code`: código estable para decisiones del cliente, sin depender del texto de `message`.
- `details`: lista opcional de errores públicos de campos, con `field` y `code`. No incluir cuerpos, tokens, consultas PostgreSQL ni datos de terceros.
- `requestId`: identificador de correlación generado o validado por el servidor.
- `400 INVALID_REQUEST`: forma, tipo, campo, filtro o cursor no admitido.
- `401 UNAUTHENTICATED`: no existe una identidad autenticada válida.
- `403 FORBIDDEN`: operación o acceso general no autorizado sin revelar entidades ajenas.
- `404 RESERVATION_NOT_FOUND` / `RESOURCE_NOT_FOUND`: entidad inexistente o no accesible.
- `409 RESERVATION_VERSION_CONFLICT`: versión esperada desactualizada.
- `409 RESOURCE_UNAVAILABLE`: colisión, bloqueo o indisponibilidad al intentar ocupar un intervalo.
- `409 RESERVATION_STATE_CONFLICT`: transición incompatible con el estado.
- `409 IDEMPOTENCY_KEY_CONFLICT`: misma clave con un contenido diferente.
- `422 INVALID_INTERVAL`, `RESERVATION_ALREADY_STARTED`, `CANCELLATION_WINDOW_CLOSED`, `POLICY_VIOLATION` o `ATTRIBUTES_INVALID`: incumplimiento de regla de negocio o esquema contextual.
- `503 SERVICE_UNAVAILABLE`: servicio temporalmente incapaz de completar la operación; no debe fingir un conflicto de negocio.
- `500 INTERNAL_ERROR`: fallo no previsto, sin detalles internos.

La precedencia entre validaciones se formalizará en OpenAPI y en los mocks; las reglas de cancelación idempotente se fijan en su contrato.

## Estado de publicación

El archivo `openapi.yaml` formaliza las catorce operaciones con OpenAPI 3.1 y se valida mediante Redocly CLI 2.75.0. Define formas, respuestas y ejemplos; no ejecuta reglas entre solicitudes. La autenticación Bearer describe el contrato sin presuponer JWT ni un proveedor real. En el contexto genérico `attributes` admite `{}`; las adaptaciones futuras deberán definir sus propios esquemas explícitos.

Con `API_ENV=dev`, el backend de Compose sirve Swagger en `http://127.0.0.1:5191/api/docs` y el contrato en `/api/docs/openapi.yaml`. **Try it out** llama al mismo origen. Se usan los valores Bearer de demostración `demo-user` y `demo-admin`. La cabecera `X-Mock-Scenario` selecciona respuestas de ejemplo. Las operaciones devuelven fixtures y la persistencia de reservas corresponde al Hito 2. Ver [instrucciones de validación y arranque](README.md).
