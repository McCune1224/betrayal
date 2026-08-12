import { render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';

import Page from '../routes/players/[id]/edit/+page.svelte';

vi.mock('$app/state', () => ({ page: { params: { id: '701' } } }));
vi.mock('$app/navigation', () => ({ goto: vi.fn() }));

afterEach(() => vi.unstubAllGlobals());

describe('player edit page', () => {
  it('loads existing player values into editable stats', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
      id: '701', alive: true, coins: 42, luck: 7, item_limit: 3, alignment: 'GOOD', role: 'Oracle'
    }), { headers: { 'content-type': 'application/json' } })));

    render(Page);

    expect(await screen.findByDisplayValue('7')).toBeInTheDocument();
    expect(screen.getByDisplayValue('42')).toBeInTheDocument();
    expect(screen.getByDisplayValue('3')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Oracle')).toBeInTheDocument();
  });
});
