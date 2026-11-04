# Contrato de listado de recursos — borrador funcional

Estado: preparado para revisión. Consulta de apoyo al formulario; no añade administración de recursos. Sin implementación ni OpenAPI validado.

## Operación

`GET /api/v1/resources`

Devuelve los recursos que el usuario puede consultar en su ámbito. Listar un recurso no garantiza que esté libre en un horario; para ello se consulta disponibilidad y se vuelve a validar al reservar.

## Parámetros de consulta

| Parámetro | Requerido | Significado |
|---|---|---|
| `enabled` | No | Booleano; por defecto `true`, para alimentar el selector con recursos habilitados. `false` filtra los deshabilitados que sean visibles para el actor. |
| `limit` | No | Entero de 1 a 100; por defecto 20. Límites aprobados el 2026-11-04. |
| `cursor` | No | Continuación opaca de la misma consulta. |

No se permite elegir otro ámbito mediante la solicitud. El orden propuesto es `id` ascendente. No se habilitan filtros arbitrarios por `attributes` en esta fase.

## Respuesta

`200 OK`:

```json
{
  "items": [
    {
      "id": "cancha-futbol-001",
      "name": "Cancha de fútbol 1",
      "enabled": true,
      "timeZone": "America/Santiago",
      "attributes": {}
    }
  ],
  "nextCursor": null
}
```

`attributes` describe el recurso según un esquema de catálogo autorizado; no es el mismo objeto que los atributos de una reserva. Solo se exponen campos aptos para el usuario del formulario. No se exponen configuración interna de MongoDB, datos privados de otras reservas ni credenciales de sistemas externos.

El cliente utiliza `id` como `resourceId` al crear una reserva y `timeZone` para presentar fechas. Cambiar de página del catálogo no cambia las políticas de negocio. La propiedad e integración del catálogo real siguen pendientes; el mock usará datos sintéticos.

## Errores y aceptación

- `400 INVALID_REQUEST`: filtro, parámetro desconocido, límite o cursor inválidos.
- `401 UNAUTHENTICATED` / `403 FORBIDDEN`: según el acceso general.
- `500 INTERNAL_ERROR` / `503 SERVICE_UNAVAILABLE`: sin detalles internos.
- Sin recursos visibles, devolver `200` con `items: []` y `nextCursor: null`.
- El selector no permite descubrir recursos de otros ámbitos mediante filtros o cursores.
- Que un recurso esté habilitado no implica que todos sus horarios estén disponibles.

Definir los esquemas específicos del catálogo y sus límites antes de implementar cada adaptación. No se añade CRUD de recursos en este contrato.
