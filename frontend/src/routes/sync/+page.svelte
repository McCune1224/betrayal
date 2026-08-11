<script lang="ts">
  import { onMount } from 'svelte';
  import { createApiClient } from '$lib/api/client';

  type Source = { id: number; name: string; kind: string; alignment: string; url: string; enabled: boolean };
  let sources = $state<Source[]>([]);
  let error = $state<string | null>(null);
  let status = $state<string | null>(null);
  let preview = $state<{ source: Source; status: string; counts?: Record<string, number>; error?: string }[]>([]);
  let busy = $state<number | 'preview' | null>(null);
  const api = createApiClient();

  async function load() {
    try { sources = (await api.get<{ sources: Source[] }>('/api/v1/sync/sources')).sources; }
    catch (cause) { error = cause instanceof Error ? cause.message : 'Could not load sync sources'; }
  }
  async function saveSource(source: Source) {
    busy = source.id; error = null; status = null;
    try {
      const saved = await api.put<Source>(`/api/v1/sync/sources/${source.id}`, { body: JSON.stringify({ url: source.url, enabled: source.enabled }), headers: { 'Content-Type': 'application/json' } });
      sources = sources.map((candidate) => candidate.id === source.id ? saved : candidate);
      status = `${source.name} updated.`;
    } catch (cause) { error = cause instanceof Error ? cause.message : 'Could not update sync source'; }
    finally { busy = null; }
  }
  async function loadPreview() {
    busy = 'preview'; error = null; status = null;
    try { const result = await api.post<{ previews: typeof preview }>('/api/v1/sync/preview'); preview = result.previews; status = 'Read-only preview generated. Review the planned changes below.'; }
    catch (cause) { error = cause instanceof Error ? cause.message : 'Preview failed'; }
    finally { busy = null; }
  }
  async function apply(source: Source) {
    if (!source.enabled || !source.url) { error = 'Enable and configure the source before applying it.'; return; }
    if (!confirm(`Apply ${source.name} to the game catalog?`)) return;
    busy = source.id; error = null; status = null;
    try { await api.post('/api/v1/sync/apply', { body: JSON.stringify({ source_id: source.id }), headers: { 'Content-Type': 'application/json' } }); status = `${source.name} applied successfully.`; }
    catch (cause) { error = cause instanceof Error ? cause.message : 'Sync apply failed'; }
    finally { busy = null; }
  }
  onMount(load);
</script>

<svelte:head><title>Sync | Betrayal Admin</title></svelte:head>
<main class="min-h-screen bg-[#080b12] p-6 text-slate-100 sm:p-10">
  <div class="mx-auto max-w-5xl">
    <p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">System operations</p>
    <h1 class="mt-3 text-4xl font-semibold tracking-tight">Catalog sync</h1>
    <p class="mt-3 max-w-3xl text-slate-400">Configure trusted spreadsheet sources, preview the current inputs, then apply one source at a time. The server re-fetches and validates the source during apply.</p>
    {#if error}<p role="alert" class="mt-6 rounded border border-red-500/40 bg-red-500/5 p-4 text-red-200">{error}</p>{/if}
    {#if status}<p role="status" class="mt-6 rounded border border-emerald-500/40 bg-emerald-500/5 p-4 text-emerald-200">{status}</p>{/if}
    <button type="button" class="mt-8 rounded border border-teal-400 px-4 py-2 text-sm text-teal-300 hover:bg-teal-400/10 disabled:opacity-50" disabled={busy !== null} onclick={loadPreview}>{busy === 'preview' ? 'Loading preview…' : 'Fetch preview'}</button>
    <section class="mt-6 space-y-3">{#each preview as result (result.source.id)}<article class="rounded border border-slate-800 bg-[#0d111b] p-4"><div class="flex flex-wrap justify-between gap-3"><h2 class="font-medium">{result.source.name}</h2><span class={result.status === 'ready' ? 'text-emerald-300' : result.status === 'failed' ? 'text-red-300' : 'text-slate-500'}>{result.status}</span></div>{#if result.counts}<p class="mt-2 text-sm text-slate-400">{#each Object.entries(result.counts) as [action, count], i}{i ? ' · ' : ''}{count} {action}{/each}</p>{/if}{#if result.error}<p class="mt-2 text-sm text-red-300">{result.error}</p>{/if}</article>{/each}</section>
    <section class="mt-8 space-y-4">
      {#each sources as source (source.id)}
        <article class="rounded border border-slate-800 bg-[#0d111b] p-5">
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div><h2 class="text-lg font-medium">{source.name}</h2><p class="mt-1 text-sm text-slate-500">{source.kind} · {source.alignment}</p></div>
            <span class={source.enabled ? 'text-emerald-300' : 'text-slate-500'}>{source.enabled ? 'Enabled' : 'Disabled'}</span>
          </div>
          <div class="mt-5 flex flex-col gap-3 md:flex-row md:items-end">
            <label class="min-w-0 flex-1 text-sm text-slate-400">Source URL<input class="mt-1 w-full rounded bg-slate-900 px-3 py-2 text-slate-200" bind:value={source.url} /></label>
            <label class="flex items-center gap-2 pb-2 text-sm text-slate-300"><input type="checkbox" bind:checked={source.enabled} /> Enabled</label>
            <button type="button" class="rounded border border-slate-600 px-3 py-2 text-sm hover:border-slate-400 disabled:opacity-50" disabled={busy !== null} onclick={() => saveSource(source)}>Save</button>
            <button type="button" class="rounded border border-amber-500 px-3 py-2 text-sm text-amber-200 hover:bg-amber-500/10 disabled:opacity-50" disabled={busy !== null} onclick={() => apply(source)}>Apply</button>
          </div>
        </article>
      {:else}
        <p class="rounded border border-slate-800 p-6 text-slate-400">No sync sources configured.</p>
      {/each}
    </section>
  </div>
</main>
