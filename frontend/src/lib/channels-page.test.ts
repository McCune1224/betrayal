import { render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Page from '../routes/channels/+page.svelte';

afterEach(() => vi.unstubAllGlobals());

describe('channels page', () => {
  it('renders channel summary and statuses', async () => {
    vi.stubGlobal('fetch', vi.fn().mockImplementation((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/discord/resources')) {
        return Promise.resolve(new Response(JSON.stringify({ guilds: [], channels: [], members: [] }), { headers: { 'content-type': 'application/json' } }));
      }
      return Promise.resolve(new Response(JSON.stringify({ discord_connected: false, entries: [{ name: 'Vote Channel', kind: 'vote', status: 'missing' }], summary: { total: 1, missing: 1, configured: 0, orphaned: 0, unverified: 0 } }), { headers: { 'content-type': 'application/json' } }));
    }));
    render(Page);
    expect(await screen.findByRole('heading', { name: 'Channels' })).toBeInTheDocument();
    expect(await screen.findByText('Vote Channel')).toBeInTheDocument();
    expect(screen.getByText('missing')).toBeInTheDocument();
  });

  it('renders configured channels by human-readable name without exposing snowflakes', async () => {
    const channelID = '1140968068705701898';
    vi.stubGlobal('fetch', vi.fn().mockImplementation((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/discord/resources')) {
        return Promise.resolve(new Response(JSON.stringify({
          guilds: [{ id: 'guild-1', name: 'Sandbox' }],
          channels: [{ id: channelID, guild_id: 'guild-1', name: 'vote-funnel', type: '0', category: 'Game' }],
          members: []
        }), { headers: { 'content-type': 'application/json' } }));
      }
      return Promise.resolve(new Response(JSON.stringify({
        discord_connected: true,
        entries: [{ name: 'Vote Channel', kind: 'vote', channel_id: channelID, label: '#vote-funnel', status: 'configured' }],
        summary: { total: 1, missing: 0, configured: 1, orphaned: 0, unverified: 0 }
      }), { headers: { 'content-type': 'application/json' } }));
    }));

    render(Page);

    expect(await screen.findByText('#vote-funnel')).toBeInTheDocument();
    expect(screen.queryByText(channelID)).not.toBeInTheDocument();
  });
});
