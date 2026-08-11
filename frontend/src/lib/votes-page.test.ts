import { render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Page from '../routes/votes/+page.svelte';

afterEach(() => vi.unstubAllGlobals());

describe('votes page', () => {
  it('renders votes and tallies from the selected cycle', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ cycle: { day: 1, phase: 'Day', is_elimination: false, is_current: true }, votes: [{ id: 7, voter_id: 10, target_id: 11, weight: 2 }], tallies: [{ target_id: 11, total_votes: 2, vote_count: 1 }], cycles: [], stats: { most_voted_players: [], most_active_voters: [], least_voted_players: [] }, total_votes: 1 }), { headers: { 'content-type': 'application/json' } })));
    render(Page);
    expect(await screen.findByRole('heading', { name: 'Votes' })).toBeInTheDocument();
    expect(await screen.findByText('Target 11')).toBeInTheDocument();
    expect(screen.getByText('2 votes')).toBeInTheDocument();
  });
});
