<script lang="ts">
  import { onMount } from 'svelte';

  import { createApiClient } from '$lib/api/client';

  type Dashboard = {
    cycle: { phase: string; number: number };
    players: { alive: number; dead: number; total: number };
  };

  let dashboard = $state<Dashboard | null>(null);
  let loadError = $state<string | null>(null);
  let loading = $state(true);

  onMount(async () => {
    try {
      dashboard = await createApiClient().get<Dashboard>('/api/v1/dashboard');
    } catch (error) {
      loadError = error instanceof Error ? error.message : 'Could not load dashboard';
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>Dashboard | Betrayal Admin</title>
  <meta name="description" content="Betrayal Discord game administration dashboard" />
</svelte:head>

<main class="min-h-screen bg-slate-950 p-6 text-slate-100">
  <header class="mx-auto max-w-4xl border-b border-slate-700 pb-5">
    <p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Betrayal Admin</p>
    <h1 class="mt-2 text-3xl font-semibold">Game dashboard</h1>
  </header>

  {#if loading}
    <p class="mx-auto max-w-4xl py-10 text-slate-300" role="status">Loading dashboard…</p>
  {:else if loadError}
    <p class="mx-auto max-w-4xl py-10 text-red-300" role="alert">{loadError}</p>
  {:else if dashboard}
    <section class="mx-auto grid max-w-4xl gap-4 py-8 sm:grid-cols-2" aria-label="Game metrics">
      <article class="border border-slate-700 bg-slate-900 p-5">
        <p class="text-sm uppercase tracking-wide text-slate-400">Current cycle</p>
        <p class="mt-2 text-2xl font-semibold">{dashboard.cycle.phase} {dashboard.cycle.number}</p>
      </article>
      <article class="border border-slate-700 bg-slate-900 p-5">
        <p class="text-sm uppercase tracking-wide text-slate-400">Players</p>
        <p class="mt-2 text-2xl font-semibold">{dashboard.players.total} players</p>
        <p class="mt-2 text-sm text-emerald-300">{dashboard.players.alive} alive</p>
        <p class="text-sm text-slate-300">{dashboard.players.dead} dead</p>
      </article>
    </section>
  {:else}
    <p class="mx-auto max-w-4xl py-10 text-slate-300">No dashboard data is available.</p>
  {/if}
</main>
