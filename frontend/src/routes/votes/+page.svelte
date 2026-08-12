<script lang="ts">
  import { onMount } from 'svelte'; import { createApiClient } from '$lib/api/client';
  type Data = { cycle: { day: number; phase: string }; votes: { id: number; voter_id: number; target_id: number; weight: number }[]; tallies: { target_id: number; total_votes: number; vote_count: number }[]; total_votes: number };
  let data = $state<Data | null>(null); let error = $state<string | null>(null);
  onMount(async () => { try { data = await createApiClient().get<Data>('/api/v1/ops/votes'); } catch (e) { error = e instanceof Error ? e.message : 'Could not load votes'; } });
</script>
<svelte:head><title>Votes | Betrayal Admin</title></svelte:head>
<main class="min-h-screen bg-slate-950 p-6 text-slate-100"><div class="mx-auto max-w-5xl">
  <h1 class="text-3xl font-semibold">Votes</h1>
  {#if !data && !error}<p role="status" class="py-10 text-slate-300">Loading votes…</p>
  {:else if error}<p role="alert" class="py-10 text-red-300">{error}</p>
  {:else if data}<p class="mt-3 text-slate-400">{data.cycle.phase} {data.cycle.day} · {data.total_votes} total votes</p>
    <section class="mt-8"><h2 class="text-xl font-semibold">Tallies</h2>{#if data.tallies.length === 0}<p class="mt-3 border border-slate-800 p-4 text-slate-400">No votes have been cast for this cycle.</p>{:else}<div class="mt-3 space-y-3">{#each data.tallies as tally (tally.target_id)}<article class="border border-slate-700 p-4"><span>Target {tally.target_id}</span><span class="ml-4">{tally.total_votes} votes</span><span class="ml-4 text-slate-400">({tally.vote_count} ballots)</span></article>{/each}</div>{/if}</section>
  {/if}
</div></main>
