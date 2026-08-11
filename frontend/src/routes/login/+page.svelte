<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  import { ApiError, createApiClient } from '$lib/api/client';

  let password = $state('');
  let error = $state<string | null>(null);
  let submitting = $state(false);
  let loggedIn = $state(false);
  let checkingSession = $state(true);

  onMount(async () => {
    try {
      const session = await createApiClient().get<{ authenticated: boolean }>('/api/v1/auth/session');
      if (session.authenticated) {
        loggedIn = true;
      }
    } catch (cause) {
      error = cause instanceof Error ? cause.message : 'Could not load session';
    } finally {
      checkingSession = false;
    }
  });

  async function submit() {
    submitting = true;
    error = null;
    try {
      await createApiClient().post<{ authenticated: boolean }>('/api/v1/auth/login', {
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password })
      });
      loggedIn = true;
      password = '';
      await goto('/');
    } catch (cause) {
      error = cause instanceof ApiError ? cause.message : 'Could not log in';
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>Log in | Betrayal Admin</title>
  <meta name="description" content="Log in to Betrayal Admin" />
</svelte:head>

<main class="flex min-h-screen items-center justify-center bg-slate-950 p-6 text-slate-100">
  <section class="w-full max-w-md border border-slate-700 bg-slate-900 p-8">
    {#if checkingSession}
      <p role="status" class="text-slate-300">Checking session…</p>
    {:else if loggedIn}
      <p role="status" class="text-emerald-300">Logged in. Loading dashboard…</p>
    {:else}
      <p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Betrayal Admin</p>
      <h1 class="mt-2 text-3xl font-semibold">Log in</h1>
      <p class="mt-3 text-slate-300">Enter the administrator password to continue.</p>
      <form class="mt-6 space-y-5" onsubmit={(event) => { event.preventDefault(); void submit(); }}>
        <div>
          <label for="password" class="mb-2 block text-sm font-medium text-slate-200">Password</label>
          <input id="password" name="password" type="password" required bind:value={password}
            class="w-full border border-slate-600 bg-slate-950 px-3 py-3 text-slate-100 outline-none focus:border-indigo-400" />
        </div>
        {#if error}<p role="alert" class="text-red-300">{error}</p>{/if}
        <button type="submit" disabled={submitting} class="w-full bg-indigo-500 px-4 py-3 font-semibold text-white disabled:opacity-60">
          {submitting ? 'Logging in…' : 'Log in'}
        </button>
      </form>
    {/if}
  </section>
</main>
