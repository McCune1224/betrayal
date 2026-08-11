import { render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';

import Page from '../routes/+page.svelte';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('dashboard page', () => {
  it('renders cycle and player metrics returned by the dashboard API', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            cycle: { phase: 'Day', number: 3 },
            players: { alive: 8, dead: 2, total: 10 }
          }),
          { headers: { 'content-type': 'application/json' } }
        )
      )
    );

    render(Page);

    expect(await screen.findByText('Day 3')).toBeInTheDocument();
    expect(screen.getByText('8 alive')).toBeInTheDocument();
    expect(screen.getByText('2 dead')).toBeInTheDocument();
    expect(screen.getByText('10 players')).toBeInTheDocument();
  });
});
