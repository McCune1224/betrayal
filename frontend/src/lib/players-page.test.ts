import { render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';

import Page from '../routes/players/+page.svelte';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('players page', () => {
  it('shows a loading state while the player request is pending', () => {
    vi.stubGlobal('fetch', vi.fn(() => new Promise<Response>(() => {})));

    render(Page);

    expect(screen.getByRole('status')).toHaveTextContent('Loading players…');
  });

  it('shows an error state when the player request fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: { message: 'authentication required' } }), {
          status: 401,
          headers: { 'content-type': 'application/json' }
        })
      )
    );

    render(Page);

    expect(await screen.findByRole('alert')).toHaveTextContent('authentication required');
  });

  it('shows an empty state when the API returns no players', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(new Response(JSON.stringify([]), { headers: { 'content-type': 'application/json' } }))
    );

    render(Page);

    expect(await screen.findByText('No players are available.')).toBeInTheDocument();
  });

  it('loads player DTOs from the shared API client endpoint', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify([
          {
            id: 701,
            alive: true,
            coins: 42,
            luck: 7,
            item_limit: 3,
            alignment: 'GOOD',
            role: 'Oracle'
          }
        ]),
        { headers: { 'content-type': 'application/json' } }
      )
    );
    vi.stubGlobal('fetch', fetchMock);

    render(Page);

    expect(await screen.findByText('Oracle')).toBeInTheDocument();
    expect(screen.getByText('701')).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
    expect(screen.getByText('7')).toBeInTheDocument();
    expect(screen.getByText('3')).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/players', {
      credentials: 'include',
      headers: expect.any(Headers),
      method: 'GET'
    });
  });
});
