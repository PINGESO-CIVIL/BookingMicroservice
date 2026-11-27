# Calidad y seguridad continua

## Estado y alcance

El stack es Go, Vue, OpenAPI y PostgreSQL. El usuario ha establecido SonarQube y Trivy como herramientas obligatorias para calidad y seguridad continua.

La configuración local se implementa con Docker Compose y `scripts/quality.ps1`. SonarQube analiza Go, TypeScript, Vue y los scripts del frontend en el proyecto privado existente `bookingms-backend` (se conserva la clave y su historial). Trivy analiza el repositorio y, mediante un comando independiente, las imágenes de la demo. La plataforma de integración continua todavía no está definida: no se ha creado un pipeline ni una automatización periódica.

El servicio real de reservas aún no está implementado. El mock HTTP y la interfaz sí tienen pruebas, pero su cobertura o un Quality Gate aprobado no acreditan reglas de negocio productivas. Esta ampliación prepara las auditorías para que las ejecute el usuario; no se han ejecutado escáneres nuevos como parte de la configuración.

### Registro previo: verificación del estado inicial del 2026-11-04

Los resultados de esta sección son anteriores al frontend y al servidor mock actuales. Se conservan como registro histórico y no son una auditoría del estado ampliado. Los comandos siguientes generan resultados nuevos.

- Compose validado; SonarQube y su base quedaron saludables.
- `Setup` se volvió a ejecutar sin cambiar las credenciales válidas existentes.
- `Check` terminó con código 0: análisis de seis archivos Go, reporte de cobertura recibido y Quality Gate aprobado; Trivy reportó cero hallazgos.
- Se ejecutaron cero pruebas de comportamiento y el perfil no contiene líneas ejecutables: la cobertura funcional todavía no está acreditada.
- Trivy reconoció el manifiesto Go y no encontró formatos de configuración con reglas aplicables en este repositorio. El resultado no certifica por sí solo el archivo Compose.
- El repositorio todavía no tiene un primer commit; SonarQube advierte que no dispone de historial de autoría. No se creó un commit para ocultar esta limitación.

## Uso local

Requisitos para auditorías de código: Docker Desktop iniciado en modo Linux, Docker Compose, Go >= 1.24, Node 22.22.2+ (rama 22), 24.15+ (rama 24) o 26+, y PowerShell. La demo por sí sola solo necesita Docker. Ejecutar desde la raíz del repositorio:

```powershell
./scripts/quality.ps1 Init
npm --prefix apps/web ci
./scripts/quality.ps1 Validate
./scripts/quality.ps1 Start
./scripts/quality.ps1 Setup
./scripts/quality.ps1 Check
# Después de construir las imágenes con docker compose up --build -d --wait:
./scripts/quality.ps1 TrivyImages
```

- `Init`: genera credenciales aleatorias en `.quality/local.env` sin sobrescribir una configuración existente.
- `Validate`: valida Compose sin imprimir su configuración resuelta ni sus secretos.
- `Start`: inicia SonarQube y su base; espera a que estén saludables. El primer inicio puede tardar varios minutos.
- `Setup`: cambia la contraseña inicial de la instalación local, crea el proyecto privado y genera un token de análisis limitado al proyecto, con vencimiento de 90 días. Una nueva ejecución conserva un token que siga siendo válido. No modifica credenciales desconocidas de una instalación existente.
- `Check`: compila Vue y genera cobertura de Go y Vue/TypeScript antes de SonarQube; ejecuta Trivy sobre el repositorio aunque falle Sonar y devuelve error si cualquiera falla. La importación de ambas coberturas y el Quality Gate forman parte del análisis SonarQube.
- `TrivyImages`: revisa `bookingms/api:demo` (incluye los archivos de Swagger), `bookingms/frontend:demo` y `bookingms/migrate:demo`. Exporta las imágenes con `docker image save`, analiza vulnerabilidades y secretos y conserva un reporte JSON por imagen. Aplica el mismo control bloqueante que el análisis del repositorio.

Para ejecutar por separado:

```powershell
./scripts/quality.ps1 Coverage
./scripts/quality.ps1 Sonar
./scripts/quality.ps1 Trivy
./scripts/quality.ps1 Status
./scripts/quality.ps1 Stop
```

`Coverage` ejecuta `go vet` y `go test` con JSON y cobertura nativa Go, compila Vue y ejecuta Vitest con cobertura V8/LCOV. Corrige las rutas de ambos perfiles para que coincidan con `apps/api/` y `apps/web/` al analizar desde la raíz. Se eliminan los perfiles anteriores antes de ejecutar; un fallo impide enviarlos como si fueran válidos. No se excluyen vistas sin pruebas ni se genera cobertura artificial. Las pruebas de navegador siguen siendo un comando separado y no se mezclan con LCOV de pruebas unitarias.

SonarQube está disponible únicamente en [localhost:9000](http://localhost:9000). Usuario: `admin`. La contraseña está en `SONAR_ADMIN_PASSWORD` dentro del archivo local `.quality/local.env`; no se publica en documentación ni consola. El token y la contraseña de base también están en ese archivo, excluido de Git. No compartirlo ni copiarlo a configuración versionada. Si se pierde el archivo, conservar los volúmenes y recuperar el acceso: no borrar datos para regenerar contraseñas.

`Stop` detiene los servicios conservando los datos. No usar `docker compose down -v` ni eliminar los volúmenes si se desea conservar el historial de análisis.

## Componentes locales

| Componente | Versión configurada |
|---|---|
| SonarQube Community Build | `26.10.0.133078-community` |
| PostgreSQL exclusivo de SonarQube | `17.11-alpine3.23` |
| SonarScanner CLI en Docker | Imagen `12.3.0.4285_8.1.0` |
| Trivy | `0.76.1` |

Las imágenes están fijadas también por digest en `infra/quality/compose.yaml`. No se actualizan silenciosamente mediante `latest`. La base PostgreSQL de SonarQube pertenece a su infraestructura interna. BookingMS usa un PostgreSQL independiente en el Compose raíz. No se cambian ajustes globales de WSL ni se desactivan comprobaciones de Elasticsearch para arrancar.

El código se monta como solo lectura en los analizadores; únicamente `reports/` y las cachés necesarias son escribibles. No se monta el socket Docker en Trivy.

## Reportes y condiciones de fallo

| Archivo | Contenido |
|---|---|
| `reports/coverage.out` | Cobertura nativa generada por Go. |
| `reports/go-tests.json` | Eventos de ejecución de pruebas Go. |
| `reports/go-tests.xml` | Ejecución Go convertida al formato genérico de SonarQube con rutas reales de archivos de prueba. |
| `reports/frontend/lcov.info` | Cobertura de TypeScript y componentes Vue para SonarQube. |
| `reports/frontend/coverage-summary.json` | Resumen de cobertura del frontend. |
| `reports/sonar/report-task.txt` | Referencia al análisis enviado al servidor. |
| `reports/trivy.json` | Hallazgos de Trivy, incluyendo severidades no bloqueantes. |
| `reports/images/*.json` | Hallazgos por imagen cuando se ejecuta `TrivyImages`. |

Los reportes y `.quality/` están excluidos de Git. Los reportes de seguridad pueden contener fragmentos sensibles si se detecta un secreto; compartirlos solo después de revisarlos.

- **SonarQube**: el scanner espera la evaluación del Quality Gate del proyecto (`sonar.qualitygate.wait=true`) y falla si no lo supera o no puede completar el análisis. El proyecto nuevo hereda el Quality Gate por defecto del servidor; el nombre y las condiciones se pueden consultar en la interfaz. No se fija un porcentaje ficticio de cobertura ni se excluye lógica de negocio de la medición.
- **Trivy**: se activan explícitamente vulnerabilidades, secretos y configuraciones soportadas, incluyendo dependencias de desarrollo como el validador OpenAPI. Se reportan todas las severidades; el comando falla por cualquier secreto y por vulnerabilidades/configuraciones `HIGH` o `CRITICAL`, incluso sin corrección disponible. Los hallazgos restantes se conservan para revisión. Un fallo del scanner, base de datos o formato del reporte también hace fallar el comando.
- `trivy.yaml` usa `exit-code: 0` para obtener el reporte completo. **El control bloqueante está en `quality.ps1 Trivy`**; ejecutar Trivy directamente con ese YAML sin evaluar el reporte no aplica este control.
- No se cargan archivos locales de supresiones ni reglas de secretos alternativos durante la ejecución configurada. No hay exclusiones de vulnerabilidades individuales.
- Se excluyen `.git`, `.quality`, reportes, `.scannerwork` y dependencias descargadas de frontend del recorrido del scanner; se mantienen los manifiestos y futuros archivos de dependencias bloqueadas para análisis. Las credenciales generadas en `.quality` son solo locales y nunca deben versionarse.
- El análisis del repositorio excluye también binarios locales anteriores, `dist` y salidas de pruebas de navegador. Los manifiestos y lockfiles siguen incluidos; `TrivyImages` analiza por separado los artefactos que realmente se construyeron para ejecutar la demo. No se agregaron excepciones de vulnerabilidades individuales.

El Compose raíz construye imágenes para Vue/Nginx, el backend Go con los archivos de Swagger y el proceso de migraciones Goose. SonarQube incluye Go, Vue, TypeScript y los scripts del frontend e importa cobertura Go y LCOV de Vitest. Los umbrales del código real corresponden al Hito 2.

## Verificaciones previstas

| Verificación | Objetivo | Momento previsto |
|---|---|---|
| Formato, compilación y análisis Go | Detectar problemas básicos antes de integrar | Cambios del backend |
| Pruebas Go y Vue | Verificar comportamiento; generar cobertura cuando existan pruebas | Cambios de código |
| Validación OpenAPI | Comprobar estructura del contrato y detectar cambios incompatibles | Cambios de contrato e integración |
| Integración PostgreSQL | Verificar atomicidad, concurrencia, idempotencia, versiones y aislamiento contra una instancia real | Cambios de persistencia y casos de uso |
| SonarQube | Analizar calidad y seguridad del código y evaluar su Quality Gate | Integración de cambios; publicación condicionada al resultado |
| Trivy sobre repositorio | Detectar vulnerabilidades en dependencias compatibles, secretos y configuraciones soportadas | Integración de cambios |
| Trivy sobre imágenes | Analizar la imagen que se pretende desplegar, cuando exista | Antes de publicar/desplegar esa imagen |

SonarQube y Trivy complementan las pruebas del dominio. No demuestran por sí mismos que no existan reservas solapadas.

## Criterios para configurar la automatización

- Definir proveedor de integración continua sin asumir GitHub Actions, GitLab CI u otro.
- La instancia local y sus versiones están configuradas. Antes de integrar un servidor corporativo, acordar su edición, lenguajes y Quality Gate, sin copiar credenciales locales.
- La condición local de Trivy ya bloquea secretos y hallazgos `HIGH`/`CRITICAL`. Cualquier excepción futura requiere motivo explícito y revisión; no ignorar hallazgos silenciosamente.
- Ejecutar explícitamente los escáneres Trivy necesarios; no asumir que todos están activos por defecto.
- Gestionar tokens y credenciales mediante secretos del sistema de integración; no escribirlos en el repositorio.
- Mantener las versiones y digests fijados, y revisarlos antes de reutilizar esta configuración en un pipeline.
- Conservar reportes y relacionarlos con la revisión del código y la imagen analizada. Un fallo de ejecución del analizador no equivale a un análisis aprobado.
- Determinar frecuencia de reevaluación de dependencias e imágenes existentes; no se ha creado ninguna automatización periódica.

## Persistencia de reservas

La demo usa PostgreSQL para Goose y para comprobar disponibilidad. El diseño del Hito 2 debe impedir reservas solapadas incluso con solicitudes concurrentes y verificar ese comportamiento contra PostgreSQL real.

## Referencias oficiales

- [SonarQube: Quality Gates](https://docs.sonarsource.com/sonarqube-server/quality-standards-administration/managing-quality-gates/introduction).
- [SonarQube: formato genérico de ejecución de pruebas](https://docs.sonarsource.com/sonarqube-server/analyzing-source-code/test-coverage/generic-test-data). La conversión conserva los eventos Go originales y falla si no puede asociar una prueba con su archivo; no inventa resultados.
- [SonarQube: cobertura JavaScript/TypeScript con LCOV](https://docs.sonarsource.com/sonarqube-server/analyzing-source-code/test-coverage/javascript-typescript-test-coverage) y [Vitest: generación de cobertura](https://vitest.dev/guide/coverage.html).
- [SonarQube: Docker](https://docs.sonarsource.com/sonarqube-community-build/server-installation/from-docker-image/set-up-and-start-container), [requisitos de base](https://docs.sonarsource.com/sonarqube-community-build/server-installation/installing-the-database) y [cobertura Go](https://docs.sonarsource.com/sonarqube-server/2025.3/analyzing-source-code/test-coverage/go-test-coverage).
- [Trivy: vulnerabilidades](https://github.com/aquasecurity/trivy/blob/main/docs/guide/scanner/vulnerability.md), [secretos](https://www.trivy.dev/docs/latest/guide/scanner/secret/) y [configuración de escáneres](https://github.com/aquasecurity/trivy/blob/main/docs/guide/configuration/skipping.md).
