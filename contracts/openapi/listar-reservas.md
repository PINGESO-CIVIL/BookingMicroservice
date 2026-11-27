# Contrato de listado de reservas — borrador funcional

Estado: preparado para revisión. Sin implementación ni OpenAPI validado. El acceso solo a reservas propias está confirmado; los filtros y la paginación se proponen aquí.

## Operación

`GET /api/v1/reservas`

Lista únicamente reservas del usuario autenticado dentro de su ámbito autorizado, tanto confirmadas como canceladas salvo filtro de estado. No recibe cuerpo ni filtros para consultar otro responsable o ámbito.

## Parámetros de consulta

| Parámetro | Requerido | Significado |
|---|---|---|
| `resourceId` | No | Filtrar por un recurso concreto. |
| `status` | No | `confirmed` o `cancelled`; omitido incluye ambos. |
| `from` | Junto con `to` | Inicio de rango de búsqueda, RFC 3339 con zona explícita. |
| `to` | Junto con `from` | Término exclusivo del rango, posterior a `from`. |
| `limit` | No | Entero de 1 a 100; por defecto 20. |
| `cursor` | No | Continuación opaca de la misma consulta. |

Cuando se envía rango, se incluyen reservas cuyo intervalo se solapa con `[from, to)`: `startAt < to` y `endAt > from`. No solo las que comienzan dentro del rango. Enviar solo uno de sus extremos es inválido. Se permiten consultas de historial; la prohibición de crear reservas retroactivas no limita su lectura.

Sin rango, se consulta el historial autorizado con paginación. El orden propuesto es `startAt` ascendente y, en caso de empate, `id` ascendente, para dar un orden determinista. No se incluyen reservas de terceros aunque ocupen el recurso filtrado.

## Respuesta

`200 OK`:

```json
{
  "items": [],
  "nextCursor": null
}
```

Cada elemento utiliza la [representación común de reserva](convenciones.md), incluyendo `attributes`, versión y condiciones originales de cancelación. Un recurso sin coincidencias visibles produce una lista vacía; el listado no revela si existen reservas o recursos ajenos.

## Errores y aceptación

- `400 INVALID_REQUEST`: rango, estado, campo desconocido, tamaño de página o cursor inválidos.
- `401 UNAUTHENTICATED` / `403 FORBIDDEN`: según el acceso general.
- `500 INTERNAL_ERROR` / `503 SERVICE_UNAVAILABLE`: sin detalles internos.
- Una búsqueda válida sin resultados devuelve `200`, no `404`.
- Aplicar filtros y continuar páginas nunca amplía la identidad ni el ámbito.
- Una reserva que termina exactamente en `from` o comienza exactamente en `to` no se incluye.
- Reservas de otro usuario o ámbito no aparecen ni en resultados ni en metadatos.

El tamaño de página por defecto (20) y el máximo (100) fueron aprobados el 2026-11-04. El límite de 31 días corresponde a la consulta de disponibilidad; no se aplica implícitamente a este listado del historial.
