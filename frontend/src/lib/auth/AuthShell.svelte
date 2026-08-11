<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { createApiClient, type ApiError } from '$lib/api/client';

  let { children } = $props();
  let authState = $state<'loading' | 'authenticated' | 'signed-out' | 'login'>('loading');
  let error = $state<string | null>(null);

  const onLoginRoute = () => typeof window !== 'undefined' && window.location.pathname === '/login';

  onMount(async () => {
    if (onLoginRoute()) {
      authState = 'login';
      return;
    }

    try {
      const session = await createApiClient().get<{ authenticated: boolean }>('/api/v1/auth/session');
      if (session.authenticated) {
        authState = 'authenticated';
      } else {
        authState = 'login';
        await goto('/login');
      }
    } catch (cause) {
      error = cause instanceof Error ? cause.message : 'Could not load session';
      authState = 'login';
      await goto('/login');
    }
  });

  async function logout() {
    error = null;
    try {
      await createApiClient().post<{ authenticated: boolean }>('/api/v1/auth/logout');
      authState = 'signed-out';
      await goto('/login');
    } catch (cause) {
      const apiError = cause as ApiError;
      error = apiError.message ?? 'Could not log out';
    }
  }
</script>

{#if authState === 'loading'}
  <p class="p-6 text-slate-300" role="status">Checking session…</p>
{:else if authState === 'authenticated'}
  <header class="flex items-center justify-between border-b border-slate-700 bg-slate-950 px-6 py-4 text-slate-100">
    <a href="/" class="font-semibold">Betrayal Admin</a>
    <nav class="flex items-center gap-4" aria-label="Admin navigation">
      <a href="/players" class="text-slate-300 hover:text-white">Players</a>
      <button type="button" class="text-slate-300 hover:text-white" onclick={logout}>Log out</button>
    </nav>
  </header>
  {@render children()}
{:else if authState === 'signed-out'}
  <p class="p-6 text-slate-300" role="status">Signed out</p>
{:else}
  {#if error}<p class="p-6 text-red-300" role="alert">{error}</p>{/if}
  {@render children()}
{/if}
