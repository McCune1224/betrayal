import { fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('$app/navigation', () => ({ goto: vi.fn() }));

import AuthShell from '../lib/auth/AuthShell.svelte';

afterEach(() => {
  vi.unstubAllGlobals();
});

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'content-type': 'application/json' }
  });
}

describe('auth shell', () => {
  it('shows the authenticated navigation and returns to login after logout', async () => {
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(json({ authenticated: true }))
      .mockResolvedValueOnce(json({ token: 'csrf-token' }))
      .mockResolvedValueOnce(json({ authenticated: false }));
    vi.stubGlobal('fetch', fetcher);

    render(AuthShell, { children: () => 'Protected content' });

    const logoutButtons = await screen.findAllByRole('button', { name: /log out/i });
    expect(logoutButtons).toHaveLength(2);
    for (const label of ['Dashboard', 'Players', 'Cycle', 'Channels', 'Votes', 'Setup', 'Healthcheck', 'Roles', 'Items', 'Abilities', 'Statuses', 'Sync', 'Audit log', 'Migrations', 'Reset game', 'Redeploy']) {
      expect(screen.getByRole('link', { name: label })).toBeInTheDocument();
    }
    await fireEvent.click(logoutButtons[0]);

    expect(await screen.findByRole('status')).toHaveTextContent(/signed out/i);
    expect(screen.queryByRole('button', { name: /log out/i })).not.toBeInTheDocument();
    expect(fetcher).toHaveBeenNthCalledWith(
      3,
      '/api/v1/auth/logout',
      expect.objectContaining({ method: 'POST' })
    );
  });
});
