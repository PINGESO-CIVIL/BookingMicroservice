import { readFileSync, writeFileSync, mkdirSync } from 'node:fs'
import { parse } from 'yaml'
import Ajv from 'ajv'
import addFormats from 'ajv-formats'
const spec = parse(readFileSync(new URL('../../contracts/openapi/openapi.yaml', import.meta.url), 'utf8'))
const fixture = JSON.parse(readFileSync(new URL('../../Backend/internal/mockapi/fixtures.json', import.meta.url), 'utf8'))
const base = process.env.MOCK_BASE_URL ?? 'http://127.0.0.1:4010'
const ajv = new Ajv({ strict: false, allErrors: true, validateFormats: true })
addFormats(ajv)
ajv.addFormat('int64', { type: 'number', validate: Number.isInteger })
const resolve = value => value.$ref ? resolve(value.$ref.slice(2).split('/').reduce((current, key) => current[key], spec)) : value
const endpoints = [
  ['get', '/api/v1/reservations', '/api/v1/reservations', null, 200],
  ['post', '/api/v1/reservations', '/api/v1/reservations', 'CreateRequest', 201],
  ['get', '/api/v1/reservations/{id}', '/api/v1/reservations/reservation-001', null, 200],
  ['patch', '/api/v1/reservations/{id}', '/api/v1/reservations/reservation-001', 'UpdateRequest', 200],
  ['post', '/api/v1/reservations/{id}/cancel', '/api/v1/reservations/reservation-001/cancel', { expectedVersion: 1 }, 200],
  ['get', '/api/v1/resources', '/api/v1/resources', null, 200],
  ['post', '/api/v1/resources', '/api/v1/resources', 'CreateResourceRequest', 201],
  ['get', '/api/v1/resources/{id}', '/api/v1/resources/resource-001', null, 200],
  ['patch', '/api/v1/resources/{id}', '/api/v1/resources/resource-001', 'UpdateResourceRequest', 200],
  ['put', '/api/v1/resources/{id}/schedule', '/api/v1/resources/resource-001/schedule', 'UpdateResourceScheduleRequest', 200],
  ['post', '/api/v1/resources/{id}/closures', '/api/v1/resources/resource-001/closures', 'CreateResourceClosureRequest', 201],
  ['delete', '/api/v1/resources/{id}/closures/{closureId}', '/api/v1/resources/resource-001/closures/closure-001', null, 204],
  ['get', '/api/v1/resources/{id}/availability', '/api/v1/resources/resource-001/availability?from=2026-11-10T03:00:00Z&to=2026-11-11T03:00:00Z', null, 200],
  ['get', '/api/v1/admin/reservations', '/api/v1/admin/reservations', null, 200],
]
const cases = [...endpoints,
  ...['empty', 'paginated', 'error', 'unauthorized', 'forbidden'].map(scenario => [...endpoints[0].slice(0, 4), { error: 503, unauthorized: 401, forbidden: 403 }[scenario] ?? 200, scenario]),
  ...['cancelled', 'expired', 'started', 'notFound'].map(scenario => [...endpoints[2].slice(0, 4), scenario === 'notFound' ? 404 : 200, scenario]),
  ...['empty', 'disabled'].map(scenario => [...endpoints[12].slice(0, 4), 200, scenario]),
  [...endpoints[3].slice(0, 4), 409, 'version'], [...endpoints[8].slice(0, 4), 409, 'conflict'],
  [...endpoints[13].slice(0, 4), 403, 'forbidden'],
]
const checks = []
for (const [method, template, path, example, expected, scenario = 'normal'] of cases) {
  const body = typeof example === 'string' ? fixture.examples[example] : example
  const headers = { Authorization: 'Bearer demo-admin', 'X-Mock-Scenario': scenario, 'Idempotency-Key': 'contract-check' }
  if (body) headers['Content-Type'] = 'application/json'
  const response = await fetch(`${base}${path}`, { method: method.toUpperCase(), headers, body: body ? JSON.stringify(body) : undefined, signal: AbortSignal.timeout(10000) })
  if (response.status !== expected) throw new Error(`${method} ${path}: esperaba ${expected}, recibió ${response.status}`)
  if (response.headers.get('X-Mock-Response') !== 'true') throw new Error('Falta el indicador de respuesta mock.')
  const operation = spec.paths[template][method]
  const responseDefinition = operation.responses[String(expected)]
  if (!responseDefinition) throw new Error(`Respuesta ${expected} no declarada para ${operation.operationId}.`)
  const schema = resolve(responseDefinition).content?.['application/json']?.schema
  if (schema) {
    const document = JSON.parse(JSON.stringify({ ...schema, definitions: spec.components.schemas }).replaceAll('#/components/schemas/', '#/definitions/'))
    const validate = ajv.compile(document), data = await response.json()
    if (!validate(data)) throw new Error(`${operation.operationId} (${scenario}): ${ajv.errorsText(validate.errors)}`)
  } else if ((await response.text()) !== '') throw new Error('La respuesta sin contenido contiene un cuerpo.')
  checks.push({ operation: operation.operationId, method: method.toUpperCase(), path, scenario, status: response.status, schemaValid: true })
}
const destination = new URL('../../reports/demo/', import.meta.url)
mkdirSync(destination, { recursive: true })
writeFileSync(new URL('http-contracts.json', destination), JSON.stringify({ checkedAt: new Date().toISOString(), mode: 'mock', contractVersion: spec.info.version, operations: endpoints.length, checks }, null, 2) + '\n')
console.log(`${endpoints.length} operaciones y ${checks.length} respuestas HTTP verificadas contra OpenAPI. Evidencia: reports/demo/http-contracts.json`)
