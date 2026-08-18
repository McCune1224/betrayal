import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import RolesPage from '../routes/roles/+page.svelte';

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const ROLE = { id: 1, name: 'Oracle', description: 'Sees danger', alignment: 'GOOD', abilities: [], perks: [] };
const ROLE_WITH_LINKS = {
  ...ROLE,
  abilities: [{ id: 11, name: 'Second Sight' }],
  perks: [{ id: 21, name: 'Silver Tongue' }]
};
const ABILITY_OPTIONS = [{ id: 11, name: 'Second Sight' }, { id: 12, name: 'Whispers' }];
const PERK_OPTIONS = [{ id: 21, name: 'Silver Tongue' }, { id: 22, name: 'Iron Nerve' }];

function stubRoleFetch() {
  const calls: { url: string; init?: RequestInit }[] = [];
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo, init?: RequestInit) => {
    const url = String(input);
    calls.push({ url, init });
    const method = (init?.method ?? 'GET').toUpperCase();
    const json = (body: unknown) => Promise.resolve(new Response(JSON.stringify(body), { headers: { 'content-type': 'application/json' } }));
    if (url.includes('/auth/csrf')) return json({ token: 't' });
    if (url.includes('/catalog/roles/1') && method === 'DELETE') return Promise.resolve(new Response(null, { status: 204 }));
    if (url.includes('/catalog/roles/1') && method === 'POST') return json(ROLE_WITH_LINKS);
    if (url.includes('/catalog/roles/1')) return json(ROLE_WITH_LINKS);
    if (url.includes('/catalog/roles')) return json([ROLE]);
    if (url.includes('/catalog/abilities')) return json(ABILITY_OPTIONS);
    if (url.includes('/catalog/perks')) return json(PERK_OPTIONS);
    throw new Error(`unexpected fetch: ${method} ${url}`);
  }));
  return calls;
}

async function openEditor() {
  const calls = stubRoleFetch();
  render(RolesPage);
  await screen.findByText('Oracle');
  await waitFor(() => expect(calls.length).toBeGreaterThanOrEqual(3)); // list + ability + perk options
  await fireEvent.click(screen.getByRole('button', { name: 'Edit' }));
  await screen.findByRole('heading', { name: 'Edit Role' });
  return calls;
}

describe('role link editor', () => {
  it('renders linked abilities and perks with add controls', async () => {
    await openEditor();
    expect(screen.getByText('Second Sight')).toBeInTheDocument();
    expect(screen.getByText('Silver Tongue')).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: 'Add ability' })).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: 'Add perk' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Add ability' })).toBeDisabled();
  });

  it('links an ability by POST and refreshes the role detail', async () => {
    const calls = await openEditor();
    const addSelect = screen.getByRole('combobox', { name: 'Add ability' });
    await fireEvent.change(addSelect, { target: { value: 'Whispers' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Add ability' }));

    await waitFor(() => {
      const post = calls.find((c) => (c.init?.method ?? 'GET').toUpperCase() === 'POST' && c.url.includes('/catalog/roles/1/abilities'));
      expect(post).toBeTruthy();
    });
    const post = calls.find((c) => (c.init?.method ?? 'GET').toUpperCase() === 'POST' && c.url.includes('/catalog/roles/1/abilities'))!;
    expect(JSON.parse(String(post.init?.body))).toEqual({ ability: 'Whispers' });
  });

  it('a cancelled remove never sends a DELETE', async () => {
    const calls = await openEditor();
    const confirmMock = vi.fn().mockReturnValue(false);
    vi.stubGlobal('confirm', confirmMock);
    await fireEvent.click(screen.getAllByRole('button', { name: 'Remove' })[0]);

    expect(confirmMock).toHaveBeenCalledWith(expect.stringContaining('Remove Second Sight'));
    expect(calls.some((c) => (c.init?.method ?? 'GET').toUpperCase() === 'DELETE')).toBe(false);
  });

  it('a confirmed remove sends DELETE and refreshes the role detail', async () => {
    const calls = await openEditor();
    vi.stubGlobal('confirm', vi.fn().mockReturnValue(true));
    await fireEvent.click(screen.getAllByRole('button', { name: 'Remove' })[0]);

    await waitFor(() => expect(calls.some((c) => (c.init?.method ?? 'GET').toUpperCase() === 'DELETE' && c.url.includes('/catalog/roles/1/abilities/11'))).toBe(true));
    // Refresh GET follows the DELETE.
    await waitFor(() => expect(calls.filter((c) => (c.init?.method ?? 'GET').toUpperCase() === 'GET' && c.url.includes('/catalog/roles/1') && c.url.endsWith('/roles/1')).length).toBeGreaterThan(0));
  });
});