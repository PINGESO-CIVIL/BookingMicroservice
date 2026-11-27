import { copyFile } from 'node:fs/promises'
await copyFile(
  new URL('../src/api/schema.d.ts', import.meta.url),
  new URL('../package-dist/types/api/schema.d.ts', import.meta.url),
)
