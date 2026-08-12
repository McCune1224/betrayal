import { render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Page from '../routes/whispers/+page.svelte';

afterEach(() => vi.unstubAllGlobals());

describe('whispers page', () => {
  it('renders human-readable player labels instead of snowflake IDs', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
      groups: [{ id: 1, name: 'Birth links', players: ['206268866714970630'] }],
      players: [{ id: '206268866714970630', label: 'Alex', detail: 'Discord member' }],
      messages: []
    }), { headers: { 'content-type': 'application/json' } })));

    render(Page);

    expect(await screen.findByText('Alex')).toBeInTheDocument();
    expect(screen.queryByText('206268866714970630')).not.toBeInTheDocument();
  });

  it('leaves loading and shows an actionable error when settings fail', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('network unavailable')));

    render(Page);

    expect(await screen.findByRole('alert')).toHaveTextContent('network unavailable');
    expect(screen.queryByText('Loading whisper settings…')).not.toBeInTheDocument();
  });
});
