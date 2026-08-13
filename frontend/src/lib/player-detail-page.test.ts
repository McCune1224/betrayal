import { cleanup, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';

import Page from '../routes/players/[id]/+page.svelte';

vi.mock('$app/state', () => ({ page: { params: { id: '701' } } }));

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('player detail page', () => {
  it('offers catalog item selection and note creation controls', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn()
        .mockResolvedValueOnce(
          new Response(
            JSON.stringify({
              id: '701', alive: true, coins: 42, luck: 7, item_limit: 3,
              alignment: 'GOOD', role: 'Oracle', items: [], abilities: [], statuses: [],
              immunities: [], perks: [], notes: []
            }),
            { headers: { 'content-type': 'application/json' } }
          )
        )
        .mockResolvedValueOnce(
          new Response(JSON.stringify([{ id: 12, name: 'Rope', description: '', rarity: 'COMMON', cost: 10 }]), {
            headers: { 'content-type': 'application/json' }
          })
        )
    );

    render(Page);

    expect(await screen.findByRole('option', { name: 'Rope' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Add item' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Add note' })).toBeInTheDocument();
    expect(screen.getByLabelText('Note text')).toBeInTheDocument();
  });

  it('shows the guild nickname instead of the Discord username', async () => {
    vi.stubGlobal('fetch', vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({
        id: '701', alive: true, coins: 42, luck: 7, item_limit: 3,
        alignment: 'GOOD', role: 'Oracle', items: [], abilities: [], statuses: [], immunities: [], perks: [], notes: []
      }), { headers: { 'content-type': 'application/json' } }))
      .mockResolvedValueOnce(new Response(JSON.stringify([]), { headers: { 'content-type': 'application/json' } }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ members: [{ id: '701', username: 'oracle_user', nickname: 'The Oracle', bot: false }] }), { headers: { 'content-type': 'application/json' } })));

    render(Page);

    expect(await screen.findByRole('heading', { name: 'The Oracle' })).toBeInTheDocument();
    expect(screen.queryByText('oracle_user')).not.toBeInTheDocument();
  });
});
