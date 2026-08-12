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

<main class="page-shell">
  <div class="mx-auto max-w-6xl">
    <header class="page-header">
      <p class="eyebrow">The mirror room · operations console</p>
      <div class="mt-3 flex flex-wrap items-end justify-between gap-4">
        <div><h1 class="display-title">Game dashboard</h1><p class="lede">A live control surface for the current Betrayal game.</p></div>
        {#if dashboard}<div class="stat-card"><p class="eyebrow">Current cycle</p><p class="mt-1 text-lg font-medium text-slate-100">{dashboard.cycle.phase} {dashboard.cycle.number}</p></div>{/if}
      </div>
    </header>

    {#if loading}
      <p class="py-12 text-slate-300" role="status">Loading dashboard…</p>
    {:else if loadError}
      <p class="py-12 text-red-300" role="alert">{loadError}</p>
    {:else if dashboard}
      <section class="stat-grid" aria-label="Game metrics">
        <article class="stat-card"><p class="eyebrow">Players</p><strong>{dashboard.players.total} players</strong><span><span class="text-slate-200">{dashboard.players.alive} alive</span> · <span>{dashboard.players.dead} dead</span></span></article>
        <article class="stat-card"><p class="eyebrow">Phase</p><strong>{dashboard.cycle.phase}</strong><span>Cycle {dashboard.cycle.number}</span></article>
        <article class="stat-card"><p class="eyebrow">System</p><strong class="text-slate-100">Online</strong><span>Admin API responding</span></article>
      </section>
      <section aria-labelledby="quick-actions-heading">
        <div class="flex items-end justify-between"><div><p class="eyebrow">Command center</p><h2 id="quick-actions-heading" class="mt-2 text-2xl font-semibold">Quick actions</h2></div><a href="/healthcheck" class="text-sm text-slate-200 hover:underline">View healthcheck →</a></div>
        <div class="mt-5 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {#each quickActions as action}<a href={action.href} class="player-card group"><div class="flex items-start justify-between"><h3 class="font-medium group-hover:text-slate-100">{action.label}</h3><span class="text-slate-500 transition group-hover:text-slate-100">↗</span></div><p class="mt-3 text-sm leading-6 text-slate-400">{action.detail}</p></a>{/each}
        </div>
      </section>
    {:else}
      <p class="py-12 text-slate-300">No dashboard data is available.</p>
    {/if}
  </div>
</main>
