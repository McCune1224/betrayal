import { fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';

vi.mock('$app/navigation', () => ({ goto: vi.fn() }));

import Page from '../routes/login/+page.svelte';

afterEach(() => {
  vi.unstubAllGlobals();
});

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'content-type': 'application/json' }
  });
}

describe('login page', () => {
  it('shows the unauthenticated login form', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(json({ authenticated: false })));

    render(Page);

    expect(await screen.findByRole('heading', { name: /log in/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument();
  });

  it('submits a successful JSON login and navigates to the dashboard', async () => {
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(json({ authenticated: false }))
      .mockResolvedValueOnce(json({ token: 'csrf-token' }))
      .mockResolvedValueOnce(json({ authenticated: true }));
    vi.stubGlobal('fetch', fetcher);

    render(Page);
    await screen.findByRole('heading', { name: /log in/i });
    await fireEvent.input(screen.getByLabelText(/password/i), { target: { value: 'secret' } });
    await fireEvent.click(screen.getByRole('button', { name: /log in/i }));

    expect(await screen.findByRole('status')).toHaveTextContent(/logged in/i);
    expect(fetcher).toHaveBeenNthCalledWith(
      3,
      '/api/v1/auth/login',
      expect.objectContaining({ method: 'POST', body: JSON.stringify({ password: 'secret' }) })
    );
  });

  it('shows the JSON error when login fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValueOnce(json({ authenticated: false }))
        .mockResolvedValueOnce(json({ token: 'csrf-token' }))
        .mockResolvedValueOnce(json({ error: { code: 'invalid_credentials', message: 'Invalid password' } }, 401))
    );

    render(Page);
    await screen.findByRole('heading', { name: /log in/i });
    await fireEvent.input(screen.getByLabelText(/password/i), { target: { value: 'wrong' } });
    await fireEvent.click(screen.getByRole('button', { name: /log in/i }));

    expect(await screen.findByRole('alert')).toHaveTextContent('Invalid password');
  });
});
