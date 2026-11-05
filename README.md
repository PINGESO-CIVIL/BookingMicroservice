# BookingMS

Núcleo genérico de reservas de un único recurso exclusivo, con frontend **Vue**, contratos **OpenAPI** y backend **Go**. MongoDB y la lógica productiva corresponden al siguiente hito.

## Hito 1: demostración HTTP con mocks

Desde la raíz, con Docker Desktop iniciado en modo Linux:

```powershell
docker compose up --build -d --wait
```

- Frontend: http://127.0.0.1:5173
- Swagger: http://127.0.0.1:8080/docs/
- Salud de la API mock: http://127.0.0.1:4010/healthz

Compose compila e inicia Vue (servido por Nginx), la API mock Go y Swagger. No requiere Node ni Go instalados en el host para levantar la demo. Para detenerla: `docker compose stop`. `scripts/demo.ps1 Start|Status|Stop` es un acceso alternativo al mismo Compose, sin procesos locales separados.

## Auditorías manuales

Para generar cobertura y auditar el código, se necesitan también Node 22.22.2+ de la rama 22, Node 24.15+ de la rama 24 o Node 26+, y Go >= 1.24 en el host. Desde PowerShell en la raíz:

```powershell
npm --prefix Frontend ci
.\scripts\quality.ps1 Start
.\scripts\quality.ps1 Setup
.\scripts\quality.ps1 Check
.\scripts\quality.ps1 TrivyImages
```

`Check` genera cobertura Go y Vue/TypeScript, envía ambas a SonarQube y ejecuta Trivy sobre el repositorio. `TrivyImages` analiza las tres imágenes ya construidas por Compose. Las auditorías se ejecutan explícitamente por el usuario; no forman parte del arranque de la demo. Resultados en `reports/` y SonarQube en http://127.0.0.1:9000.

Después de actualizar dependencias o corregir imágenes, reconstruir las tres imágenes para incorporar las actualizaciones del sistema operativo de Nginx y Swagger. En Git Bash, desde la raíz:

```bash
npm --prefix Frontend ci
docker compose build --pull --no-cache
docker compose up -d --wait
powershell.exe -NoProfile -File ./scripts/quality.ps1 Check
powershell.exe -NoProfile -File ./scripts/quality.ps1 TrivyImages
```

Swagger se construye como `bookingms/swagger-ui:demo` a partir de la versión fijada y con paquetes del sistema actualizados; `TrivyImages` analiza esa imagen. No hay supresiones de vulnerabilidades. Los reportes anteriores describen las imágenes anteriores hasta volver a ejecutar las auditorías.

En Swagger, autorizar con `demo-user` o `demo-admin`. Son valores de demostración. Las escrituras devuelven ejemplos y no persisten cambios.

La [guía del Hito 1](docs/hito-1.md) incluye arquitectura, 14 operaciones HTTP, escenarios, guion de presentación, verificaciones y límites. La revisión/calificación formal de arquitectura queda pendiente del evaluador.

## Documentación

- [Frontend Vue](Frontend/README.md).
- [Backend y separación del mock](Backend/README.md).
- [Contratos OpenAPI](contracts/openapi/README.md).
- [Calidad y seguridad](docs/calidad-seguridad.md).
- [Revisión funcional de la interfaz](docs/revision-mockup-vue.md).

El Hito 2 reemplazará los mocks por lógica real y persistencia, con los umbrales de calidad correspondientes. Esta demo no acredita todavía exclusividad concurrente, idempotencia real ni autorización productiva.
