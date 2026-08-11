import { render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Page from '../routes/channels/+page.svelte';

afterEach(() => vi.unstubAllGlobals());

describe('channels page', () => {
  it('renders channel summary and statuses', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ discord_connected: false, entries: [{ name: 'Vote Channel', kind: 'vote', status: 'missing' }], summary: { total: 1, missing: 1, configured: 0, orphaned: 0, unverified: 0 } }), { headers: { 'content-type': 'application/json' } })));
    render(Page);
    expect(await screen.findByRole('heading', { name: 'Channels' })).toBeInTheDocument();
    expect(await screen.findByText('Vote Channel')).toBeInTheDocument();
    expect(screen.getByText('missing')).toBeInTheDocument();
  });
});
