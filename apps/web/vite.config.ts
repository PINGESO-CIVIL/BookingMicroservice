import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
const proxy = { '/api': 'http://127.0.0.1:4010' }
export default defineConfig({ plugins: [vue()], server: { port: 5173, strictPort: true, proxy }, preview: { proxy } })
