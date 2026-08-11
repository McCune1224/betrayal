import { render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Page from '../routes/cycle/+page.svelte';

afterEach(() => vi.unstubAllGlobals());

describe('cycle page', () => {
  it('renders cycle DTO state', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ day: 2, phase: 'Elimination', is_elimination: true }), { headers: { 'content-type': 'application/json' } })));
    render(Page);
    expect(await screen.findByRole('heading', { name: 'Cycle' })).toBeInTheDocument();
    expect(await screen.findByText('Elimination 2')).toBeInTheDocument();
  });
});
