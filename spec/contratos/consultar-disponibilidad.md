# Contrato de disponibilidad de un recurso — borrador funcional

Estado: preparado para revisión. Sin implementación ni OpenAPI validado. Esta consulta alimenta el calendario y no retiene ni confirma una reserva.

## Operación

`GET /api/v1/recursos/{id}/disponibilidades`

## Entrada

- `id`: identificador del recurso visible para el actor en su ámbito.
- `from` y `to`: parámetros obligatorios RFC 3339 con zona explícita, que delimitan `[from, to)`, con `from < to`.
- Sin cuerpo de solicitud. No se acepta un identificador de reserva para excluir ocupación en esta primera consulta.
- Máximo de 31 días por consulta, aprobado el 2026-11-04. Este límite acota la carga del calendario y no define el horizonte de anticipación para reservar. Para navegar a otro período se realiza una nueva consulta con otro rango.

## Semántica propuesta

La respuesta contiene ventanas continuas libres dentro del rango solicitado, calculadas a partir de horarios habilitados menos bloqueos y reservas confirmadas del recurso. Se consideran también las reservas de otros usuarios que ocupan ese recurso, sin revelar sus datos.

Las ventanas están ordenadas por inicio, no se solapan y no exceden el rango. Las ventanas contiguas se agrupan cuando no existe ninguna separación de disponibilidad. Las reservas canceladas no descuentan ocupación.

Decisión del usuario: las reservas se seleccionan como **bloques de duración fija configurables por recurso**, sin escribir manualmente inicio y término. `slotDurationSeconds` define esa duración (mínimo 60 segundos, múltiplo de 60). `availableWindows` conserva su significado de ventana continua; `availableSlots` contiene los bloques completos que puede seleccionar la interfaz.

Los bloques se alinean desde la apertura de cada franja semanal y deben caber íntegramente antes del cierre, dentro del rango consultado y en el futuro. Una ocupación o bloqueo que solape parcialmente un bloque elimina ese bloque completo; no desplaza su alineación. Un remanente menor a la duración no genera una opción adicional. Se aplican horarios y zona del recurso, y no se muestran opciones en días cerrados o recursos deshabilitados. La selección no retiene el recurso y la escritura debe volver a validarla.

Si el recurso visible está deshabilitado, la consulta devuelve una lista vacía con `enabled: false`. No confundirlo con un recurso inexistente o ajeno, que devuelve `404`.

En edición, esta consulta incluye la ocupación de la reserva existente; la interfaz debe conservar su selección actual al presentar el formulario. La validación real de modificación sí excluye su propia ocupación conforme a su contrato. No se debe añadir un parámetro arbitrario para ignorar reservas ajenas.

## Respuesta

`200 OK`, por ejemplo:

```json
{
  "resourceId": "cancha-futbol-001",
  "enabled": true,
  "timeZone": "America/Santiago",
  "from": "2026-11-10T21:00:00Z",
  "to": "2026-11-11T01:00:00Z",
  "slotDurationSeconds": 3600,
  "availableSlots": [
    {"startAt": "2026-11-10T21:00:00Z", "endAt": "2026-11-10T22:00:00Z"},
    {"startAt": "2026-11-10T23:00:00Z", "endAt": "2026-11-11T00:00:00Z"},
    {"startAt": "2026-11-11T00:00:00Z", "endAt": "2026-11-11T01:00:00Z"}
  ],
  "availableWindows": [
    {
      "startAt": "2026-11-10T21:00:00Z",
      "endAt": "2026-11-10T22:00:00Z"
    },
    {
      "startAt": "2026-11-10T23:00:00Z",
      "endAt": "2026-11-11T01:00:00Z"
    }
  ]
}
```

El intervalo [22:00, 23:00) puede estar ocupado o bloqueado; la respuesta no revela el motivo privado, la reserva ni su responsable. Si no hay ventanas libres se devuelve `availableWindows: []`, no un error. No se pagina para evitar omitir silenciosamente partes del rango; se aplica el máximo de 31 días a la consulta completa.

## Errores y aceptación

- `400 INVALID_REQUEST`: extremos ausentes, formato inválido, rango vacío/invertido, parámetro desconocido o rango superior a 31 días.
- `401 UNAUTHENTICATED` / `403 FORBIDDEN`: según el acceso general.
- `404 RESOURCE_NOT_FOUND`: recurso inexistente o no visible.
- `500 INTERNAL_ERROR` / `503 SERVICE_UNAVAILABLE`: sin detalles internos.
- La ocupación de otros usuarios se descuenta, aunque esos usuarios no aparezcan en el listado de reservas propias.
- Dos intervalos consecutivos no se tratan como una colisión.
- La consulta no modifica reservas, bloqueos, versiones ni estado.
- Que una ventana aparezca libre no obliga a aceptar una creación posterior si otra operación la ocupa antes.

Los fixtures se generan previamente a partir de configuración, reservas y bloqueos de ejemplo; el mock HTTP Go los sirve sin escrituras persistentes ni garantías de concurrencia en PostgreSQL. React consume estos datos por HTTP y no calcula la disponibilidad durante el uso. El backend real debe implementar y validar estas reglas. En edición se conserva además el horario original como opción propia; no se incorpora al listado público de horarios libres ni habilita otras reservas ocupadas.
