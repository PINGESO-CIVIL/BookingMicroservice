# Backend de reservas

Servicio de reservas en Go con arquitectura por capas: `model`, `service` y `handler`. Actualmente contiene los modelos iniciales del dominio; todavía no tiene servidor ejecutable ni casos de uso implementados.

Stack acordado: Go para backend, Vue para frontend, MongoDB para persistencia y OpenAPI para contratos HTTP. SonarQube y Trivy tienen configuración local mediante Docker y scripts; la integración con una plataforma de integración continua sigue pendiente.

## Directorios

```text
Backend/
├── cmd/
│   └── api/                  # Punto de entrada y composición de dependencias
├── internal/
│   ├── config/               # Lectura y validación de configuración
│   ├── model/                # Entidades, tipos y reglas propias del dominio
│   ├── service/              # Casos de uso e interfaces de persistencia requeridas
│   ├── handler/              # Entrada/salida HTTP y traducción de errores
│   ├── middleware/           # Comportamiento transversal de HTTP
│   └── repository/
│       └── mongodb/          # Documentos, índices y operaciones atómicas
├── migrations/               # Evolución versionada de documentos y configuración de datos
└── README.md
```

Los archivos `.gitkeep` permiten versionar los directorios vacíos. Se eliminarán al incorporar archivos reales en cada directorio.

## Responsabilidades y dependencias

El recorrido de una solicitud será `middleware → handler → service → persistencia`. El sentido de las dependencias del código se define por estas reglas:

- **model**: contiene entidades, tipos de valor y validaciones del dominio. No depende de HTTP, MongoDB ni de las otras capas. No se limita a representar documentos de una colección.
- **service**: implementa los casos de uso y coordina políticas, permisos de negocio y operaciones atómicas. Depende de `model` y declara las interfaces que necesita de la persistencia. No conoce consultas MongoDB, respuestas HTTP ni implementaciones concretas del repositorio.
- **handler**: interpreta solicitudes, valida su formato, llama a los servicios y traduce resultados y errores a HTTP conforme a OpenAPI. No contiene reglas de disponibilidad ni accede directamente a MongoDB. Los tipos de transporte se separan de las entidades del dominio.
- **middleware**: resuelve comportamientos transversales de HTTP, como identificación de solicitudes, recuperación de errores y autenticación cuando se defina su integración. Entrega contexto autenticado a los handlers; no sustituye las autorizaciones de negocio del servicio ni depende de handlers concretos.
- **repository/mongodb**: implementa los contratos de persistencia requeridos por `service`, traduce documentos BSON a modelos y protege las operaciones mediante mecanismos atómicos y, cuando corresponda, transacciones. No depende de `handler` ni de `middleware`. El diseño de concurrencia debe impedir solapamientos entre documentos distintos; consultar disponibilidad y luego insertar no constituye una protección suficiente.
- **config**: carga, interpreta y valida configuración de ejecución. No contiene reglas de reservas ni credenciales incorporadas al código. Las políticas configurables de negocio pertenecen al dominio y sus casos de uso, no se convierten automáticamente en variables de entorno.
- **cmd/api**: carga configuración, construye dependencias, conecta persistencia, servicios, handlers y middleware, y administra el inicio y cierre del servidor. Es el punto de composición, sin lógica de negocio.
- **migrations**: contendrá cambios versionados de documentos, validadores e índices cuando se apruebe el diseño de persistencia, sin asumir migraciones SQL. No contiene migraciones todavía.

Los paquetes bajo `internal` son propios de este backend. La reutilización entre líneas de negocio se realiza mediante el servicio y sus contratos, sin exigir que otros proyectos importen paquetes Go internos.

## Convenciones iniciales

- Usar nombres de paquetes Go en minúsculas: `model`, `service`, `handler`, `middleware`, `config` y `mongodb`.
- Organizar archivos por concepto dentro de cada capa cuando se implemente el dominio; no crear paquetes ni interfaces sin necesidad concreta.
- Declarar interfaces donde se consumen. No crear una interfaz de repositorio genérico con operaciones CRUD por anticipado.
- Colocar las pruebas junto al paquete que verifican, en archivos `*_test.go`. No se crean pruebas ni directorios de pruebas vacíos en esta etapa.
- Mantener un núcleo común. No crear carpetas con copias de servicios para deporte, medicina o estudio.
- No guardar secretos ni archivos locales de configuración con credenciales en el repositorio.

## Estado del diseño

Las reglas acordadas y las decisiones funcionales pendientes se encuentran en [la especificación del núcleo](../spec/spec-design-nucleo-reservas.md).

## Modelos iniciales

| Tipo | Propósito |
|---|---|
| `Resource` | Recurso exclusivo con ámbito, habilitación, zona horaria y referencia opcional al catálogo externo. |
| `Reservation` | Reserva de un único recurso, responsable, intervalo, estado, plazo aceptado y versión. |
| `IdentityReference` | Referencia externa por emisor (`Issuer`) y sujeto (`Subject`), sin duplicar perfiles ni credenciales. |
| `Interval` | Instantes de inicio y fin con semántica `[inicio, fin)`. |
| `ReservationStatus` | Estados `confirmed` y `cancelled`; sin estados pendientes ni expiración automática. |
| `CancellationPolicySnapshot` | Plazo original (`Notice`) y referencia a la versión de política aceptada al confirmar. |
| `ExternalReference` | Identificador del recurso dentro de un catálogo externo opcional. |

Los modelos no incluyen etiquetas JSON ni BSON. Los identificadores son cadenas opacas; esto no fija UUID, ObjectID ni un formato de almacenamiento. Los documentos del repositorio definirán el mapeo BSON. `Notice` es una duración Go, no una unidad ni formato público de API o almacenamiento. El servicio calculará el límite como `Interval.Start - Notice` y garantizará que el plazo aceptado no cambie.

`Attributes` contiene datos específicos del contexto; su esquema y validación se implementarán explícitamente antes de admitir escrituras. No permite cambiar estados, permisos, políticas u ocupación por medio de campos adicionales. Los mapas deben copiarse al aplicar cambios para no modificar accidentalmente la versión original.

Los campos exportados son representación de datos, no una garantía de validez. Los servicios deberán proteger responsable y plazo de cancelación, validar estados e intervalos y coordinar transacciones; `Version` por sí solo no impide escrituras concurrentes. No se añaden métodos de creación, modificación ni cancelación en esta etapa.

No se modelan todavía horarios recurrentes, excepciones, selección de políticas ni contratos de auditoría: sus detalles pendientes se mantienen en la especificación. Tampoco se crea una entidad de asignación separada, porque `Reservation.ResourceID` representa la relación uno a uno acordada.

## Verificación local

Desde `Backend`, ejecutar `go test ./...` y `go vet ./...`. Por ahora verifican compilación y análisis estático; no existen pruebas de comportamiento ni garantías de concurrencia implementadas.

El módulo local se llama `bookingms` y utiliza únicamente la biblioteca estándar. Su ruta podrá ajustarse cuando se defina el repositorio remoto. El código de arranque, las dependencias externas, los contratos HTTP y las migraciones se definirán en sus etapas correspondientes.

## Contratos y calidad continua

El directorio compartido [contracts/openapi](../contracts/openapi/README.md) contiene el contrato OpenAPI 3.0.3 entre Go y Vue, con siete operaciones y ejemplos validables. Los modelos del dominio no sustituyen los esquemas de solicitudes y respuestas. Esta fase se limita a contrato y visualización con mocks prefijados; los servicios reales Go y la persistencia MongoDB corresponden al siguiente hito. Swagger UI está publicado en `http://localhost:8080/docs/`; los endpoints mock todavía no están levantados.

La [guía de calidad y seguridad](../docs/calidad-seguridad.md) describe el servidor local SonarQube, Trivy, generación de cobertura y condiciones de fallo. Desde la raíz, `./scripts/quality.ps1 Check` ejecuta ambas verificaciones una vez iniciada y preparada la instalación local. No existe todavía un pipeline porque su plataforma no está definida. Generar cobertura en esta etapa de modelos no equivale a contar con pruebas funcionales.
