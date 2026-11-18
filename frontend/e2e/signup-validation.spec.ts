import { test, expect } from '@playwright/test';

const WEAK = 'Use at least 8 characters, including a letter and a number.';

test.beforeEach(async ({ page }) => {
  await page.goto('/signup');
});

test('TC-017 signup form shows the three fields', async ({ page }) => {
  await expect(page.getByLabel('Full name')).toBeVisible();
  await expect(page.getByLabel('Email')).toBeVisible();
  await expect(page.getByLabel('Password')).toBeVisible();
});

test('TC-018 empty submit shows all three messages', async ({ page }) => {
  await page.locator('form button[type="submit"]').click();
  await expect(page.getByText('Enter your full name.')).toBeVisible();
  await expect(page.getByText('Enter your email address.')).toBeVisible();
  await expect(page.getByText('Choose a password.')).toBeVisible();
});

test('TC-019 invalid email is rejected', async ({ page }) => {
  await page.getByLabel('Full name').fill('Test User');
  await page.getByLabel('Email').fill('abc@');
  await page.getByLabel('Password').fill('Abcdef12');
  await page.locator('form button[type="submit"]').click();
  await expect(page.getByText('Enter a valid email address.')).toBeVisible();
});

for (const [id, pw] of [
  ['TC-020 letters only', 'abcdefgh'],
  ['TC-021 too short', 'abc12'],
  ['numbers only', '12345678'],
]) {
  test(`${id}: weak password is rejected (BUG-008)`, async ({ page }) => {
    await page.getByLabel('Full name').fill('Test User');
    await page.getByLabel('Email').fill('test.user@example.com');
    await page.getByLabel('Password').fill(pw);
    await page.locator('form button[type="submit"]').click();
    await expect(page.getByText(WEAK)).toBeVisible();
  });
}

test('TC-022 digits are stripped from the name field', async ({ page }) => {
  const name = page.getByLabel('Full name');
  await name.fill('John123');
  await expect(name).toHaveValue('John');
});