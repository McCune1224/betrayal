<script lang="ts">
  import { onMount } from 'svelte';
  import { afterNavigate, goto } from '$app/navigation';
  import { createApiClient, type ApiError } from '$lib/api/client';

  let { children } = $props();
  let authState = $state<'loading' | 'authenticated' | 'signed-out' | 'login'>('loading');
  let error = $state<string | null>(null);
  let lastPath = $state(typeof window !== 'undefined' ? window.location.pathname : '');
  let menuOpen = $state(false);

  const onLoginRoute = () => typeof window !== 'undefined' && window.location.pathname === '/login';
  const sections = [
    { label: 'Overview', links: [{ label: 'Dashboard', href: '/', glyph: '◈' }, { label: 'Players', href: '/players', glyph: '♙' }] },
    {
      label: 'Game operations',
      links: [
        { label: 'Cycle', href: '/cycle', glyph: '◌' },
        { label: 'Channels', href: '/channels', glyph: '⌁' },
        { label: 'Votes', href: '/votes', glyph: '◇' },
        { label: 'Setup', href: '/setup', glyph: '✧' },
        { label: 'Whispers', href: '/whispers', glyph: '⌇' },
        { label: 'Healthcheck', href: '/healthcheck', glyph: '⊙' }
      ]
    },
    {
      label: 'Catalog',
      links: [
        { label: 'Roles', href: '/roles', glyph: '♜' },
        { label: 'Items', href: '/items', glyph: '◈' },
        { label: 'Abilities', href: '/abilities', glyph: '✦' },
        { label: 'Statuses', href: '/statuses', glyph: '☽' }
      ]
    },
    {
      label: 'System',
      links: [
        { label: 'Sync', href: '/sync', glyph: '⇄' },
        { label: 'Audit log', href: '/admin/audit', glyph: '≡' },
        { label: 'Migrations', href: '/admin/migrations', glyph: '⟲' },
        { label: 'Reset game', href: '/admin/reset', glyph: '⊗' },
        { label: 'Redeploy', href: '/admin/redeploy', glyph: '↗' }
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

  function closeMenu() {
    menuOpen = false;
  }
</script>

{#if authState === 'loading'}
  <p class="p-6 text-slate-300" role="status">Checking session…</p>
{:else if authState === 'authenticated'}
  <div class="admin-shell min-h-screen text-slate-100 lg:flex">
    <aside class="admin-sidebar border-b border-slate-800 lg:min-h-screen lg:w-72 lg:border-b-0 lg:border-r" aria-label="Admin navigation">
      <div class="mobile-nav-bar flex items-center justify-between px-4 py-3 lg:block">
        <a href="/" class="brand-mark"><span class="brand-glyph">◇</span><span>Betrayal <small>ADMIN CONSOLE</small></span></a>
        <button type="button" class="menu-toggle lg:hidden" aria-expanded={menuOpen} aria-controls="mobile-admin-menu" onclick={() => (menuOpen = !menuOpen)}>
          <span class="menu-toggle-icon" aria-hidden="true">{menuOpen ? '×' : '☰'}</span>
          <span>{menuOpen ? 'Close' : 'Menu'}</span>
        </button>
      </div>
      <nav id="mobile-admin-menu" class:mobile-menu-open={menuOpen} class="mobile-admin-menu" aria-label="Admin sections">
        {#each sections as section}
          <div class="mobile-menu-section">
            <p>{section.label}</p>
            <div>
              {#each section.links as link}
                <a href={link.href} aria-current={lastPath === link.href ? 'page' : undefined} class={`nav-link ${lastPath === link.href ? 'nav-link-active' : ''}`} onclick={closeMenu}><span aria-hidden="true">{link.glyph}</span>{link.label}</a>
              {/each}
            </div>
          </div>
        {/each}
        <button type="button" class="mobile-menu-logout" onclick={logout}>Log out</button>
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
