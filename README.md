# BookingMS · Microservicio de Agendamiento · PINGESO

Microservicio genérico de reservas de recursos exclusivos de la Línea de Productos de Software (**LPS**), construido sobre la plantilla técnica base del curso (arquitectura Clean Architecture, contrato OpenAPI 3.1 API-first, microcomponentes React autocontenidos y versiones fijadas). Permite consultar disponibilidad, crear, modificar y cancelar reservas, así como administrar recursos, horarios semanales y bloqueos, integrándose por contrato con **Identidad** (JWT/JWKS) y **Notificaciones**.

> **Estado actual (Hito 1):** el contrato OpenAPI, la infraestructura Compose, los microcomponentes y los controles de seguridad/calidad están activos; las 14 operaciones responden con datos simulados (fixtures). La lógica de exclusividad concurrente y persistencia real en PostgreSQL corresponden al **Hito 2**.

## Índice

- [Stack y estructura del repositorio](#stack-y-estructura-del-repositorio)
  - [Estructura de carpetas](#estructura-de-carpetas)
- [Guía para Desarrolladores](#guía-para-desarrolladores)
  - [1. Tecnologías utilizadas y su rol](#1-tecnologías-utilizadas-y-su-rol)
  - [2. ¿Qué debes descargar e instalar?](#2-qué-debes-descargar-e-instalar)
  - [3. ¿Cómo trabajar en el repositorio?](#3-cómo-trabajar-en-el-repositorio)
  - [4. SonarQube Cloud: Qué hace, cómo funciona y cómo ingresar](#4-sonarqube-cloud-qué-hace-cómo-funciona-y-cómo-ingresar)
- [Comandos principales (`Makefile` y `pnpm`)](#comandos-principales-makefile-y-pnpm)
- [Escáneres obligatorios: SonarQube y Trivy](#escáneres-obligatorios-sonarqube-y-trivy)
- [Documentación adicional](#documentación-adicional)

## Stack y estructura del repositorio

| Componente             | Tecnología y Versión Exacta                                                                                      | Ubicación                                |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| **Backend**            | Go `1.26.5` (`go 1.26.0` + `toolchain go1.26.5`), Huma `v2.34.1`, pgx `v5.10.0`, sqlc `v1.29.0`, goose `v3.24.1` | `apps/api/`                              |
| **Frontend**           | Node `22.22.0`, pnpm `10.13.1`, React `19.2.8`, TypeScript `6.0.3`, Vite `8.1.5`, Tailwind `4.3.3`               | `apps/web/`                              |
| **Microcomponentes**   | Paquete `@pingeso-civil/agenda-ui` (`Base UI 1.6.0`, `TanStack Query 5.101.4`, `React Hook Form 7.83.0`)         | `apps/web/src/componentes/`              |
| **Base de datos**      | PostgreSQL `17.11`                                                                                               | `docker-compose.yaml`                    |
| **Contrato API-first** | OpenAPI 3.1.0 generado desde structs de Huma y commiteado                                                        | `contracts/openapi.yaml`                 |
| **Contratos externos** | Cliente Go generado con `oapi-codegen v2.8.2`                                                                    | `contracts/externos/notificaciones.yaml` |
| **Modos de falla**     | Documentación de resiliencia, idempotencia, timeouts y verificación                                              | `docs/modos-de-falla.md`                 |

### Estructura de carpetas

```text
.
├── apps/
│   ├── api/
│   │   ├── cmd/api/            # main: arranque, config, slog JSON, señales, /api/docs
│   │   ├── cmd/openapi/        # exportador del contrato OpenAPI 3.1 desde Huma
│   │   ├── internal/api/       # handlers Huma + structs del contrato + JWT/JWKS + acceso_test.go
│   │   ├── internal/service/   # reglas de negocio, sin HTTP ni SQL
│   │   ├── internal/store/     # acceso a PostgreSQL, fixtures y cliente HTTP Notificaciones (5s)
│   │   ├── internal/repository/# generado por sqlc, no editar
│   │   ├── db/migrations/      # migraciones goose
│   │   ├── db/queries/         # consultas SQL para sqlc
│   │   ├── e2e/                # pruebas e2e + testdata/*.golden.md
│   │   ├── Dockerfile          # golang:1.26.5-alpine3.23 -> distroless static-debian13:nonroot
│   │   └── Dockerfile.migrate  # goose v3.24.1 sobre alpine:3.23.7
│   └── web/
│       ├── src/api/            # tipos OpenAPI (schema.d.ts), cliente openapi-fetch y fixtures
│       ├── src/componentes/    # microcomponentes React (<nombre>.tsx, <nombre>.api.ts, use-<nombre>.ts)
│       ├── src/features/       # app demo local con mocks MSW (no se publica)
│       └── tests/              # pruebas Vitest y E2E Playwright
├── contracts/
│   ├── openapi.yaml            # contrato OpenAPI 3.1 generado y commiteado
│   └── externos/               # contratos de servicios externos consumidos
├── docs/modos-de-falla.md
├── docker-compose.yaml
├── .env.example
├── .editorconfig
├── .oxlintrc.json
├── .prettierrc.json
├── lefthook.yml
├── sonar-project.properties
├── trivy.yaml
└── .github/workflows/
    ├── ci.yml                  # calidad, seguridad e imagen
    └── release.yml             # publicación GHCR, GitHub Packages, SBOM y Cosign
```

## Guía para Desarrolladores

### 1. Tecnologías utilizadas y su rol

#### Backend (`apps/api/`)

- **Go (`1.26.5`)**: Lenguaje principal de la API (`bookingms-api`) y del servidor estático/proxy del frontend (`apps/web/server.go`).
- **Huma (`v2.34.1`)**: Framework HTTP y generador del contrato **OpenAPI 3.1** directamente desde structs de Go (_API-first_).
- **pgx (`v5.10.0`)**: Driver de PostgreSQL.
- **sqlc (`v1.29.0`)**: Genera código Go tipado en `internal/repository/` desde `db/queries/*.sql`.
- **Goose (`v3.24.1`)**: Migraciones incrementales de esquema en `db/migrations/` (ejecutadas por el contenedor `migrate` antes de iniciar la API).
- **oapi-codegen (`v2.8.2`)**: Genera el cliente HTTP tipado en Go para servicios externos (`contracts/externos/notificaciones.yaml`).
- **golang-jwt (`v5.3.1`) + keyfunc (`v3.9.1`)**: Validación de JWT (`EdDSA` y `RS256`) contra el endpoint `AUTH_JWKS_URL` de Identidad.
- **Testcontainers (`v0.43.0`)**: Levanta PostgreSQL real en contenedores efímeros para pruebas de integración.

#### Frontend (`apps/web/`)

- **React (`19.2.8`) + TypeScript (`6.0.3`) + Vite (`8.1.5`)**: Base de los microcomponentes publicables (`@pingeso-civil/agenda-ui`) y de la demo local.
- **Tailwind CSS (`4.3.3`) + Base UI (`1.6.0`)**: Estilos componibles con tokens CSS `--agenda-*` y primitivas accesibles.
- **openapi-typescript (`7.13.0`) + openapi-fetch (`0.15.2`)**: Generación de tipos TypeScript (`src/api/schema.d.ts`) y cliente HTTP con `credentials: 'include'` y timeout de 5 s.
- **TanStack Query (`5.101.4`) + React Hook Form (`7.83.0`)**: Estado remoto y formularios.
- **MSW (`2.15.0`)**: Mocks de desarrollo local para montar y probar microcomponentes en aislamiento.
- **Vitest (`4.1.10`) + Playwright (`1.62.0`)**: Pruebas unitarias, de componentes y E2E en Chromium.

#### Base de Datos, Calidad, Seguridad y Contratos

- **PostgreSQL (`17.11`)**: Base de datos exclusiva de BookingMS.
- **Lefthook (`2.1.10`)**: Git hooks locales (`pre-commit` y `pre-push`).
- **Oxlint (`1.75.0`) + Prettier (`3.9.6`)**: Linter con `--deny-warnings` (bloquea `localStorage`/`sessionStorage`) y formateador (`semi: false`, `printWidth: 90`).
- **Gitleaks (`v8.30.1`)**: Escáner de secretos en `pre-commit` (`gitleaks protect --staged`) y en CI sobre todo el historial.
- **govulncheck (`v1.9.0`)**: Detecta vulnerabilidades alcanzables en la librería estándar de Go y dependencias de `go.mod`.
- **Trivy (`v0.76.1`)**: Escaneo de repositorio (`trivy fs`: vulnerabilidades, misconfigs y secretos) y de imágenes (`trivy image`), más generación de SBOM CycloneDX en releases.
- **Hadolint (`v2.15.1`)**: Linter de `Dockerfile` (falla en nivel `error`).
- **Dockle (`v0.4.15`)**: Auditoría de seguridad CIS de las imágenes construidas.
- **oasdiff (`v1.35.0`)**: Bloquea cambios incompatibles (_breaking changes_) en `contracts/openapi.yaml` respecto a la rama base.
- **Cosign (`v3.1.4`)**: Firma keyless (OIDC) de las imágenes publicadas en GHCR.
- **SonarQube Cloud**: Análisis de calidad, bugs, vulnerabilidades, security hotspots y cobertura sobre código nuevo.

---

### 2. ¿Qué debes descargar e instalar?

1. **Git**
2. **Docker, Docker Compose v2 y Docker Buildx** (Docker Desktop en modo Linux en Windows).
3. **GNU Make** (nativo en Linux/macOS; en Windows vía WSL2 o Git Bash/Make).
4. **Go `1.26.5`**: Requerido para compilar el backend y ejecutar `sqlc`, `goose`, `oapi-codegen`, `govulncheck` y `oasdiff`.
5. **Node.js `22.22.0`**: Fijado en `.nvmrc` (`nvm use`).
6. **pnpm `10.13.1`**:
   ```bash
   corepack enable && corepack prepare pnpm@10.13.1 --activate
   ```
7. **Gitleaks `v8.30.1`**: Obligatorio en `PATH` local para el hook `pre-commit`.
8. **Trivy `v0.76.1` y SonarScanner CLI** _(recomendados)_: Para ejecutar localmente los mismos escáneres de CI antes de abrir un PR.

---

### 3. ¿Cómo trabajar en el repositorio?

#### Paso 1: Configuración inicial (solo la primera vez)

```bash
# 1. Instalar dependencias de Node respetando el lockfile exacto
pnpm install --frozen-lockfile

# 2. Activar los Git Hooks locales (.npmrc tiene ignore-scripts=true por seguridad)
pnpm exec lefthook install

# 3. Descargar módulos de Go
cd apps/api && go mod download && cd ../..

# 4. Crear .env local desde .env.example y reemplazar BOOKING_DB_PASSWORD
cp .env.example .env
```

> Reemplaza `BOOKING_DB_PASSWORD=replace-with-local-database-password` por una contraseña local real antes de levantar Compose. El proceso de la API rechaza arrancar si alguna variable obligatoria conserva el prefijo `replace-with-`.

#### Paso 2: Levantar el entorno de desarrollo

```bash
docker compose up -d --build
docker compose ps -a
```

Compose inicia `postgres`, ejecuta `migrate` (Goose) hasta finalizar con código `0`, luego arranca `backend` y finalmente `frontend`.

**Endpoints locales:**

- **Frontend demo (React 19)**: http://127.0.0.1:5190
- **Documentación interactiva OpenAPI 3.1 (`API_ENV=dev` o `test`)**: http://127.0.0.1:5191/api/docs
- **Salud de la API (`/healthz`)**: http://127.0.0.1:5191/healthz
- **Readiness PostgreSQL (`/readyz`)**: http://127.0.0.1:5191/readyz

Para detener el entorno: `docker compose stop` (o `docker compose down --volumes` para limpiar el volumen).

#### Paso 3: Flujo de desarrollo diario

1. **Clean Architecture en Backend (`Handler → Service → Store`)**:
   - `internal/api/`: Solo traduce HTTP, valida entradas y define structs Huma. Nunca importa `store` ni `pgx` (verificado por `apps/api/cmd/api/architecture_test.go`).
   - `internal/service/`: Reglas de negocio puras, sin `net/http` ni SQL.
   - `internal/store/` + `internal/repository/`: Persistencia PostgreSQL y cliente externo `notificaciones`.
2. **Contrato API-first**:
   - Al editar structs o handlers en `apps/api/internal/api/`, ejecuta `make generate` (o `pnpm codegen`) para actualizar `contracts/openapi.yaml` y `apps/web/src/api/schema.d.ts`. Ambos archivos se commitean.
3. **Microcomponentes en Frontend**:
   - Cada microcomponente vive en `apps/web/src/componentes/<kebab-case>/` con `<nombre>.tsx`, `<nombre>.api.ts` y `use-<nombre>.ts`.
   - No se importan microcomponentes entre sí y está prohibido `localStorage`/`sessionStorage`.
   - La app demo local (`apps/web/src/features/`) no se incluye en el paquete npm `@pingeso-civil/agenda-ui`.

#### Paso 4: Verificación antes de abrir un Pull Request

```bash
# Linter (--deny-warnings), typecheck, tests y build de frontend
pnpm verify

# Pruebas completas de backend (incluye TestLayerImports, TestMatrizAcceso y TestObservabilityLogging)
pnpm test:api

# Verificar contrato OpenAPI y tipos sincronizados
make generate && git diff --exit-code

# Verificación HTTP y Swagger sobre Compose levantado
pnpm --dir apps/web run verify:http
pnpm --dir apps/web run verify:swagger
```

> Nunca subas versiones de dependencias a mano dentro de un PR de funcionalidad; las actualizaciones de versiones van en PRs propios (gestionados por Dependabot o por remediación documentada de seguridad).

---

### 4. SonarQube Cloud: Qué hace, cómo funciona y cómo ingresar

#### ¿Qué hace SonarQube Cloud?

Mientras Trivy y `govulncheck` vigilan dependencias externas y contenedores, **SonarQube Cloud** audita el código propio en `apps/api` y `apps/web/src`:

- **Bugs y Vulnerabilidades** en la lógica de negocio.
- **Code Smells** y duplicación (`< 3%`).
- **Security Hotspots** (100% revisados).
- **Cobertura en código nuevo** (`≥ 60%`, leída desde `reports/coverage.out` y `reports/frontend/lcov.info`).

#### ¿Cómo ingresar a SonarQube Cloud?

1. Ingresa a https://sonarcloud.io con tu cuenta de **GitHub**.
2. Accede al proyecto de la organización:
   - **Dashboard**: https://sonarcloud.io/project/overview?id=PINGESO-CIVIL_BookingMicroservice
   - **Organización (`sonar.organization`)**: `pingeso-civil`
   - **Project Key (`sonar.projectKey`)**: `PINGESO-CIVIL_BookingMicroservice`
3. En **Administration → Analysis Method**, mantener **Automatic Analysis** desactivado para permitir el análisis desde GitHub Actions (`CI-based Analysis`) con lectura de cobertura.

## Comandos principales (`Makefile` y `pnpm`)

- `make generate`: Ejecuta `sqlc`, `oapi-codegen` (contratos externos) y `pnpm codegen` (`contracts/openapi.yaml` + `schema.d.ts`).
- `make sqlc`: Regenera `apps/api/internal/repository/` con `sqlc@v1.29.0`.
- `make externos`: Regenera el cliente Go de Notificaciones con `oapi-codegen@v2.8.2`.
- `make migrate`: Aplica migraciones en `apps/api/db/migrations/` con `goose@v3.24.1`.
- `pnpm verify`: Ejecuta `lint` (`oxlint --deny-warnings`), `typecheck`, `test` y `build` de `apps/web`.
- `pnpm test:api`: Ejecuta `go test ./...` en `apps/api`.
- `pnpm --dir apps/web build:package`: Compila el paquete publicable `@pingeso-civil/agenda-ui`.

## Escáneres obligatorios: SonarQube y Trivy

- **Pipeline CI (`calidad`, `seguridad`, `imagen`)**: https://github.com/PINGESO-CIVIL/BookingMicroservice/actions/workflows/ci.yml
- **Proyecto SonarQube Cloud**: https://sonarcloud.io/project/overview?id=PINGESO-CIVIL_BookingMicroservice
- **GitHub Actions Secrets requeridos**: `SONAR_HOST_URL` (`https://sonarcloud.io`) y `SONAR_TOKEN`.

### Ejecución local previa a PR

**SonarQube** (con `SONAR_HOST_URL` y `SONAR_TOKEN` exportados en el entorno local):

```bash
sonar-scanner -Dsonar.host.url="$SONAR_HOST_URL" -Dsonar.token="$SONAR_TOKEN"
```

**Trivy (`v0.76.1`)**:

```bash
trivy fs --scanners vuln,misconfig,secret --severity HIGH,CRITICAL --ignore-unfixed --exit-code 1 .
docker build -t api:local -f apps/api/Dockerfile .
trivy image --scanners vuln,secret --severity HIGH,CRITICAL --ignore-unfixed --exit-code 1 api:local
docker build -t web:local -f apps/web/Dockerfile .
trivy image --scanners vuln,secret --severity HIGH,CRITICAL --ignore-unfixed --exit-code 1 web:local
```

## Documentación adicional

- [Frontend React y microcomponentes](apps/web/README.md)
- [Backend Go y separación del mock](apps/api/README.md)
- [Contratos OpenAPI](spec/contratos/README.md)
- [Modos de falla](docs/modos-de-falla.md)
