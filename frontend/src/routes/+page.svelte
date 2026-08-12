<script lang="ts">
  import { onMount } from 'svelte';
  import { createApiClient } from '$lib/api/client';

  type Dashboard = { cycle: { phase: string; number: number }; players: { alive: number; dead: number; total: number } };
  const quickActions = [
    { label: 'Advance cycle', detail: 'Move the game to the next phase', href: '/cycle', tone: 'teal' },
    { label: 'Manage players', detail: 'Create, edit, and manage inventories', href: '/players', tone: 'indigo' },
    { label: 'Configure channels', detail: 'Validate Discord game channels', href: '/channels', tone: 'violet' },
    { label: 'Sync catalog', detail: 'Preview and apply spreadsheet data', href: '/sync', tone: 'amber' }
  ];
  let dashboard = $state<Dashboard | null>(null);
  let loadError = $state<string | null>(null);
  let loading = $state(true);

  onMount(async () => {
    try { dashboard = await createApiClient().get<Dashboard>('/api/v1/dashboard'); }
    catch (error) { loadError = error instanceof Error ? error.message : 'Could not load dashboard'; }
    finally { loading = false; }
  });
</script>

<svelte:head>
  <title>Dashboard | Betrayal Admin</title>
  <meta name="description" content="Betrayal game administration dashboard" />
</svelte:head>

<main class="min-h-screen bg-[#080b12] p-6 text-slate-100 sm:p-10">
  <div class="mx-auto max-w-6xl">
    <header class="border-b border-slate-800 pb-8">
      <p class="text-xs font-semibold uppercase tracking-[0.2em] text-teal-300">Operations console</p>
      <div class="mt-3 flex flex-wrap items-end justify-between gap-4">
        <div><h1 class="text-4xl font-semibold tracking-tight">Game dashboard</h1><p class="mt-3 text-slate-400">A live control surface for the current Betrayal game.</p></div>
        {#if dashboard}<div class="rounded border border-slate-700 bg-[#0d111b] px-4 py-3 text-right"><p class="text-xs uppercase tracking-wide text-slate-500">Current cycle</p><p class="mt-1 text-lg font-medium text-teal-200">{dashboard.cycle.phase} {dashboard.cycle.number}</p></div>{/if}
      </div>
    </header>

    {#if loading}
      <p class="py-12 text-slate-300" role="status">Loading dashboard…</p>
    {:else if loadError}
      <p class="py-12 text-red-300" role="alert">{loadError}</p>
    {:else if dashboard}
      <section class="grid gap-4 py-8 sm:grid-cols-3" aria-label="Game metrics">
        <article class="rounded border border-slate-800 bg-[#0d111b] p-5"><p class="text-xs uppercase tracking-wide text-slate-500">Players</p><p class="mt-3 text-3xl font-semibold">{dashboard.players.total} players</p><p class="mt-2 text-sm text-slate-400"><span class="text-emerald-300">{dashboard.players.alive} alive</span> · <span>{dashboard.players.dead} dead</span></p></article>
        <article class="rounded border border-slate-800 bg-[#0d111b] p-5"><p class="text-xs uppercase tracking-wide text-slate-500">Phase</p><p class="mt-3 text-3xl font-semibold">{dashboard.cycle.phase}</p><p class="mt-2 text-sm text-slate-400">Cycle {dashboard.cycle.number}</p></article>
        <article class="rounded border border-slate-800 bg-[#0d111b] p-5"><p class="text-xs uppercase tracking-wide text-slate-500">System</p><p class="mt-3 text-3xl font-semibold text-emerald-300">Online</p><p class="mt-2 text-sm text-slate-400">Admin API responding · use healthcheck for readiness</p></article>
      </section>
      <section aria-labelledby="quick-actions-heading">
        <div class="flex items-end justify-between"><div><p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Command center</p><h2 id="quick-actions-heading" class="mt-2 text-2xl font-semibold">Quick actions</h2></div><a href="/healthcheck" class="text-sm text-teal-300 hover:underline">View healthcheck →</a></div>
        <div class="mt-5 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {#each quickActions as action}<a href={action.href} class="group rounded border border-slate-800 bg-[#0d111b] p-5 transition hover:-translate-y-0.5 hover:border-slate-600"><div class="flex items-start justify-between"><h3 class="font-medium group-hover:text-teal-200">{action.label}</h3><span class="text-slate-600 transition group-hover:text-teal-300">↗</span></div><p class="mt-3 text-sm leading-6 text-slate-400">{action.detail}</p></a>{/each}
        </div>
      </section>
    {:else}
      <p class="py-12 text-slate-300">No dashboard data is available.</p>
    {/if}
  </div>
</main>
