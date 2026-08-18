import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';

import Page from '../routes/players/[id]/+page.svelte';

vi.mock('$app/state', () => ({ page: { params: { id: '701' } } }));
const gotoMock = vi.hoisted(() => vi.fn());
vi.mock('$app/navigation', () => ({ goto: gotoMock }));

const PLAIN_PLAYER = {
  id: '701', alive: true, coins: 42, luck: 7, item_limit: 3,
  alignment: 'GOOD', role: 'Oracle', items: [], abilities: [], statuses: [],
  immunities: [], perks: [], notes: []
};

function stubDetailFetch(deleteResponse?: () => Response) {
  const calls: { url: string; init?: RequestInit }[] = [];
  vi.stubGlobal(
    'fetch',
    vi.fn((input: RequestInfo, init?: RequestInit) => {
      const url = String(input);
      calls.push({ url, init });
      if (url.includes('/auth/csrf')) return Promise.resolve(new Response(JSON.stringify({ token: 'test-token' }), { headers: { 'content-type': 'application/json' } }));
      if (url.includes('/players/701') && (init?.method ?? 'GET') !== 'DELETE') {
        return Promise.resolve(new Response(JSON.stringify(PLAIN_PLAYER), { headers: { 'content-type': 'application/json' } }));
      }
      if (url.includes('/catalog/items')) return Promise.resolve(new Response(JSON.stringify([]), { headers: { 'content-type': 'application/json' } }));
      if (url.includes('/discord/resources')) return Promise.resolve(new Response(JSON.stringify({ members: [{ id: '701', username: 'oracle_user', nickname: 'The Oracle', bot: false }] }), { headers: { 'content-type': 'application/json' } }));
      if (deleteResponse) return Promise.resolve(deleteResponse());
      throw new Error(`unexpected fetch: ${url}`);
    })
  );
  return { calls, fetchMock: vi.mocked(fetch) };
}

function deleteCalls(calls: { url: string; init?: RequestInit }[]) {
  return calls.filter((call) => (call.init?.method ?? 'GET').toUpperCase() === 'DELETE' && call.url.includes('/players/701'));
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  gotoMock.mockReset();
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

  it('requires typing the exact player label to enable removal', async () => {
    stubDetailFetch();
    render(Page);

    const confirm = await screen.findByLabelText('Type player label to confirm');
    const remove = screen.getByRole('button', { name: 'Remove player' }) as HTMLButtonElement;
    expect(remove.disabled).toBe(true);

    await fireEvent.input(confirm, { target: { value: 'The Or' } });
    expect(remove.disabled).toBe(true);
    await fireEvent.input(confirm, { target: { value: 'The Oracle' } });
    expect(remove.disabled).toBe(false);
  });

  it('a mismatched confirmation phrase never sends a delete request', async () => {
    const { calls } = stubDetailFetch();
    render(Page);

    const confirm = await screen.findByLabelText('Type player label to confirm');
    const remove = screen.getByRole('button', { name: 'Remove player' }) as HTMLButtonElement;
    await fireEvent.input(confirm, { target: { value: 'Not the label' } });
    expect(remove.disabled).toBe(true);
    await fireEvent.click(remove);
    expect(deleteCalls(calls)).toHaveLength(0);
    expect(gotoMock).not.toHaveBeenCalled();
  });

  it('removes the player after a confirmed delete and returns to the roster', async () => {
    const { calls } = stubDetailFetch(() => new Response(null, { status: 204 }));
    render(Page);

    const confirm = await screen.findByLabelText('Type player label to confirm');
    await fireEvent.input(confirm, { target: { value: 'The Oracle' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Remove player' }));

    await vi.waitFor(() => expect(deleteCalls(calls)).toHaveLength(1));
    expect(gotoMock).toHaveBeenCalledWith('/players');
  });

  it('surfaces the error when the delete request fails', async () => {
    const { calls } = stubDetailFetch(() => new Response(JSON.stringify({ error: { code: 'player_delete_failed', message: 'could not delete player' } }), {
      status: 500,
      headers: { 'content-type': 'application/json' }
    }));
    render(Page);

    const confirm = await screen.findByLabelText('Type player label to confirm');
    await fireEvent.input(confirm, { target: { value: 'The Oracle' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Remove player' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('could not delete player');
    expect(gotoMock).not.toHaveBeenCalled();
    expect(deleteCalls(calls)).toHaveLength(1);
  });
});
