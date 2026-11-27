# Modos de falla de BookingMS

El Hito 1 usa un contrato OpenAPI y datos simulados. Los controles de esta tabla aplican al backend, al frontend de la demo y a su entorno local.

| Falla | Control | Cómo se verifica | Estado |
| --- | --- | --- | --- |
| PostgreSQL inicia después del migrador | Compose espera el healthcheck de PostgreSQL antes de ejecutar Goose | Levantar Compose con un volumen nuevo y comprobar el orden de estados | Hito 1 |
| Goose falla o pierde su historial al recrear el entorno | El backend espera que el migrador termine con código 0. Goose usa `public.goose_db_version` y PostgreSQL conserva un volumen propio | Verificar que una migración fallida bloquea el backend en un proyecto aislado. Recrear el entorno conservando el volumen y comparar la versión aplicada | Hito 1 |
| PostgreSQL se cae durante la ejecución | `/readyz` consulta PostgreSQL con un plazo de dos segundos. `/healthz` comprueba el proceso HTTP | Detener PostgreSQL y comprobar `/readyz = 503` y `/healthz = 200`. Reiniciar PostgreSQL y comprobar `/readyz = 200` con el mismo proceso HTTP | Hito 1 |
| Vue abre antes de que la API esté lista o pierde su conexión | Compose espera el healthcheck del backend y el cliente Vue muestra el error | Levantar la demo desde cero, comprobar el proxy en el navegador y ejecutar el escenario de recuperación del frontend | Hito 1 |
| La documentación se publica fuera de desarrollo | El enrutador habilita Swagger, los archivos estáticos y el YAML con `API_ENV=dev` | Verificar Swagger y **Try it out** en desarrollo. En `prod`, comprobar 404 para Swagger y 200 para una consulta a la API | Hito 1 |
| Una entrada tiene UUID inválido, campos desconocidos o fechas inválidas | El handler valida tipos, cuerpo, ruta y consulta antes de procesar la operación | Ejecutar las pruebas Go de entradas inválidas y comprobar 400 | Hito 1 |
| Falta la identidad de demostración o el rol carece de permisos | El servicio comprueba el rol simulado y limita las operaciones administrativas | Probar una consulta sin credenciales y comparar la misma operación administrativa con `demo-user` y `demo-admin` | Hito 1 |
| No existe la reserva, el recurso o el bloqueo consultado | El servicio comprueba la existencia en los fixtures y el handler traduce el resultado a 404 | Ejecutar las pruebas de identificadores inexistentes o el escenario `notFound` | Hito 1 |
| Un cursor se reutiliza con otros filtros, identidad o escenario | El servicio vincula el cursor con la consulta que lo originó | Pedir la primera página y reutilizar su cursor con un filtro distinto. La API devuelve 400 | Hito 1 |
| Una escritura de demostración altera las consultas posteriores | El store de fixtures ofrece lecturas y el servicio responde con ejemplos | Comparar las consultas antes y después de crear, modificar y cancelar. Ejecutar `verify:http` para comprobar el contrato | Hito 1 |
| El contrato y las respuestas simuladas divergen | El verificador valida las 14 combinaciones de método y ruta contra OpenAPI | Ejecutar Redocly y `npm --prefix apps/web run verify:http` sobre Compose | Hito 1 |
| Dos usuarios solicitan el mismo horario, se repite una operación o vence la cancelación | Implementar restricciones transaccionales, idempotencia y políticas temporales | Probar concurrencia, reintentos y vencimientos contra PostgreSQL real y comparar con un golden | Hito 2 |
| Un proveedor externo falla o responde tarde | Confirmar proveedores con el documento de componentes. Añadir clientes por API con timeout y propagación del contexto | Simular indisponibilidad y comprobar el error y el estado resultante cuando exista la integración | Pendiente de confirmar dependencias |
