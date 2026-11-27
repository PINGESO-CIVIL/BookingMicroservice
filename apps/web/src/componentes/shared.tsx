import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Button } from '@base-ui/react/button'
import { useState, type ComponentProps, type ReactNode } from 'react'
import './styles.css'

export function AgendaBoundary({ children }: { children: ReactNode }) {
  const [client] = useState(() => new QueryClient({ defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false } } }))
  return <div className="agenda-ui"><QueryClientProvider client={client}>{children}</QueryClientProvider></div>
}
export function Action(props: ComponentProps<typeof Button>) {
  return <Button {...props} className={'agenda-button ' + (props.className ?? '')} />
}
export function Notice({ error, retry }: { error: unknown; retry?: () => void }) {
  if (!error) return null
  return <div className="agenda-notice" role="alert">{error instanceof Error ? error.message : 'Ocurrió un error.'} {retry && <Action className="agenda-button-secondary" onClick={retry}>Reintentar</Action>}</div>
}
