# BookingMS

Núcleo genérico de reservas de un único recurso exclusivo, con frontend **React**, contratos **OpenAPI** y backend **Go**. La lógica real de reservas y su persistencia en PostgreSQL corresponden al Hito 2.

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

Para personalizar el entorno, copiar `.env.example` a `.env` y asignar la contraseña local antes de crear el volumen.

Para detener la demo, ejecutar `docker compose stop`.

### Verificación del contrato

Con la demo levantada, ejecutar desde la raíz:

```powershell
pnpm --dir apps/web ci
pnpm --dir apps/web run verify:http
pnpm --dir apps/web run verify:swagger
```

El primer verificador comprueba las 14 operaciones y que las escrituras simuladas dejan intactas las consultas. El segundo usa Edge para comprobar Swagger y el proxy del frontend. Los informes quedan en `reports/demo/`.

## Verificación técnica

Instalar Node y pnpm según `.nvmrc` y `package.json`, y ejecutar `pnpm install --frozen-lockfile` desde la raíz. `make generate` regenera OpenAPI, sus tipos, el cliente de Notificaciones y el repositorio sqlc. CI verifica los artefactos generados, la compatibilidad con oasdiff, las pruebas, SonarQube y las imágenes con Trivy.

## Documentación

- [Frontend React](apps/web/README.md).
- [Backend y separación del mock](apps/api/README.md).
- [Contratos OpenAPI](spec/contratos/README.md).
- [Modos de falla](docs/modos-de-falla.md).

El Hito 2 reemplazará los mocks por lógica real y persistencia, con los umbrales de calidad correspondientes. Esta demo no acredita todavía exclusividad concurrente, idempotencia real ni autorización productiva.
