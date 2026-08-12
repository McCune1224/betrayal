import { fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, expect, it, vi } from 'vitest';
import Page from '../routes/roles/+page.svelte';

afterEach(() => vi.unstubAllGlobals());

it('renders role catalog data and CRUD affordances', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify([
    { id: 1, name: 'Oracle', description: 'Sees danger', alignment: 'GOOD', abilities: [], perks: [] }
  ]), { headers: { 'content-type': 'application/json' } })));
  render(Page);
  expect(await screen.findByText('Oracle')).toBeInTheDocument();
  expect(screen.getByRole('button', { name: 'Create role' })).toBeInTheDocument();
});

it('does not delete a catalog record when the confirmation is cancelled', async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify([
    { id: 1, name: 'Oracle', description: 'Sees danger', alignment: 'GOOD', abilities: [], perks: [] }
  ]), { headers: { 'content-type': 'application/json' } }));
  vi.stubGlobal('fetch', fetchMock);
  const confirmMock = vi.fn().mockReturnValue(false);
  vi.stubGlobal('confirm', confirmMock);

  render(Page);
  await screen.findByText('Oracle');
  await fireEvent.click(screen.getByRole('button', { name: 'Delete' }));

  expect(confirmMock).toHaveBeenCalledWith(expect.stringContaining('Delete Oracle'));
  expect(fetchMock).toHaveBeenCalledTimes(1);
});
