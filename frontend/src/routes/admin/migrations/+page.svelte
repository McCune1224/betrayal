<script lang="ts">
  import { onMount } from 'svelte';
  import { createApiClient } from '$lib/api/client';

  type Migration = { version: number; name: string; applied: boolean; dirty: boolean };
  let migrations = $state<Migration[]>([]);
  let error = $state<string | null>(null);
  let status = $state<string | null>(null);
  let busy = $state(false);
  let confirm = $state('');
  const api = createApiClient();

  async function load() {
    try { migrations = (await api.get<{ migrations: Migration[] }>('/api/v1/admin/migrations')).migrations; }
    catch (cause) { error = cause instanceof Error ? cause.message : 'Could not load migrations'; }
  }
  async function migrateUp() {
    if (!window.confirm('Apply all pending migrations?')) return;
    busy = true; error = null; status = null;
    try { await api.post('/api/v1/admin/migrations/up'); status = 'Pending migrations applied.'; await load(); }
    catch (cause) { error = cause instanceof Error ? cause.message : 'Could not apply migrations'; }
    finally { busy = false; }
  }
  async function migrateDown() {
    const latest = migrations.filter((migration) => migration.applied).at(-1);
    if (!latest || confirm !== latest.name) { error = latest ? `Type ${latest.name} exactly to roll back.` : 'No applied migration is available.'; return; }
    busy = true; error = null; status = null;
    try { await api.post('/api/v1/admin/migrations/down', { body: JSON.stringify({ steps: 1, confirm }), headers: { 'Content-Type': 'application/json' } }); status = `Rolled back ${latest.name}.`; confirm = ''; await load(); }
    catch (cause) { error = cause instanceof Error ? cause.message : 'Could not roll back migration'; }
    finally { busy = false; }
  }
  onMount(load);
</script>

<svelte:head><title>Migrations | Betrayal Admin</title></svelte:head>
<main class="min-h-screen bg-[#080b12] p-6 text-slate-100 sm:p-10">
  <div class="mx-auto max-w-5xl">
    <p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">System operations</p>
    <h1 class="mt-3 text-4xl font-semibold tracking-tight">Database migrations</h1>
    <p class="mt-3 text-slate-400">The embedded migration set is the source of truth for the running service.</p>
    {#if error}<p role="alert" class="mt-6 rounded border border-red-500/40 bg-red-500/5 p-4 text-red-200">{error}</p>{/if}
    {#if status}<p role="status" class="mt-6 rounded border border-emerald-500/40 bg-emerald-500/5 p-4 text-emerald-200">{status}</p>{/if}
    <section class="mt-8 flex flex-wrap gap-3 rounded border border-slate-800 bg-[#0d111b] p-5">
      <button type="button" class="rounded border border-teal-400 px-4 py-2 text-sm text-teal-300 disabled:opacity-50" disabled={busy} onclick={migrateUp}>Apply pending migrations</button>
      <div class="flex min-w-[min(100%,28rem)] flex-1 gap-2">
        <input aria-label="Rollback confirmation" bind:value={confirm} placeholder="Type latest migration name to roll back" class="min-w-0 flex-1 rounded bg-slate-900 px-3 py-2 text-sm" />
        <button type="button" class="rounded border border-red-500 px-4 py-2 text-sm text-red-300 disabled:opacity-50" disabled={busy} onclick={migrateDown}>Rollback latest</button>
      </div>
    </section>
    <section class="mt-8 overflow-x-auto rounded border border-slate-800 bg-[#0d111b]">
      {#if migrations.length === 0}<p class="p-6 text-slate-400">Loading migrations…</p>{:else}<table class="w-full text-left"><thead class="border-b border-slate-800 text-xs uppercase tracking-wide text-slate-500"><tr><th class="p-4">Version</th><th class="p-4">Name</th><th class="p-4">State</th></tr></thead><tbody>{#each migrations as migration (migration.version)}<tr class="border-b border-slate-800/70"><td class="p-4 font-mono">{migration.version}</td><td class="p-4">{migration.name}</td><td class="p-4">{migration.dirty ? 'Dirty — manual recovery required' : migration.applied ? 'Applied' : 'Pending'}</td></tr>{/each}</tbody></table>{/if}
    </section>
  </div>
</main>
