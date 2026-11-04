# Contrato de cancelación de reserva — borrador funcional

Estado: preparado para revisión. Las reglas temporales provienen del diseño aprobado; la forma concreta del contrato se propone aquí. Sin implementación ni OpenAPI validado.

## Operación

`POST /api/v1/reservations/{id}/cancel`

Cancela una reserva propia, dentro del ámbito autenticado, conservando su historial. No borra la reserva ni modifica recurso, intervalo, responsable o atributos.

## Solicitud

Tipo de contenido: `application/json`.

```json
{
  "expectedVersion": 1
}
```

`expectedVersion` es obligatorio, entero positivo y representa la versión consultada. No se admiten otros campos ni `null`. No se incorpora motivo de cancelación obligatorio ni se permite enviar una política alternativa.

## Reglas y orden de evaluación

1. Validar autenticación y formato; localizar una reserva propia del ámbito. Una reserva ajena o inexistente devuelve `404`.
2. Si ya está cancelada, devolver `200` con su representación actual, aunque el valor válido de `expectedVersion` sea anterior. No repetir transición, incremento de versión, liberación, auditoría de transición ni notificación. Tampoco volver a aplicar el plazo de una transición ya completada.
3. Si sigue confirmada, exigir coincidencia de versión; una versión anterior devuelve `409 RESERVATION_VERSION_CONFLICT`.
4. Evaluar el plazo original contra el inicio vigente: `ahora < inicio` y `ahora <= inicio - plazo original`. En el límite se permite; después del límite o desde el inicio se rechaza con `422`.
5. Cambiar a `cancelled`, liberar la ocupación y registrar el cambio de forma atómica. Incrementar la versión una sola vez y actualizar `updatedAt`.

Si una modificación concurre con una cancelación, no pueden aplicarse ambas sobre la misma versión original. La cancelación no libera un recurso si falla la escritura. La reserva cancelada no se reactiva.

## Respuesta

`200 OK` con la [representación común de reserva](convenciones.md), estado `cancelled`. No se devuelve `204` porque la respuesta incluye el estado y la versión resultantes.

## Errores

| HTTP | Código / situación |
|---|---|
| `400` | `INVALID_REQUEST`: cuerpo, versión o campo no admitido. |
| `401` | `UNAUTHENTICATED`. |
| `403` | `FORBIDDEN`: acceso general a la operación no autorizado. |
| `404` | `RESERVATION_NOT_FOUND`: reserva inexistente, ajena o de otro ámbito. |
| `409` | `RESERVATION_VERSION_CONFLICT`: reserva confirmada con otra versión. |
| `422` | `CANCELLATION_WINDOW_CLOSED` o `RESERVATION_ALREADY_STARTED`. |
| `500` / `503` | Fallo interno o indisponibilidad temporal, según las convenciones comunes. |

## Casos de aceptación

- Antes del límite, cancelar libera el recurso y conserva la reserva cancelada.
- En el límite exacto se permite; un instante posterior se rechaza sin efectos.
- Con plazo cero, en el inicio exacto se rechaza.
- Vencido el límite, la reserva sigue confirmada y mantiene su ocupación.
- Reintentar una cancelación completada devuelve la reserva cancelada sin repetir efectos, aunque la respuesta original se haya perdido.
- Una reserva ajena no revela su estado ni versión.

La necesidad de notificar una cancelación queda pendiente; no se añaden llamadas a G10. Si se incorpora notificación, no puede enviarse de nuevo por cada reintento.
