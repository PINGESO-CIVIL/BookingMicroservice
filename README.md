# BookingMS

Núcleo genérico de reservas de un único recurso exclusivo, con frontend **Vue**, contratos **OpenAPI** y backend **Go**. La lógica real de reservas y su persistencia en PostgreSQL corresponden al Hito 2.

## Hito 1: demostración HTTP con mocks

Desde la raíz, con Docker Desktop iniciado en modo Linux:

```powershell
docker compose up -d --build
docker compose ps -a
```

- Frontend: http://127.0.0.1:5190
- Swagger: http://127.0.0.1:5191/api/docs
- Salud de la API: http://127.0.0.1:5191/healthz
- Disponibilidad de PostgreSQL: http://127.0.0.1:5191/readyz

Compose inicia `postgres`, ejecuta `migrate` una vez y espera su salida con código 0 antes de arrancar `backend`. Cuando el backend está listo, inicia `frontend`. El volumen `bookingms_data` conserva la base y la versión de Goose. PostgreSQL se conecta por la red interna y las 14 operaciones siguen usando fixtures.

El backend sirve Swagger con `API_ENV=dev`, el valor predeterminado de Compose. Con `API_ENV=prod`, las rutas de documentación responden 404. `BOOKING_DB_PASSWORD` permite definir la contraseña al inicializar el volumen. El valor predeterminado es para esta demo local.

Para detener la demo, ejecutar `docker compose stop`. `scripts/demo.ps1 Start|Status|Stop` usa el mismo Compose.

### Verificación del contrato

Con la demo levantada, ejecutar desde la raíz:

```powershell
npm --prefix apps/web ci
npm --prefix apps/web run verify:http
npm --prefix apps/web run verify:swagger
```

El primer verificador comprueba las 14 operaciones y que las escrituras simuladas dejan intactas las consultas. El segundo usa Edge para comprobar Swagger y el proxy de Vue. Los informes quedan en `reports/demo/`.

## Auditorías manuales

Para generar cobertura y auditar el código, se necesitan también Node 22.22.2+ de la rama 22, Node 24.15+ de la rama 24 o Node 26+, y Go >= 1.24 en el host. Desde PowerShell en la raíz:

```powershell
npm --prefix apps/web ci
.\scripts\quality.ps1 Start
.\scripts\quality.ps1 Setup
.\scripts\quality.ps1 Check
.\scripts\quality.ps1 TrivyImages
```

`Check` genera cobertura Go y Vue/TypeScript, envía ambas a SonarQube y ejecuta Trivy sobre el repositorio. `TrivyImages` analiza las imágenes de backend, frontend y migraciones construidas por Compose. Las auditorías se ejecutan por separado. Los resultados quedan en `reports/` y SonarQube está en http://127.0.0.1:9000.

Después de actualizar dependencias o corregir imágenes, reconstruir las imágenes antes de evaluar los resultados. En Git Bash, desde la raíz:

```bash
npm --prefix apps/web ci
docker compose build --pull --no-cache
docker compose up -d --wait
powershell.exe -NoProfile -File ./scripts/quality.ps1 Check
powershell.exe -NoProfile -File ./scripts/quality.ps1 TrivyImages
```

Los archivos estáticos de Swagger se copian desde la versión fijada al construir el backend. Los reportes de auditoría corresponden a las imágenes analizadas en cada ejecución.

En Swagger, autorizar con `demo-user` o `demo-admin`. Son valores de demostración. Las escrituras devuelven ejemplos y no persisten cambios.

El backend separa HTTP, casos de uso y acceso a datos en `apps/api`. La app Vue de transición vive en `apps/web`. La revisión formal de arquitectura corresponde al evaluador.

## Documentación

- [Frontend Vue](apps/web/README.md).
- [Backend y separación del mock](apps/api/README.md).
- [Contratos OpenAPI](contracts/openapi/README.md).
- [Calidad y seguridad](docs/calidad-seguridad.md).
- [Modos de falla](docs/modos-de-falla.md).

El Hito 2 reemplazará los mocks por lógica real y persistencia, con los umbrales de calidad correspondientes. Esta demo no acredita todavía exclusividad concurrente, idempotencia real ni autorización productiva.
