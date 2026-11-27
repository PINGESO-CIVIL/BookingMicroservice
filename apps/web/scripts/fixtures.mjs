import { readFileSync, writeFileSync } from 'node:fs'
import { parse } from 'yaml'
const contract = parse(readFileSync(new URL('../../../contracts/openapi.yaml', import.meta.url), 'utf8'))
const examples = Object.fromEntries(Object.entries(contract.components.examples).map(([name, example]) => [name, example.value]))
writeFileSync(new URL('../src/api/examples.json', import.meta.url), JSON.stringify(examples, null, 2) + '\n')
console.log('Ejemplos extraídos del contrato OpenAPI.')
