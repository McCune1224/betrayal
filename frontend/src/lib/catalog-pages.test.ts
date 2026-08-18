import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, expect, it, vi } from 'vitest';
import Page from '../routes/roles/+page.svelte';
import PerksPage from '../routes/perks/+page.svelte';
import CategoriesPage from '../routes/categories/+page.svelte';

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

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
  const deletes = fetchMock.mock.calls.filter(([, init]) => (init?.method ?? 'GET').toUpperCase() === 'DELETE');
  expect(deletes).toHaveLength(0);
});

it('renders the perks library with create affordances', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify([
    { id: 1, name: 'Silvertongue', description: 'Words bend in your favor.' }
  ]), { headers: { 'content-type': 'application/json' } })));
  render(PerksPage);
  expect(await screen.findByText('Silvertongue')).toBeInTheDocument();
  expect(screen.getByRole('button', { name: 'Create perk' })).toBeInTheDocument();
});

it('renders the category library and hides the description field', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify([
    { id: 1, name: 'Poisons' }
  ]), { headers: { 'content-type': 'application/json' } })));
  render(CategoriesPage);
  expect(await screen.findByText('Poisons')).toBeInTheDocument();
  expect(screen.getByRole('button', { name: 'Create category' })).toBeInTheDocument();
  expect(screen.queryByRole('textbox', { name: 'Description' })).not.toBeInTheDocument();
});
