import { render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Page from '../routes/healthcheck/+page.svelte';

afterEach(() => vi.unstubAllGlobals());

describe('healthcheck page', () => {
  it('renders readiness and component counts', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ ready: true, discord_connected: true, channels: { admin: 1, admin_ready: true, vote: 'v', vote_ready: true, action: 'a', action_ready: true, lifeboard_ready: false }, players: { total: 4, alive: 3, dead: 1 }, cycle: { ready: true, day: 2, phase: 'Day' } }), { headers: { 'content-type': 'application/json' } })));
    render(Page);
    expect(await screen.findByRole('heading', { name: 'Healthcheck' })).toBeInTheDocument();
    expect(await screen.findByText('Ready')).toBeInTheDocument();
    expect(await screen.findByText(/3 alive \/ 1 dead/)).toBeInTheDocument();
  });
});
