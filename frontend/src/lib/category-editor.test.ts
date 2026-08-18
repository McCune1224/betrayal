import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, expect, it, vi } from 'vitest';
import ItemsPage from '../routes/items/+page.svelte';

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const ITEM = { id: 1, name: 'Veilstone', description: 'Smoky shard', rarity: 'RARE', cost: 5, categories: ['Charm'] };
const CATEGORY_OPTIONS = [{ id: 5, name: 'Charm' }, { id: 6, name: 'Portents' }];

function stubFetch() {
  const calls: { url: string; init?: RequestInit }[] = [];
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo, init?: RequestInit) => {
    const url = String(input);
    calls.push({ url, init });
    const method = (init?.method ?? 'GET').toUpperCase();
    const json = (body: unknown) => Promise.resolve(new Response(JSON.stringify(body), { headers: { 'content-type': 'application/json' } }));
    if (url.includes('/auth/csrf')) return json({ token: 't' });
    if (url.includes('/catalog/items/1') && method === 'DELETE') return Promise.resolve(new Response(null, { status: 204 }));
    if (url.includes('/catalog/items/1') && method === 'POST') return json({ ...ITEM, categories: ['Charm', 'Portents'] });
    if (url.includes('/catalog/items/1')) return json(ITEM);
    if (url.includes('/catalog/items')) return json([ITEM]);
    if (url.includes('/catalog/categories')) return json(CATEGORY_OPTIONS);
    throw new Error(`unexpected fetch: ${method} ${url}`);
  }));
  return calls;
}

async function openEditor() {
  const calls = stubFetch();
  render(ItemsPage);
  await screen.findByText('Veilstone');
  await waitFor(() => expect(calls.length).toBeGreaterThanOrEqual(2)); // list + category options
  await fireEvent.click(screen.getByRole('button', { name: 'Edit' }));
  await screen.findByRole('heading', { name: 'Edit Item' });
  return calls;
}

it('renders assigned categories with an add control in item edit mode', async () => {
  await openEditor();
  expect(screen.getByText('Charm')).toBeInTheDocument();
  expect(screen.getByRole('combobox', { name: 'Add category' })).toBeInTheDocument();
  expect(screen.getByRole('button', { name: 'Add category' })).toBeDisabled();
});

it('assigns a category by POST with the exact name payload', async () => {
  const calls = await openEditor();
  await fireEvent.change(screen.getByRole('combobox', { name: 'Add category' }), { target: { value: 'Portents' } });
  await fireEvent.click(screen.getByRole('button', { name: 'Add category' }));

  await waitFor(() => {
    const post = calls.find((c) => (c.init?.method ?? 'GET').toUpperCase() === 'POST' && c.url.includes('/catalog/items/1/categories'));
    expect(post).toBeTruthy();
  });
  const post = calls.find((c) => (c.init?.method ?? 'GET').toUpperCase() === 'POST' && c.url.includes('/catalog/items/1/categories'))!;
  expect(JSON.parse(String(post.init?.body))).toEqual({ category: 'Portents' });
});

it('a cancelled category remove never sends a DELETE', async () => {
  const calls = await openEditor();
  vi.stubGlobal('confirm', vi.fn().mockReturnValue(false));
  await fireEvent.click(screen.getByRole('button', { name: 'Remove' }));

  expect(calls.some((c) => (c.init?.method ?? 'GET').toUpperCase() === 'DELETE')).toBe(false);
});

it('a confirmed category remove resolves the id by name and deletes it', async () => {
  const calls = await openEditor();
  vi.stubGlobal('confirm', vi.fn().mockReturnValue(true));
  await fireEvent.click(screen.getByRole('button', { name: 'Remove' }));

  await waitFor(() =>
    expect(calls.some((c) => (c.init?.method ?? 'GET').toUpperCase() === 'DELETE' && c.url.includes('/catalog/items/1/categories/5'))).toBe(true)
  );
});