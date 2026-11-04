# Contrato de consulta de una reserva — borrador funcional

Estado: preparado para revisión. Sin implementación ni OpenAPI validado. El acceso solo a reservas propias está confirmado.

## Operación

`GET /api/v1/reservations/{id}`

Obtiene el detalle vigente de una reserva propia del ámbito autenticado, confirmada o cancelada. No recibe cuerpo ni produce cambios.

## Entrada y respuesta

- `id`: identificador opaco no vacío de la reserva.
- `200 OK`: [representación común de reserva](convenciones.md), incluyendo `attributes`, responsable, estado, intervalo, versión y condiciones originales de cancelación.
- `version` permite construir la siguiente solicitud de modificación o cancelación con `expectedVersion`.
- El plazo de cancelación visible no sustituye validar nuevamente el estado, el tiempo y la versión en una escritura posterior.

Ejemplo ilustrativo de una reserva confirmada:

```json
{
  "id": "reservation-001",
  "resourceId": "cancha-futbol-001",
  "responsible": {
    "issuer": "identidad-demo",
    "subject": "usuario-demo-001"
  },
  "startAt": "2026-11-10T22:00:00Z",
  "endAt": "2026-11-10T23:00:00Z",
  "status": "confirmed",
  "attributes": {},
  "cancellationNoticeSeconds": 7200,
  "cancellationDeadline": "2026-11-10T20:00:00Z",
  "version": 1,
  "createdAt": "2026-11-04T12:00:00Z",
  "updatedAt": "2026-11-04T12:00:00Z"
}
```

El plazo de dos horas es un dato ilustrativo, no un plazo obligatorio del núcleo.

## Errores y aceptación

- `400 INVALID_REQUEST`: identificador o parámetros no admitidos.
- `401 UNAUTHENTICATED` / `403 FORBIDDEN`: según el acceso general.
- `404 RESERVATION_NOT_FOUND`: inexistente, ajena o de otro ámbito, sin distinguir esos casos.
- `500 INTERNAL_ERROR` / `503 SERVICE_UNAVAILABLE`: sin detalles internos.
- Consultar una reserva cancelada devuelve `200` con `status: cancelled`, conservando su historial.
- Consultar no actualiza versión, ocupación, estado ni condiciones de cancelación.
