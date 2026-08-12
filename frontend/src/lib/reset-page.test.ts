import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, expect, it, vi } from 'vitest';
import Page from '../routes/admin/reset/+page.svelte';

afterEach(() => vi.unstubAllGlobals());

it('requires explicit reset acknowledgement before executing', async () => {
  const fetchMock = vi.fn()
    .mockResolvedValueOnce(new Response(JSON.stringify({ summary: { players: 0 } }), { headers: { 'content-type': 'application/json' } }))
    .mockResolvedValueOnce(new Response(JSON.stringify({ token: 'csrf-token' }), { headers: { 'content-type': 'application/json' } }))
    .mockResolvedValueOnce(new Response(JSON.stringify({ status: 'reset' }), { headers: { 'content-type': 'application/json' } }));
  vi.stubGlobal('fetch', fetchMock);

  render(Page);
  const execute = await screen.findByRole('button', { name: 'Execute reset' });
  expect(execute).toBeDisabled();

  await fireEvent.input(screen.getByLabelText(/Type RESET BETRAYAL GAME/i), { target: { value: 'RESET BETRAYAL GAME' } });
  expect(execute).toBeDisabled();

  const acknowledgement = screen.getByRole('checkbox', { name: /understand/i });
  await fireEvent.click(acknowledgement);
  expect(execute).toBeEnabled();
  await fireEvent.click(execute);

  await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(3));
  const request = fetchMock.mock.calls.at(-1)?.[1] as RequestInit;
  expect(JSON.parse(String(request.body))).toEqual({ confirm: 'RESET BETRAYAL GAME', understand: true });
});
