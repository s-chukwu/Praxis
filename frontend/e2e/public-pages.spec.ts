import { test, expect } from '@playwright/test';

test('TC-028 landing page loads with sign in and sign up links', async ({ page }) => {
  await page.goto('/');
  await expect(page.locator('body')).not.toBeEmpty();
  await expect(page.getByRole('link', { name: /sign in|log in/i }).first()).toBeVisible();
  await expect(page.getByRole('link', { name: /sign up|get started|create/i }).first()).toBeVisible();
});

test('TC-016 login page shows email and password', async ({ page }) => {
  await page.goto('/login');
  await expect(page.getByLabel('Email')).toBeVisible();
  await expect(page.getByLabel('Password')).toBeVisible();
});

test('TC-029 unknown route shows the 404 page', async ({ page }) => {
  const res = await page.goto('/does-not-exist');
  expect(res?.status()).toBe(404);
});

test('TC-030 dashboard content is not shown when signed out', async ({ page }) => {
  await page.goto('/dashboard');
  await page.waitForLoadState('networkidle');
  await expect(page.getByText(/your next step|skill progress/i)).toHaveCount(0);
});

test('TC-003 there is no standalone skill catalog route', async ({ page }) => {
  const res = await page.goto('/skill-catalog');
  expect(res?.status()).toBe(404);
});