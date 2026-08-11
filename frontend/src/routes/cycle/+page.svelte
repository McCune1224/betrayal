<script lang="ts">
  import { onMount } from 'svelte';
  import { createApiClient } from '$lib/api/client';
  type Cycle = { day: number; phase: string; is_elimination: boolean };
  let cycle = $state<Cycle | null>(null);
  let error = $state<string | null>(null);
  onMount(async () => { try { cycle = await createApiClient().get<Cycle>('/api/v1/ops/cycle'); } catch (e) { error = e instanceof Error ? e.message : 'Could not load cycle'; } });
</script>
<svelte:head><title>Cycle | Betrayal Admin</title></svelte:head>
<main class="min-h-screen bg-slate-950 p-6 text-slate-100"><div class="mx-auto max-w-4xl">
  <h1 class="text-3xl font-semibold">Cycle</h1>
  {#if !cycle && !error}<p role="status" class="py-10 text-slate-300">Loading cycle…</p>
  {:else if error}<p role="alert" class="py-10 text-red-300">{error}</p>
  {:else if cycle}<section class="mt-8 border border-slate-700 p-6"><p class="text-sm uppercase tracking-wide text-slate-400">Current phase</p><p class="mt-2 text-2xl">{cycle.phase} {cycle.day}</p><p class="mt-2 text-slate-400">{cycle.is_elimination ? 'Elimination is active' : 'Day phase is active'}</p></section>{/if}
</div></main>
