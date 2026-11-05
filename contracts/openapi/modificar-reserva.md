# Contrato de modificación de reserva — definición funcional aprobada

Estado: definición funcional aprobada por el usuario el 2026-11-04, incluyendo `expectedVersion` en el cuerpo y reemplazo completo de `attributes` cuando se envía. Los detalles técnicos señalados como pendientes siguen abiertos. Solo desarrolla la modificación; no implementa un endpoint ni constituye una especificación OpenAPI validada.

## Operación propuesta

`PATCH /api/v1/reservations/{id}`

Modificar parcialmente una reserva confirmada. Los campos omitidos conservan sus valores. Los cambios se aplican todos juntos o se rechaza toda la solicitud, conservando la reserva original.

La interfaz conserva el intervalo original por separado y permite elegir otro bloque de duración fija de `availableSlots` del recurso. No ofrece ingreso manual de horas ni expone reservas ajenas. La consulta pública sigue incluyendo la ocupación propia; conservar el intervalo original no lo convierte en disponible para otros usuarios. La escritura futura revalida el nuevo bloque excluyendo solo su propia ocupación.

El actor debe estar autenticado y autorizado sobre la reserva, su ámbito y el recurso resultante. El usuario confirmó que esta primera versión solo permite modificar reservas propias dentro de su ámbito. El contrato no permite cambiar responsable ni ámbito y no incorpora permisos administrativos. Una reserva ajena o inexistente devuelve `404`.

## Solicitud propuesta

Tipo de contenido: `application/json`. El cuerpo describe campos de negocio; no es una secuencia de operaciones JSON Patch.

| Campo | Requerido | Descripción |
|---|---|---|
| `expectedVersion` | Sí | Entero positivo que corresponde a la versión que el cliente consultó. Es una precondición, no un nuevo valor de versión. |
| `resourceId` | No | Nuevo recurso; siempre se mantiene exactamente un recurso por reserva. |
| `startAt` | No | Nuevo inicio en RFC 3339 con desplazamiento explícito o `Z`. |
| `endAt` | No | Nuevo término en RFC 3339 con desplazamiento explícito o `Z`. |
| `attributes` | No | Objeto de datos específicos validado contra el esquema del contexto. |

Debe proporcionarse al menos uno de los campos modificables además de `expectedVersion`. No se admite `null` para los campos de esta solicitud. Los campos desconocidos se rechazan.

```json
{
  "expectedVersion": 1,
  "startAt": "2026-11-10T20:00:00-03:00",
  "endAt": "2026-11-10T21:00:00-03:00",
  "attributes": {}
}
```

Semántica acordada para `attributes`: omitirlo conserva el objeto actual; enviarlo reemplaza el objeto completo. `{}` elimina los atributos anteriores únicamente si el esquema del contexto permite un objeto vacío. No se mezclan automáticamente claves ni objetos anidados.

Si solo se envía uno de los extremos del intervalo, se combina con el extremo persistido y se valida el intervalo resultante completo. No se calcula automáticamente una nueva duración.

`responsible`, `scopeId`, `status`, `version`, las condiciones de cancelación y las fechas de auditoría no son campos modificables. Intentar enviarlos rechaza la solicitud íntegra.

## Reglas acordadas

- La reserva debe estar confirmada y no haber comenzado. No se modifican reservas canceladas ni ya iniciadas.
- El inicio resultante debe ser futuro y anterior al término resultante.
- Para cambiar horario o recurso, primero se evalúa el límite de cancelación con el inicio persistido y el plazo original aceptado. La frontera es inclusiva: `ahora <= inicio anterior - plazo original`, siempre antes del inicio anterior.
- Si el límite venció, se rechaza cambiar horario o recurso aunque el nuevo inicio propuesto sea más lejano.
- Los cambios exclusivamente de atributos siguen permitidos antes del inicio según su esquema y permisos; no pueden afectar identidad, estado, ocupación ni políticas.
- La modificación no cambia al responsable ni el plazo original de cancelación ni la referencia a la política aceptada.
- Se valida la disponibilidad del intervalo y recurso resultantes, incluidos horarios y bloqueos, excluyendo únicamente la ocupación de esta misma reserva.
- Si cambia el recurso, se libera el anterior y se ocupa el nuevo de forma atómica. Ante un fallo se conserva íntegra la asignación anterior.
- Una versión desactualizada rechaza la modificación. No se sobrescriben cambios concurrentes.
- Si el inicio cambia válidamente, el nuevo límite se calcula con ese inicio y la misma duración original de cancelación.

Las demás políticas temporales aplicables siguen pendientes en DEC-004; este borrador no elige implícitamente entre políticas históricas y actuales para todas las restricciones.

## Respuesta propuesta

`200 OK`, con la representación actualizada de la reserva:

- Los campos definidos en la respuesta de [creación](crear-reserva.md), incluyendo `attributes`, responsable y plazo original.
- `version` incrementada para un cambio efectivo aceptado.
- `updatedAt` con el instante de modificación.
- `createdAt` conservado.
- `status` conservado como `confirmed`.
- `cancellationDeadline` calculado a partir del inicio vigente y el plazo original.

No se usa este endpoint para cancelar la reserva. La cancelación tendrá su propio contrato.

## Errores propuestos

| HTTP | Situación |
|---|---|
| `400` | Cuerpo incorrecto, versión esperada ausente o inválida, ningún campo modificable, `null` o campo no admitido. |
| `401` | Identidad no autenticada. |
| `403` | Modificación no autorizada. |
| `404` | Reserva o recurso inexistente o no visible para el solicitante. |
| `409` | Versión desactualizada, reserva cancelada o conflicto de disponibilidad. |
| `422` | Inicio ya alcanzado, límite vencido para cambiar horario/recurso, intervalo inválido o política/esquema de atributos incumplido. |

Los errores necesitarán códigos estables que distingan sus motivos y mantengan la privacidad de otros ámbitos. Los fallos de infraestructura no deben presentarse como errores de negocio.

## Criterios de aceptación

- Un cambio válido actualiza los campos enviados y conserva los omitidos.
- Intentar cambiar responsable o condiciones de cancelación rechaza todos los cambios de la solicitud.
- Un cambio fuera de plazo no puede habilitarse mediante un nuevo inicio futuro.
- Un fallo de disponibilidad conserva el recurso y el intervalo originales.
- Dos modificaciones concurrentes con la misma versión esperada no pueden sobrescribirse silenciosamente.
- Cambiar únicamente atributos no altera ocupación, responsable ni política aceptada.
- Enviar `attributes` reemplaza el objeto completo, omitirlo lo conserva y `{}` solo se acepta si cumple el esquema del contexto.

## Detalles técnicos pendientes

- Cerrar integración de identidad, errores, límites y esquemas en OpenAPI, respetando el acceso exclusivo a reservas propias.
- Definir tratamiento de una solicitud cuyos valores coinciden con los persistidos, y recuperación ante pérdida de la respuesta de una modificación confirmada. No reintentar a ciegas con una versión nueva.
- Determinar si una modificación requiere notificación; este documento no añade llamadas a G10.

Referencia: [diseño funcional del núcleo](../../spec/spec-design-nucleo-reservas.md), en particular REQ-019, REQ-020, REQ-031 y REQ-032.
