import { expect, Page, test } from '@playwright/test';

async function mockSession(page: Page, authenticated: boolean) {
  await page.route('**/api/v1/auth/session', async (route) => {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify({ authenticated })
    });
  });
}

test.describe('admin browser smoke flows', () => {
  test('loads and bootstraps the login shell without failed requests', async ({ page }) => {
    await mockSession(page, false);
    const failedRequests: string[] = [];
    page.on('requestfailed', (request) => failedRequests.push(request.url()));
    await page.goto('/login');

    await expect(page.getByRole('heading', { name: 'Log in' })).toBeVisible();
    expect(failedRequests).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)).toBe(true);
  });

  test('redirects protected frontend routes when the session is absent', async ({ page }) => {
    await mockSession(page, false);
    await page.goto('/admin/reset');

    await expect(page).toHaveURL(/\/login$/);
    await expect(page.getByRole('textbox', { name: 'Password' })).toBeVisible();
  });

  test('requires the reset phrase and acknowledgement before enabling reset', async ({ page }) => {
    await mockSession(page, true);
    await page.route('**/api/v1/admin/reset', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          contentType: 'application/json',
          body: JSON.stringify({ summary: { Players: 0, CycleRows: 1 } })
        });
      }
    });

    await page.goto('/admin/reset');

    const phrase = page.getByRole('textbox', { name: 'RESET BETRAYAL GAME' });
    const acknowledgement = page.getByRole('checkbox', { name: /permanently clears/i });
    const execute = page.getByRole('button', { name: 'Execute reset' });

    await expect(phrase).toBeVisible();
    await expect(acknowledgement).toBeVisible();
    await expect(execute).toBeDisabled();

    await phrase.fill('RESET BETRAYAL GAME');
    await expect(execute).toBeDisabled();
    await acknowledgement.check();
    await expect(execute).toBeEnabled();
  });

  test('does not issue a catalog delete when confirmation is dismissed', async ({ page }) => {
    await mockSession(page, true);
    await page.route('**/api/v1/catalog/roles', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          contentType: 'application/json',
          body: JSON.stringify([{ id: 1, name: 'Oracle', description: 'Sees danger', alignment: 'GOOD', abilities: [], perks: [] }])
        });
      }
    });

    let deleteRequests = 0;
    page.on('request', (request) => {
      if (request.method() === 'DELETE') deleteRequests += 1;
    });
    page.on('dialog', async (dialog) => dialog.dismiss());

    await page.goto('/roles');
    await page.getByRole('button', { name: 'Delete' }).click();

    await expect.poll(() => deleteRequests).toBe(0);
  });
});
