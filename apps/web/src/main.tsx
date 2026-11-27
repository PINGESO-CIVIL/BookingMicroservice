import { createRoot } from 'react-dom/client'
import { HashRouter } from 'react-router'
import { AgendaDemo } from './features/agenda/agenda-demo'
import './features/agenda/demo.css'

async function mount() {
  if (import.meta.env.DEV && import.meta.env.VITE_MOCK_API !== 'false') {
    const { startMocks } = await import('./features/demo/mocks')
    await startMocks()
  }
  createRoot(document.getElementById('app')!).render(<HashRouter><AgendaDemo /></HashRouter>)
}
void mount()
