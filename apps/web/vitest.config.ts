import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
export default defineConfig({
  plugins: [react()],
  test: {
    include: ['tests/**/*.test.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      include: ['src/**/*.ts', 'src/**/*.tsx'],
      exclude: ['src/**/*.d.ts'],
      reportsDirectory: '../../reports/frontend',
      reporter: ['text-summary', 'lcov', 'json-summary'],
      thresholds: { statements: 80, branches: 80, functions: 80, lines: 80 },
      reportOnFailure: false,
    },
  },
})
