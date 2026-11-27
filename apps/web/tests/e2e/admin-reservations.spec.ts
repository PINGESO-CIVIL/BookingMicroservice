import { expect, test } from '@playwright/test'
test('admin accede a reservas y administra horarios del recurso', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('button', { name: 'Admin.' }).click()
  await page.getByRole('link', { name: 'Reservas realizadas' }).click()
  await expect(page.getByRole('article')).toHaveCount(6)
  await expect(page.getByRole('button', { name: 'Nueva reserva' })).toHaveCount(0)
  await page.getByRole('link', { name: 'Recursos' }).click()
  await page.getByRole('button', { name: 'Administrar recurso' }).first().click()
  await page.getByRole('tab', { name: 'Horarios' }).click()
  await expect(page.getByText('Horario semanal')).toBeVisible()
  await page.getByRole('tab', { name: 'Bloqueos' }).click()
  await expect(page.getByRole('button', { name: 'Agregar bloqueo' })).toBeVisible()
})
