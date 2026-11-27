# Microcomponentes React de Agenda

`@pingeso-civil/agenda-ui` exporta `ListaReservas`, `FormularioReserva`, `DetalleReserva`, `ListaRecursos` y `EditorRecurso`. La versión del contrato consumida por el paquete se declara en `agendaApiVersion` en el manifiesto.

Cada componente recibe `baseUrl` como origen o prefijo anterior a `/api/v1`, `token` sin el prefijo Bearer, los permisos opcionales en `permissions` y, cuando corresponde, un identificador y callbacks para navegar en la aplicación anfitriona. El anfitrión debe permitir el acceso a Agenda mediante un proxy propio o una configuración de CORS apropiada.

```tsx
import { ListaReservas } from '@pingeso-civil/agenda-ui'
import '@pingeso-civil/agenda-ui/styles.css'

<ListaReservas baseUrl="https://agenda.example" token={token} onOpen={abrirReserva} />
```

Desde la raíz, instalar con `pnpm install --frozen-lockfile` y ejecutar `pnpm --dir apps/web dev`. La app local monta los componentes con MSW. Con `VITE_MOCK_API=false` usa el proxy de Vite hacia el servidor Go, que se inicia con `go -C apps/api run ./cmd/api`.

`pnpm --dir apps/web build:package` compila la librería. `pnpm --dir apps/web pack` crea el archivo instalable para probarlo de forma local. La publicación remota depende del registro acordado por el curso.
