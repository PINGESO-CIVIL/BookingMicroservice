import { readFileSync, writeFileSync, mkdirSync, mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, dirname, basename } from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'
import ts from 'typescript'
const root = new URL('../', import.meta.url)
const examples = JSON.parse(readFileSync(new URL('src/api/examples.json', root), 'utf8'))
const temporary = mkdtempSync(join(tmpdir(), 'bookingms-fixtures-'))
try {
  for (const [input, output] of [['src/dates.ts', 'dates.mjs'], ['src/api/availability.ts', 'availability.mjs'], ['src/api/fixtures.ts', 'fixtures.mjs']]) {
    let source = readFileSync(new URL(input, root), 'utf8')
    source = source.replace("import examples from './examples.json'", `const examples = ${JSON.stringify(examples)}`)
    const compiled = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText.replace("'../dates'", "'./dates.mjs'")
    writeFileSync(join(temporary, output), compiled)
  }
  const { resources, reservations, adminReservations, resourceDetail } = await import(pathToFileURL(join(temporary, 'fixtures.mjs')).href)
  const { configuredAvailability } = await import(pathToFileURL(join(temporary, 'availability.mjs')).href)
  const details = Object.fromEntries(resources.map(r => [r.id, resourceDetail(r.id)]))
  // Fixed scenario data is prepared here; the HTTP server does not calculate schedules.
  const availability = Object.fromEntries(resources.map(r => [r.id, configuredAvailability(details[r.id], adminReservations, '2026-11-04T03:00:00Z', '2027-02-01T03:00:00Z', '2026-11-04T16:00:00Z')]))
  const destination = new URL('../../Backend/internal/mockapi/fixtures.json', import.meta.url)
  mkdirSync(dirname(fileURLToPath(destination)), { recursive: true })
  writeFileSync(destination, JSON.stringify({ examples, resources, reservations, adminReservations, details, availability }, null, 2) + '\n')
  console.log('Fixtures HTTP generados desde OpenAPI y escenarios de demostración (noviembre–enero).')
} finally {
  if (dirname(temporary) === tmpdir() && basename(temporary).startsWith('bookingms-fixtures-')) rmSync(temporary, { recursive: true })
}
