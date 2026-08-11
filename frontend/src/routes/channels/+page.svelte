<script lang="ts">
  import { onMount } from 'svelte';
  import { createApiClient } from '$lib/api/client';
  type Entry = { name: string; kind: string; channel_id?: string; status: string; note?: string };
  type Data = { discord_connected: boolean; entries: Entry[]; summary: { total: number; missing: number; configured: number; orphaned: number; unverified: number } };
  let data = $state<Data | null>(null); let error = $state<string | null>(null);
  onMount(async () => { try { data = await createApiClient().get<Data>('/api/v1/ops/channels'); } catch (e) { error = e instanceof Error ? e.message : 'Could not load channels'; } });
</script>
<svelte:head><title>Channels | Betrayal Admin</title></svelte:head>
<main class="min-h-screen bg-slate-950 p-6 text-slate-100"><div class="mx-auto max-w-5xl">
  <h1 class="text-3xl font-semibold">Channels</h1>
  {#if !data && !error}<p role="status" class="py-10 text-slate-300">Loading channels…</p>
  {:else if error}<p role="alert" class="py-10 text-red-300">{error}</p>
  {:else if data}<p class="mt-3 text-slate-400">Discord: {data.discord_connected ? 'connected' : 'disabled'}</p><p class="mt-2 text-slate-300">{data.summary.configured} configured · {data.summary.missing} missing</p>
    <div class="mt-8 space-y-3">{#each data.entries as entry (entry.kind + entry.name)}<article class="border border-slate-700 p-4"><div class="flex justify-between gap-4"><h2>{entry.name}</h2><span>{entry.status}</span></div>{#if entry.channel_id}<p class="mt-1 font-mono text-sm text-slate-400">{entry.channel_id}</p>{/if}</article>{/each}</div>{/if}
</div></main>
