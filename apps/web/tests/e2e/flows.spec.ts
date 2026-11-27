import { expect, test } from '@playwright/test'
const reservation = '00000000-0000-4000-8000-200000000001'

test('filtra reservas, recupera errores y pagina', async ({ page }) => {
  await page.goto('/#/reservas')
  await expect(page.getByRole('article')).toHaveCount(4)
  await page.getByLabel('Estado').selectOption('cancelled')
  await expect(page.getByRole('article')).toHaveCount(1)
  await page.getByRole('button', { name: 'Limpiar filtros' }).click()
  await page.getByLabel('Lectura').selectOption('paginated')
  await expect(page.getByRole('button', { name: 'Cargar más' })).toBeVisible()
  await page.getByRole('button', { name: 'Cargar más' }).click()
  await page.getByLabel('Lectura').selectOption('error')
  await expect(page.getByRole('alert')).toContainText('Servicio')
  await page.getByLabel('Lectura').selectOption('normal')
  await expect(page.getByRole('article')).toHaveCount(4)
})

test('crea una reserva y reintenta un conflicto', async ({ page }) => {
  await page.goto('/#/reservas/nueva')
  await page.getByLabel('Recurso').selectOption({ label: 'Recurso 1' })
  await page.getByLabel('Fecha').fill('2026-11-10')
  await page.getByRole('button', { name: /\d\d:\d\d .* \d\d:\d\d/ }).first().click()
  await page.getByLabel('Escritura').selectOption('conflict')
  await page.getByRole('button', { name: 'Confirmar reserva' }).click()
  await expect(page.getByRole('alert')).toBeVisible()
  await page.getByLabel('Escritura').selectOption('normal')
  await page.getByRole('button', { name: 'Confirmar reserva' }).click()
  await expect(page.getByText('Reserva confirmada')).toBeVisible()
})

test('consulta y cancela una reserva', async ({ page }) => {
  await page.goto('/#/reservas/' + reservation)
  await page.getByRole('button', { name: 'Cancelar reserva' }).click()
  await page.getByRole('button', { name: 'Confirmar cancelación' }).click()
  await expect(page.getByText('Reserva cancelada')).toBeVisible()
})

test('muestra el listado en pantalla móvil', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 812 })
  await page.goto('/#/reservas')
  await expect(page.getByRole('article')).toHaveCount(4)
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(375)
})
