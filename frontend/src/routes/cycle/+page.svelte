<script lang="ts">
  import { onMount } from 'svelte';
  import { createApiClient } from '$lib/api/client';
  type Cycle = { day: number; phase: string; is_elimination: boolean };
  let cycle = $state<Cycle | null>(null);
  let error = $state<string | null>(null);
  let busy = $state(false);
  const api = createApiClient();
  async function load() { try { cycle = await api.get<Cycle>('/api/v1/ops/cycle'); error = null; } catch (e) { error = e instanceof Error ? e.message : 'Could not load cycle'; } }
  async function advance() { if (!confirm('Advance the game to the next phase?')) return; busy = true; try { cycle = await api.post<Cycle>('/api/v1/ops/cycle/advance'); error = null; } catch (e) { error = e instanceof Error ? e.message : 'Could not advance cycle'; } finally { busy = false; } }
  async function setCycle(event: SubmitEvent) { event.preventDefault(); const form = new FormData(event.currentTarget as HTMLFormElement); busy = true; try { cycle = await api.post<Cycle>('/api/v1/ops/cycle/set', { body: JSON.stringify({ phase: form.get('phase'), day: Number(form.get('day')) }), headers: { 'Content-Type': 'application/json' } }); error = null; } catch (e) { error = e instanceof Error ? e.message : 'Could not set cycle'; } finally { busy = false; } }
  onMount(load);
</script>
<svelte:head><title>Cycle | Betrayal Admin</title></svelte:head>
<main class="min-h-screen bg-slate-950 p-6 text-slate-100"><div class="mx-auto max-w-4xl">
  <h1 class="text-3xl font-semibold">Cycle</h1>
  {#if error}<p role="alert" class="mt-6 text-red-300">{error}</p>{/if}
  {#if !cycle && !error}<p role="status" class="py-10 text-slate-300">Loading cycle…</p>
  {:else if cycle}<section class="mt-8 border border-slate-700 p-6"><p class="text-sm uppercase tracking-wide text-slate-400">Current phase</p><p class="mt-2 text-2xl">{cycle.phase} {cycle.day}</p><p class="mt-2 text-slate-400">{cycle.is_elimination ? 'Elimination is active' : 'Day phase is active'}</p>
    <button class="mt-6 border border-emerald-400 px-4 py-2 text-emerald-300" disabled={busy} onclick={advance}>Advance to next phase</button>
    <form class="mt-8 flex flex-wrap gap-3" onsubmit={setCycle}><select name="phase" class="bg-slate-900 p-2"><option>Day</option><option>Elimination</option></select><input name="day" type="number" min="0" value={cycle.day} class="w-24 bg-slate-900 p-2" aria-label="Day" /><button class="border border-slate-500 px-4 py-2" disabled={busy}>Set cycle</button></form>
  </section>{/if}
</div></main>
