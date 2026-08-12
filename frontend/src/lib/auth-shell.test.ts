import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';

const navigation = vi.hoisted(() => ({ goto: vi.fn(), afterNavigate: vi.fn() }));
vi.mock('$app/navigation', () => navigation);

import AuthShell from '../lib/auth/AuthShell.svelte';

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  window.history.replaceState({}, '', '/');
  navigation.afterNavigate.mockReset();
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

    const menuButton = await screen.findByRole('button', { name: /menu/i });
    await fireEvent.click(menuButton);
    expect(menuButton).toHaveAttribute('aria-expanded', 'true');
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

  it('rechecks authentication when navigating away from the login route', async () => {
    window.history.pushState({}, '', '/login');
    let navigate: (() => void) | undefined;
    navigation.afterNavigate.mockImplementation((callback: () => void) => {
      navigate = callback;
    });
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(json({ authenticated: true })));

    render(AuthShell, { children: () => 'Protected content' });
    expect(screen.queryByRole('link', { name: 'Dashboard' })).not.toBeInTheDocument();

    window.history.pushState({}, '', '/');
    navigate?.();

    expect(await screen.findByRole('link', { name: 'Dashboard' })).toBeInTheDocument();
  });
});
