<script lang="ts">
  import { onMount } from 'svelte'; import { createApiClient } from '$lib/api/client';
  type Data = { ready: boolean; discord_connected: boolean; channels: { admin: number; admin_ready: boolean; vote_ready: boolean; action_ready: boolean; lifeboard_ready: boolean }; players: { total: number; alive: number; dead: number }; cycle: { ready: boolean; day: number; phase: string } };
  let data = $state<Data | null>(null); let error = $state<string | null>(null);
  onMount(async () => { try { data = await createApiClient().get<Data>('/api/v1/ops/healthcheck'); } catch (e) { error = e instanceof Error ? e.message : 'Could not load readiness'; } });
</script>
<svelte:head><title>Healthcheck | Betrayal Admin</title></svelte:head>
<main class="min-h-screen bg-slate-950 p-6 text-slate-100"><div class="mx-auto max-w-4xl">
  <h1 class="text-3xl font-semibold">Healthcheck</h1>
  {#if !data && !error}<p role="status" class="py-10 text-slate-300">Loading readiness…</p>
  {:else if error}<p role="alert" class="py-10 text-red-300">{error}</p>
  {:else if data}<p class="mt-3 text-xl">{data.ready ? 'Ready' : 'Not ready'}</p><section class="mt-8 space-y-3 border border-slate-700 p-6"><p>Players: {data.players.alive} alive / {data.players.dead} dead</p><p>Cycle: {data.cycle.ready ? `${data.cycle.phase} ${data.cycle.day}` : 'unavailable'}</p><p>Discord: {data.discord_connected ? 'connected' : 'disabled'}</p><p>Admin channels: {data.channels.admin}</p></section>{/if}
</div></main>
