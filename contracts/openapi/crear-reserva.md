# Contrato de creación de reserva — definición funcional aprobada

Estado: definición funcional aprobada por el usuario el 2026-11-04, incluyendo `attributes` y creación únicamente para el usuario autenticado. Esta aprobación corresponde al contrato funcional descrito; los detalles técnicos identificados como pendientes aún deben resolverse. Este documento no es una especificación OpenAPI validada ni implementa un endpoint. Solo desarrolla la creación; los demás contratos se definirán por separado.

## Operación propuesta

`POST /api/v1/reservations`

Crear una reserva confirmada de un único recurso para el usuario autenticado. La ruta se mantiene entre contextos; cambian los recursos, las políticas y los esquemas de atributos específicos.

Actualización de selección confirmada por el usuario: el intervalo se elige de los bloques de duración fija del recurso (`policy.slotDurationSeconds`), publicados como `availableSlots` en disponibilidad. El formulario no solicita escribir horas. La futura escritura debe revalidar duración, alineación con la apertura, horarios, cierres y ocupación; consultar un bloque no lo retiene.

## Identidad: decisión confirmada

- El responsable siempre es el usuario autenticado que crea la reserva.
- El backend obtiene su referencia de identidad del contexto de autenticación validado, no de campos enviados por el cliente.
- El ámbito también procede del contexto autorizado.
- No se admite reservar para otra persona en la primera versión.
- El mecanismo de autenticación y su representación en OpenAPI están pendientes; este contrato no presupone un proveedor ni un formato de token.

## Solicitud propuesta

Tipo de contenido: `application/json`.

Cabecera `Idempotency-Key` obligatoria para identificar reintentos de la misma operación. El alcance, formato, longitud y retención de la clave siguen pendientes en DEC-008 de la especificación funcional.

| Campo | Requerido | Descripción |
|---|---|---|
| `resourceId` | Sí | Identificador opaco del único recurso solicitado. |
| `startAt` | Sí | Instante de inicio en RFC 3339 con desplazamiento explícito o `Z`. |
| `endAt` | Sí | Instante de término en RFC 3339 con desplazamiento explícito o `Z`. |
| `attributes` | Según contexto | Objeto con datos adicionales de la reserva, validado contra el esquema del contexto. En el caso genérico puede omitirse o ser `{}`. |

```json
{
  "resourceId": "cancha-futbol-001",
  "startAt": "2026-11-10T19:00:00-03:00",
  "endAt": "2026-11-10T20:00:00-03:00",
  "attributes": {}
}
```

`responsible`, `scopeId`, `status`, `version` y las condiciones de cancelación no son campos de entrada. Los campos ajenos al esquema se rechazan; no se ignoran silenciosamente los intentos de sustituir al responsable.

`attributes` forma parte del contrato base y permite incorporar datos propios de una reserva en cada contexto. Sus claves, tipos, campos obligatorios y límites deben definirse mediante un esquema explícito antes de implementar la adaptación. No es un objeto libre sin validación. Estos atributos no pueden alterar identidad, ámbito, permisos, estado, ocupación ni las condiciones de cancelación.

Por ejemplo, una adaptación deportiva podría admitir `"attributes": {"activityId": "basquetbol"}` si ese campo está definido en su esquema. No se incorpora `activityId` como campo obligatorio del núcleo genérico.

## Resultado exitoso propuesto

`201 Created` para una creación exitosa.

| Campo | Contenido |
|---|---|
| `id` | Identificador de la reserva. |
| `resourceId` | Recurso reservado. |
| `responsible` | Referencia al responsable obtenida de la identidad autenticada. |
| `startAt`, `endAt` | Intervalo confirmado. |
| `status` | `confirmed`. |
| `attributes` | Datos específicos aceptados y validados para la reserva; objeto vacío si no hay datos adicionales. |
| `cancellationNoticeSeconds` | Plazo original aceptado expresado en segundos; unidad propuesta para revisar en OpenAPI. |
| `cancellationDeadline` | Límite calculado como inicio menos plazo original. Siempre se exige además estar estrictamente antes del inicio. |
| `version` | Versión inicial utilizada en futuras escrituras concurrentes. |
| `createdAt` | Instante de creación. |

Falta cerrar el esquema exacto de `responsible`, los formatos de identificadores y la respuesta de un reintento idempotente. El cliente no controla ninguno de los campos calculados.

## Errores propuestos

| HTTP | Situación |
|---|---|
| `400` | Cuerpo o formato incorrecto, campo no admitido o cabecera obligatoria ausente. |
| `401` | Identidad no autenticada. |
| `403` | Operación no autorizada. |
| `404` | Recurso inexistente o no visible para el solicitante. |
| `409` | Colisión de disponibilidad o clave de idempotencia reutilizada con otro contenido. |
| `422` | Intervalo o política de negocio incumplidos. |

La estructura de error debe permitir distinguir motivos mediante códigos estables. Los fallos de infraestructura y su política de reintento se especificarán antes de implementar; no deben presentarse como rechazos de negocio.

## Comportamiento esperado

1. Obtener identidad y ámbito autorizados y validar la solicitud, incluido `attributes` según el esquema del contexto.
2. Resolver y validar el recurso y las políticas aplicables.
3. Evaluar el intervalo, horarios, bloqueos y colisiones.
4. Confirmar de forma atómica reserva, ocupación y condiciones originales de cancelación, con trazabilidad e idempotencia.
5. En la fase de mocks, confirmar por consola la simulación de la notificación de correo, sin realizar llamadas a G10 ni enviar correos reales. La semántica ante fallos de notificación se definirá antes de implementar este flujo.

El contrato no garantiza exclusividad por existir un esquema válido: la implementación y sus pruebas deberán demostrar ese comportamiento.

## Validaciones de aceptación prioritarias

- Una creación válida registra al usuario autenticado como responsable.
- Intentar proporcionar otro responsable o ámbito rechaza la solicitud sin efectos parciales.
- El recurso exclusivo no admite intervalos solapados; sí admite intervalos consecutivos si cumplen las demás reglas.
- Un reintento identificado no produce una segunda reserva.
- El plazo original queda conservado y no puede reemplazarse desde el cliente.
- Los atributos admitidos se conservan en la reserva y se incluyen en la respuesta; campos o tipos no permitidos se rechazan.

Referencia: [diseño funcional del núcleo](../../spec/spec-design-nucleo-reservas.md).
