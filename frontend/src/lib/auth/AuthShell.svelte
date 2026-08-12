<script lang="ts">
  import { onMount } from 'svelte';
  import { afterNavigate, goto } from '$app/navigation';
  import { createApiClient, type ApiError } from '$lib/api/client';

  let { children } = $props();
  let authState = $state<'loading' | 'authenticated' | 'signed-out' | 'login'>('loading');
  let error = $state<string | null>(null);
  let lastPath = typeof window !== 'undefined' ? window.location.pathname : '';

  const onLoginRoute = () => typeof window !== 'undefined' && window.location.pathname === '/login';
  const sections = [
    { label: 'Overview', links: [{ label: 'Dashboard', href: '/' }, { label: 'Players', href: '/players' }] },
    {
      label: 'Game operations',
      links: [
        { label: 'Cycle', href: '/cycle' },
        { label: 'Channels', href: '/channels' },
        { label: 'Votes', href: '/votes' },
        { label: 'Setup', href: '/setup' },
        { label: 'Healthcheck', href: '/healthcheck' }
      ]
    },
    {
      label: 'Catalog',
      links: [
        { label: 'Roles', href: '/roles' },
        { label: 'Items', href: '/items' },
        { label: 'Abilities', href: '/abilities' },
        { label: 'Statuses', href: '/statuses' }
      ]
    },
    {
      label: 'System',
      links: [
        { label: 'Sync', href: '/sync' },
        { label: 'Audit log', href: '/admin/audit' },
        { label: 'Migrations', href: '/admin/migrations' },
        { label: 'Reset game', href: '/admin/reset' },
        { label: 'Redeploy', href: '/admin/redeploy' }
      ]
    }
  ];

  async function checkSession() {
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
  }

  onMount(() => {
    void checkSession();
  });

  afterNavigate(() => {
    const path = window.location.pathname;
    if (path === lastPath) return;
    lastPath = path;
    void checkSession();
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
  <div class="min-h-screen bg-[#080b12] text-slate-100 lg:flex">
    <aside class="border-b border-slate-800 bg-[#0d111b] lg:min-h-screen lg:w-72 lg:border-b-0 lg:border-r" aria-label="Admin navigation">
      <div class="flex items-center justify-between px-5 py-5 lg:block">
        <a href="/" class="text-lg font-semibold tracking-tight">Betrayal <span class="text-teal-300">/</span> Admin</a>
        <button type="button" class="rounded border border-slate-700 px-3 py-2 text-xs text-slate-300 lg:hidden" onclick={logout}>Log out</button>
      </div>
      <nav class="flex gap-6 overflow-x-auto px-5 pb-5 lg:block lg:space-y-6 lg:px-4" aria-label="Admin sections">
        {#each sections as section}
          <div class="min-w-max lg:min-w-0">
            <p class="mb-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-slate-500">{section.label}</p>
            <div class="flex gap-1 lg:block">
              {#each section.links as link}
                <a href={link.href} class="block rounded px-3 py-2 text-sm text-slate-300 hover:bg-slate-800 hover:text-white">{link.label}</a>
              {/each}
            </div>
          </div>
        {/each}
      </nav>
      <div class="hidden border-t border-slate-800 p-4 lg:block">
        <button type="button" class="w-full rounded border border-slate-700 px-3 py-2 text-left text-sm text-slate-300 hover:border-slate-500 hover:text-white" onclick={logout}>Log out</button>
      </div>
    </aside>
    <div class="min-w-0 flex-1">
      {@render children()}
    </div>
  </div>
{:else if authState === 'signed-out'}
  <p class="p-6 text-slate-300" role="status">Signed out</p>
{:else}
  {#if error}<p class="p-6 text-red-300" role="alert">{error}</p>{/if}
  {@render children()}
{/if}
