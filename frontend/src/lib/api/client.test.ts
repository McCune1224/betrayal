import { describe, expect, it, vi } from 'vitest';

import { createApiClient } from './client';

describe('API client', () => {
  it('includes credentials on safe requests', async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ online: true }), {
        headers: { 'content-type': 'application/json' }
      })
    );
    const client = createApiClient({ fetcher });

    await client.get<{ online: boolean }>('/api/v1/status');

    expect(fetcher).toHaveBeenCalledWith(
      '/api/v1/status',
      expect.objectContaining({ credentials: 'include', method: 'GET' })
    );
  });

  it('obtains and sends a CSRF token for unsafe requests', async () => {
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ token: 'csrf-123' }), {
          headers: { 'content-type': 'application/json' }
        })
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ saved: true }), {
          headers: { 'content-type': 'application/json' }
        })
      );
    const client = createApiClient({ fetcher });

    await client.post<{ saved: boolean }>('/api/v1/items', { body: JSON.stringify({ name: 'rope' }) });

    expect(fetcher).toHaveBeenNthCalledWith(
      1,
      '/api/v1/auth/csrf',
      expect.objectContaining({ credentials: 'include', method: 'GET' })
    );
    expect(fetcher).toHaveBeenNthCalledWith(
      2,
      '/api/v1/items',
      expect.objectContaining({ credentials: 'include', method: 'POST' })
    );
    const requestInit = fetcher.mock.calls[1]?.[1] as RequestInit;
    expect(new Headers(requestInit.headers).get('X-CSRF-Token')).toBe('csrf-123');
  });

  it('uses a custom CSRF path when configured', async () => {
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ token: 'csrf-123' }), {
          headers: { 'content-type': 'application/json' }
        })
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ saved: true }), {
          headers: { 'content-type': 'application/json' }
        })
      );
    const client = createApiClient({ fetcher, csrfPath: '/csrf-override' });

    await client.post<{ saved: boolean }>('/api/v1/items');

    expect(fetcher).toHaveBeenNthCalledWith(1, '/csrf-override', expect.anything());
  });

  it('parses JSON error envelopes', async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ error: { code: 'NOT_FOUND', message: 'Item not found' } }), {
        status: 404,
        headers: { 'content-type': 'application/json' }
      })
    );
    const client = createApiClient({ fetcher });

    await expect(client.get('/api/v1/items/missing')).rejects.toMatchObject({
      code: 'NOT_FOUND',
      message: 'Item not found',
      status: 404
    });
  });

  it('rejects HTML returned with a successful HTTP status', async () => {
    const fetcher = vi.fn().mockResolvedValue(
      new Response('<!doctype html><title>Login</title>', {
        status: 200,
        headers: { 'content-type': 'text/html' }
      })
    );
    const client = createApiClient({ fetcher });

    await expect(client.get('/api/v1/status')).rejects.toMatchObject({
      message: 'Expected a JSON API response',
      status: 200
    });
  });
});
