export type Fetcher = typeof fetch;

export interface ApiClientOptions {
  fetcher?: Fetcher;
  csrfPath?: string;
}

interface ErrorEnvelope {
  error?: {
    code?: string;
    message?: string;
  };
}

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly code?: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export function createApiClient({ fetcher = fetch, csrfPath = '/api/v1/auth/csrf' }: ApiClientOptions = {}) {
  async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const method = (init.method ?? 'GET').toUpperCase();
    const headers = new Headers(init.headers);

    if (!['GET', 'HEAD', 'OPTIONS'].includes(method)) {
      const csrfResponse = await fetcher(csrfPath, {
        credentials: 'include',
        method: 'GET'
      });
      const csrf = (await csrfResponse.json()) as { token: string };
      headers.set('X-CSRF-Token', csrf.token);
    }

    const response = await fetcher(path, {
      ...init,
      credentials: 'include',
      headers,
      method
    });

    const contentType = response.headers.get('content-type') ?? '';
    if (!contentType.includes('application/json')) {
      throw new ApiError('Expected a JSON API response', response.status);
    }

    const body = (await response.json()) as T & ErrorEnvelope;
    if (!response.ok) {
      throw new ApiError(body.error?.message ?? `Request failed (${response.status})`, response.status, body.error?.code);
    }

    return body;
  }

  return {
    get<T>(path: string, init: RequestInit = {}) {
      return request<T>(path, { ...init, method: 'GET' });
    },
    post<T>(path: string, init: RequestInit = {}) {
      return request<T>(path, { ...init, method: 'POST' });
    }
  };
}
