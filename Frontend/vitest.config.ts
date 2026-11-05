import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
export default defineConfig({
  plugins: [vue()],
  test: {
    include: ['tests/**/*.test.ts'],
    coverage: {
      provider: 'v8',
      include: ['src/**/*.ts', 'src/**/*.vue'],
      exclude: ['src/**/*.d.ts'],
      reportsDirectory: '../reports/frontend',
      reporter: ['text-summary', 'lcov', 'json-summary'],
      thresholds: { statements: 80, branches: 80, functions: 80, lines: 80 },
      reportOnFailure: false,
    },
  },
})
