# Calidad y seguridad continua

## Estado y alcance

El stack es Go, Vue, OpenAPI y MongoDB. El usuario ha establecido SonarQube y Trivy como herramientas obligatorias para calidad y seguridad continua.

La configuración local se implementa con Docker Compose y `scripts/quality.ps1`. SonarQube Community Build utiliza un proyecto privado `bookingms-backend`; Trivy analiza el repositorio y conserva su reporte JSON. La plataforma de integración continua todavía no está definida: no se ha creado un pipeline ni una automatización periódica.

El backend aún contiene solo modelos. No hay pruebas de comportamiento; generar un archivo de cobertura o aprobar un Quality Gate no acredita por sí mismo las reglas de reservas.

### Verificación local del 2026-11-04

- Compose validado; SonarQube y su base quedaron saludables.
- `Setup` se volvió a ejecutar sin cambiar las credenciales válidas existentes.
- `Check` terminó con código 0: análisis de seis archivos Go, reporte de cobertura recibido y Quality Gate aprobado; Trivy reportó cero hallazgos.
- Se ejecutaron cero pruebas de comportamiento y el perfil no contiene líneas ejecutables: la cobertura funcional todavía no está acreditada.
- Trivy reconoció el manifiesto Go y no encontró formatos de configuración con reglas aplicables en este repositorio. El resultado no certifica por sí solo el archivo Compose.
- El repositorio todavía no tiene un primer commit; SonarQube advierte que no dispone de historial de autoría. No se creó un commit para ocultar esta limitación.

## Uso local

Requisitos: Docker Desktop iniciado en modo Linux, Docker Compose, Go y PowerShell. Ejecutar desde la raíz del repositorio:

```powershell
./scripts/quality.ps1 Init
./scripts/quality.ps1 Validate
./scripts/quality.ps1 Start
./scripts/quality.ps1 Setup
./scripts/quality.ps1 Check
```

- `Init`: genera credenciales aleatorias en `.quality/local.env` sin sobrescribir una configuración existente.
- `Validate`: valida Compose sin imprimir su configuración resuelta ni sus secretos.
- `Start`: inicia SonarQube y su base; espera a que estén saludables. El primer inicio puede tardar varios minutos.
- `Setup`: cambia la contraseña inicial de la instalación local, crea el proyecto privado y genera un token de análisis limitado al proyecto, con vencimiento de 90 días. Una nueva ejecución conserva un token que siga siendo válido. No modifica credenciales desconocidas de una instalación existente.
- `Check`: ejecuta SonarQube y Trivy, intenta ambos aunque falle uno y devuelve error si cualquiera falla. La importación de cobertura y el Quality Gate forman parte del análisis SonarQube.

Para ejecutar por separado:

```powershell
./scripts/quality.ps1 Coverage
./scripts/quality.ps1 Sonar
./scripts/quality.ps1 Trivy
./scripts/quality.ps1 Status
./scripts/quality.ps1 Stop
```

`Coverage` ejecuta `go vet` y `go test` con JSON y cobertura nativa Go; corrige las rutas del perfil para que coincidan con `Backend/` al analizar desde la raíz. Cada ejecución elimina reportes anteriores y un fallo de pruebas impide enviar cobertura como si fuera válida. No genera pruebas artificiales para elevar porcentajes.

SonarQube está disponible únicamente en [localhost:9000](http://localhost:9000). Usuario: `admin`. La contraseña está en `SONAR_ADMIN_PASSWORD` dentro del archivo local `.quality/local.env`; no se publica en documentación ni consola. El token y la contraseña de base también están en ese archivo, excluido de Git. No compartirlo ni copiarlo a configuración versionada. Si se pierde el archivo, conservar los volúmenes y recuperar el acceso: no borrar datos para regenerar contraseñas.

`Stop` detiene los servicios conservando los datos. No usar `docker compose down -v` ni eliminar los volúmenes si se desea conservar el historial de análisis.

## Componentes locales

| Componente | Versión configurada |
|---|---|
| SonarQube Community Build | `26.10.0.133078-community` |
| PostgreSQL exclusivo de SonarQube | `17.11-alpine3.23` |
| SonarScanner CLI en Docker | Imagen `12.3.0.4285_8.1.0` |
| Trivy | `0.76.1` |

Las imágenes están fijadas también por digest en `infra/quality/compose.yaml`. No se actualizan silenciosamente mediante `latest`. La base PostgreSQL es infraestructura interna de SonarQube y no expone un puerto al host; el microservicio de reservas sigue utilizando MongoDB. No se cambian ajustes globales de WSL ni se desactivan comprobaciones de Elasticsearch para arrancar.

El código se monta como solo lectura en los analizadores; únicamente `reports/` y las cachés necesarias son escribibles. No se monta el socket Docker en Trivy.

## Reportes y condiciones de fallo

| Archivo | Contenido |
|---|---|
| `reports/coverage.out` | Cobertura nativa generada por Go. |
| `reports/go-tests.json` | Eventos de ejecución de pruebas Go. |
| `reports/sonar/report-task.txt` | Referencia al análisis enviado al servidor. |
| `reports/trivy.json` | Hallazgos de Trivy, incluyendo severidades no bloqueantes. |

Los reportes y `.quality/` están excluidos de Git. Los reportes de seguridad pueden contener fragmentos sensibles si se detecta un secreto; compartirlos solo después de revisarlos.

- **SonarQube**: el scanner espera la evaluación del Quality Gate del proyecto (`sonar.qualitygate.wait=true`) y falla si no lo supera o no puede completar el análisis. El proyecto nuevo hereda el Quality Gate por defecto del servidor; el nombre y las condiciones se pueden consultar en la interfaz. No se fija un porcentaje ficticio de cobertura ni se excluye lógica de negocio de la medición.
- **Trivy**: se activan explícitamente vulnerabilidades, secretos y configuraciones soportadas. Se reportan todas las severidades; el comando falla por cualquier secreto y por vulnerabilidades/configuraciones `HIGH` o `CRITICAL`, incluso sin corrección disponible. Los hallazgos restantes se conservan para revisión. Un fallo del scanner, base de datos o formato del reporte también hace fallar el comando.
- `trivy.yaml` usa `exit-code: 0` para obtener el reporte completo. **El control bloqueante está en `quality.ps1 Trivy`**; ejecutar Trivy directamente con ese YAML sin evaluar el reporte no aplica este control.
- No se cargan archivos locales de supresiones ni reglas de secretos alternativos durante la ejecución configurada. No hay exclusiones de vulnerabilidades individuales.
- Se excluyen `.git`, `.quality`, reportes, `.scannerwork` y dependencias descargadas de frontend del recorrido del scanner; se mantienen los manifiestos y futuros archivos de dependencias bloqueadas para análisis. Las credenciales generadas en `.quality` son solo locales y nunca deben versionarse.

Actualmente no existe una imagen del microservicio ni un frontend Vue implementado. El alcance SonarQube se limita a `Backend/**/*.go`; su configuración de pruebas y cobertura Vue se incorporará cuando exista. Trivy solo analiza formatos que soporta: habilitar `misconfig` no garantiza que todos los archivos YAML, incluido Compose, tengan reglas aplicables.

## Verificaciones previstas

| Verificación | Objetivo | Momento previsto |
|---|---|---|
| Formato, compilación y análisis Go | Detectar problemas básicos antes de integrar | Cambios del backend |
| Pruebas Go y Vue | Verificar comportamiento; generar cobertura cuando existan pruebas | Cambios de código |
| Validación OpenAPI | Comprobar estructura del contrato y detectar cambios incompatibles | Cambios de contrato e integración |
| Integración MongoDB | Verificar atomicidad, concurrencia, idempotencia, versiones y aislamiento contra una instancia real | Cambios de persistencia y casos de uso |
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

## Consideraciones MongoDB

La estrategia de persistencia debe proteger la exclusividad entre reservas diferentes del mismo recurso y los cambios de disponibilidad concurrentes. Una transacción no debe tratarse como prueba automática de ausencia de solapamientos: se diseñará una coordinación por recurso u otra estrategia equivalente y se demostrará mediante pruebas concurrentes.

Si se adoptan transacciones multidocumento, los entornos de desarrollo, pruebas y VPS deberán proporcionar una topología MongoDB compatible. Un despliegue standalone no admite esas transacciones. La topología, persistencia, copias de seguridad y recuperación se diseñarán antes del despliegue; todavía no están configuradas.

## Referencias oficiales

- [SonarQube: Quality Gates](https://docs.sonarsource.com/sonarqube-server/quality-standards-administration/managing-quality-gates/introduction).
- [SonarQube: Docker](https://docs.sonarsource.com/sonarqube-community-build/server-installation/from-docker-image/set-up-and-start-container), [requisitos de base](https://docs.sonarsource.com/sonarqube-community-build/server-installation/installing-the-database) y [cobertura Go](https://docs.sonarsource.com/sonarqube-server/2025.3/analyzing-source-code/test-coverage/go-test-coverage).
- [Trivy: vulnerabilidades](https://github.com/aquasecurity/trivy/blob/main/docs/guide/scanner/vulnerability.md), [secretos](https://www.trivy.dev/docs/latest/guide/scanner/secret/) y [configuración de escáneres](https://github.com/aquasecurity/trivy/blob/main/docs/guide/configuration/skipping.md).
- [MongoDB: transacciones](https://www.mongodb.com/docs/manual/core/transactions/) y [consideraciones de producción y concurrencia](https://www.mongodb.com/docs/manual/core/transactions-production-consideration/).
