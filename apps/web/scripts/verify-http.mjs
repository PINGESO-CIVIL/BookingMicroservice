import { readFileSync, writeFileSync, mkdirSync } from 'node:fs'
import { strict as assert } from 'node:assert'
import { parse } from 'yaml'
import Ajv2020 from 'ajv/dist/2020.js'
import addFormats from 'ajv-formats'
const spec = parse(readFileSync(new URL('../../../contracts/openapi.yaml', import.meta.url), 'utf8'))
const fixture = JSON.parse(readFileSync(new URL('../../api/internal/store/fixtures.json', import.meta.url), 'utf8'))
const base = process.env.MOCK_BASE_URL ?? 'http://127.0.0.1:5191'
const ajv = new Ajv2020({ strict: false, allErrors: true, validateFormats: true })
addFormats(ajv)
ajv.addFormat('int64', { type: 'number', validate: Number.isInteger })
const resolve = value => value.$ref ? resolve(value.$ref.slice(2).split('/').reduce((current, key) => current[key], spec)) : value
const endpoints = [
  ['get', '/api/v1/reservas', '/api/v1/reservas', null, 200],
  ['post', '/api/v1/reservas', '/api/v1/reservas', 'CreateRequest', 201],
  ['get', '/api/v1/reservas/{id}', '/api/v1/reservas/00000000-0000-4000-8000-200000000001', null, 200],
  ['patch', '/api/v1/reservas/{id}', '/api/v1/reservas/00000000-0000-4000-8000-200000000001', 'UpdateRequest', 200],
  ['post', '/api/v1/reservas/{id}/cancelaciones', '/api/v1/reservas/00000000-0000-4000-8000-200000000001/cancelaciones', 'CancelRequest', 200],
  ['get', '/api/v1/recursos', '/api/v1/recursos', null, 200],
  ['post', '/api/v1/recursos', '/api/v1/recursos', 'CreateResourceRequest', 201],
  ['get', '/api/v1/recursos/{id}', '/api/v1/recursos/00000000-0000-4000-8000-100000000001', null, 200],
  ['patch', '/api/v1/recursos/{id}', '/api/v1/recursos/00000000-0000-4000-8000-100000000001', 'UpdateResourceRequest', 200],
  ['put', '/api/v1/recursos/{id}/horarios', '/api/v1/recursos/00000000-0000-4000-8000-100000000001/horarios', 'UpdateResourceScheduleRequest', 200],
  ['post', '/api/v1/recursos/{id}/bloqueos', '/api/v1/recursos/00000000-0000-4000-8000-100000000001/bloqueos', 'CreateResourceClosureRequest', 201],
  ['delete', '/api/v1/recursos/{id}/bloqueos/{closureId}', '/api/v1/recursos/00000000-0000-4000-8000-100000000001/bloqueos/00000000-0000-4000-8000-300000000001', null, 204],
  ['get', '/api/v1/recursos/{id}/disponibilidades', '/api/v1/recursos/00000000-0000-4000-8000-100000000001/disponibilidades?from=2026-11-10T03:00:00Z&to=2026-11-11T03:00:00Z', null, 200],
  ['get', '/api/v1/admin/reservas', '/api/v1/admin/reservas', null, 200],
]
const cases = [...endpoints,
  ...['empty', 'paginated', 'error', 'unauthorized', 'forbidden'].map(scenario => [...endpoints[0].slice(0, 4), { error: 503, unauthorized: 401, forbidden: 403 }[scenario] ?? 200, scenario]),
  ...['cancelled', 'expired', 'started', 'notFound'].map(scenario => [...endpoints[2].slice(0, 4), scenario === 'notFound' ? 404 : 200, scenario]),
  ...['empty', 'disabled'].map(scenario => [...endpoints[12].slice(0, 4), 200, scenario]),
  [...endpoints[3].slice(0, 4), 409, 'version'], [...endpoints[8].slice(0, 4), 409, 'conflict'],
  [...endpoints[13].slice(0, 4), 403, 'forbidden'],
]
const checks = []
const documented = Object.entries(spec.paths).flatMap(([path, operations]) =>
  Object.entries(operations)
    .filter(([method]) => ['get', 'post', 'patch', 'put', 'delete', 'head', 'options', 'trace'].includes(method))
    .map(([method, operation]) => {
      assert.ok(operation['x-permission'], `Falta el permiso para ${method} ${path}`)
      return `${method} ${path}`
    }),
)
assert.deepEqual(endpoints.map(([method, path]) => `${method} ${path}`).sort(), documented.sort(),
  'Cada método y ruta del contrato debe tener exactamente un caso principal.')

function request(method, path, scenario = 'normal', body) {
  const headers = { Authorization: 'Bearer demo-admin', 'X-Mock-Scenario': scenario, 'Idempotency-Key': 'contract-check' }
  if (body) headers['Content-Type'] = 'application/json'
  return fetch(`${base}${path}`, { method: method.toUpperCase(), headers, body: body ? JSON.stringify(body) : undefined, signal: AbortSignal.timeout(10000) })
}

function readSnapshot() {
  return Promise.all(endpoints.filter(([method]) => method === 'get').map(async ([method, , path]) => {
    const response = await request(method, path)
    assert.equal(response.status, 200, `No se pudo comprobar la consulta ${path}`)
    return response.json()
  }))
}

const beforeWrites = await readSnapshot()
const validateSchema = schema => {
  const document = JSON.parse(JSON.stringify({ ...schema, $defs: spec.components.schemas }).replaceAll('#/components/schemas/', '#/$defs/'))
  return ajv.compile(document)
}
for (const [method, template, path, example, expected, scenario = 'normal'] of cases) {
  const body = typeof example === 'string' ? fixture.examples[example] : example
  const operation = spec.paths[template][method]
  const requestSchema = operation.requestBody?.content?.['application/json']?.schema
  if (requestSchema) {
    const validate = validateSchema(requestSchema)
    if (!validate(body)) throw new Error(`${operation.operationId} (solicitud): ${ajv.errorsText(validate.errors)}`)
  }
  const response = await request(method, path, scenario, body)
  if (response.status !== expected) throw new Error(`${method} ${path}: esperaba ${expected}, recibió ${response.status}`)
  if (response.headers.get('X-Mock-Response') !== 'true') throw new Error('Falta el indicador de respuesta mock.')
  const responseDefinition = operation.responses[String(expected)]
  if (!responseDefinition) throw new Error(`Respuesta ${expected} no declarada para ${operation.operationId}.`)
  const schema = resolve(responseDefinition).content?.['application/json']?.schema
  if (schema) {
    const validate = validateSchema(schema), data = await response.json()
    if (!validate(data)) throw new Error(`${operation.operationId} (${scenario}): ${ajv.errorsText(validate.errors)}`)
  } else if ((await response.text()) !== '') throw new Error('La respuesta sin contenido contiene un cuerpo.')
  checks.push({ operation: operation.operationId, method: method.toUpperCase(), path, scenario, status: response.status, schemaValid: true })
}
assert.deepEqual(await readSnapshot(), beforeWrites, 'Las escrituras simuladas cambiaron una consulta posterior.')
const destination = new URL('../../../reports/demo/', import.meta.url)
mkdirSync(destination, { recursive: true })
writeFileSync(new URL('http-contracts.json', destination), JSON.stringify({ checkedAt: new Date().toISOString(), mode: 'mock', contractVersion: spec.info.version, operations: endpoints.length, writesLeaveReadsUnchanged: true, checks }, null, 2) + '\n')
console.log(`${endpoints.length} operaciones y ${checks.length} respuestas HTTP verificadas contra OpenAPI. Evidencia: reports/demo/http-contracts.json`)
