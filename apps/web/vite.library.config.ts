import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
export default defineConfig({
  publicDir: false,
  plugins: [react(), tailwindcss()],
  build: {
    outDir: 'package-dist',
    lib: { entry: 'src/componentes/index.ts', formats: ['es'], fileName: 'index', cssFileName: 'styles' },
    rolldownOptions: { external: id => /^(react|react-dom)(\/|$)/.test(id) },
  },
})
