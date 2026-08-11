<script lang="ts">
  import { onMount } from 'svelte'; import { createApiClient } from '$lib/api/client';
  type Entry = { name: string; kind: string; channel_id?: string; status: string; note?: string };
  type Data = { discord_connected: boolean; entries: Entry[]; summary: { total: number; missing: number; configured: number; orphaned: number; unverified: number } };
  let data = $state<Data | null>(null); let error = $state<string | null>(null); let busy = $state(false); const api = createApiClient();
  async function load() { try { data = await api.get<Data>('/api/v1/ops/channels'); error = null; } catch (e) { error = e instanceof Error ? e.message : 'Could not load channels'; } }
  async function update(event: SubmitEvent) { event.preventDefault(); const form = new FormData(event.currentTarget as HTMLFormElement); busy = true; try { data = await api.post<Data>('/api/v1/ops/channels', { body: JSON.stringify({ kind: form.get('kind'), channel_id: form.get('channel_id'), message_id: form.get('message_id') }), headers: { 'Content-Type': 'application/json' } }); } catch (e) { error = e instanceof Error ? e.message : 'Could not update channel'; } finally { busy = false; } }
  async function remove(entry: Entry) { if (!entry.channel_id || !confirm(`Delete ${entry.name} ${entry.channel_id}?`)) return; busy = true; try { await api.delete(`/api/v1/ops/channels/${entry.kind}/${entry.channel_id}`); await load(); } catch (e) { error = e instanceof Error ? e.message : 'Could not delete channel'; } finally { busy = false; } }
  onMount(load);
</script>
<svelte:head><title>Channels | Betrayal Admin</title></svelte:head>
<main class="min-h-screen bg-slate-950 p-6 text-slate-100"><div class="mx-auto max-w-5xl">
  <h1 class="text-3xl font-semibold">Channels</h1>
  {#if error}<p role="alert" class="mt-6 text-red-300">{error}</p>{/if}
  {#if !data && !error}<p role="status" class="py-10 text-slate-300">Loading channels…</p>
  {:else if data}<p class="mt-3 text-slate-400">Discord: {data.discord_connected ? 'connected' : 'disabled'} · {data.summary.configured} configured · {data.summary.missing} missing</p>
    <form class="mt-8 flex flex-wrap gap-3" onsubmit={update}><select name="kind" class="bg-slate-900 p-2"><option value="vote">Vote</option><option value="action">Action</option><option value="log">Command log</option><option value="admin">Admin</option><option value="lifeboard">Lifeboard</option></select><input name="channel_id" required placeholder="Channel ID" class="bg-slate-900 p-2" /><input name="message_id" placeholder="Lifeboard message ID" class="bg-slate-900 p-2" /><button class="border border-emerald-400 px-4 py-2" disabled={busy}>Save channel</button></form>
    <div class="mt-8 space-y-3">{#each data.entries as entry (entry.kind + entry.name + entry.channel_id)}<article class="border border-slate-700 p-4"><div class="flex justify-between gap-4"><h2>{entry.name}</h2><span>{entry.status}</span></div>{#if entry.channel_id}<p class="mt-1 font-mono text-sm text-slate-400">{entry.channel_id}</p>{#if entry.kind === 'admin'}<button class="mt-3 text-red-300" disabled={busy} onclick={() => remove(entry)}>Delete</button>{/if}{/if}{#if entry.note}<p class="mt-2 text-sm text-slate-400">{entry.note}</p>{/if}</article>{/each}</div>{/if}
</div></main>
