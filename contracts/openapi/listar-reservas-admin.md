# Consulta administrativa de reservas

Ampliación solicitada por el usuario para el Hito 1: el administrador debe consultar las reservas realizadas por todos los usuarios de su ámbito, sin mostrar acciones de creación de reservas en esa vista.

`GET /api/v1/admin/reservas`, operación `listAdminReservations`, requiere rol administrador de ámbito. No recibe un ámbito elegido por el cliente ni concede acceso a otras organizaciones. En el mock se utiliza `Bearer demo-admin`; `demo-user` responde `403`.

La respuesta es `ReservationPage`, con recurso, responsable (`issuer` y `subject`), intervalo, estado y los demás campos de lectura. Incluye confirmadas y canceladas, con filtros opcionales `resourceId`, `status`, `from` y `to`. Los extremos temporales se envían juntos y seleccionan solapamientos. Orden: `startAt`, luego `id`, ascendentes. Paginación por cursor, 20 elementos por defecto, máximo 100.

El listado y su detalle visual son de consulta. No se agregan operaciones administrativas para crear reservas para terceros, modificar o cancelar reservas ajenas. Los endpoints personales conservan su alcance exclusivo al responsable autenticado.

Los ejemplos del mock incluyen tres responsables del mismo ámbito. Las escrituras siguen sin persistencia. La implementación de aislamiento e identidad productiva se realizará en el backend real.
