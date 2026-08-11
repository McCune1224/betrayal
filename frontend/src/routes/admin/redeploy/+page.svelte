<script lang="ts">
  import { createApiClient, type ApiError } from '$lib/api/client';

  let busy = $state(false);
  let status = $state<string | null>(null);
  let error = $state<string | null>(null);

  async function redeploy() {
    if (!confirm('Restart the latest production deployment?')) return;
    busy = true;
    status = null;
    error = null;
    try {
      await createApiClient().post('/api/v1/admin/redeploy');
      status = 'Redeploy requested. Railway is restarting the latest deployment.';
    } catch (cause) {
      const apiError = cause as ApiError;
      error = apiError.message ?? 'Could not request redeploy';
    } finally {
      busy = false;
    }
  }
</script>

<svelte:head>
  <title>Redeploy | Betrayal Admin</title>
</svelte:head>

<main class="min-h-screen bg-[#080b12] p-6 text-slate-100 sm:p-10">
  <div class="mx-auto max-w-3xl">
    <p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">System operations</p>
    <h1 class="mt-3 text-4xl font-semibold tracking-tight">Redeploy</h1>
    <p class="mt-4 max-w-2xl text-slate-400">Restart the latest Railway deployment after a verified release. This does not create a new deployment or change environment variables.</p>

    <section class="mt-10 rounded-lg border border-amber-500/40 bg-amber-500/5 p-6">
      <h2 class="text-lg font-medium text-amber-200">Production action</h2>
      <p class="mt-2 text-sm leading-6 text-slate-300">Only use this after checking the pushed commit and migration status. The server performs the authenticated Railway operation; the browser never receives Railway credentials.</p>
      <button type="button" class="mt-6 rounded border border-amber-400 px-4 py-2 text-sm font-medium text-amber-200 hover:bg-amber-400/10 disabled:opacity-50" disabled={busy} onclick={redeploy}>
        {busy ? 'Requesting restart…' : 'Restart latest deployment'}
      </button>
    </section>

    {#if status}<p class="mt-6 rounded border border-emerald-500/40 bg-emerald-500/5 p-4 text-emerald-200" role="status">{status}</p>{/if}
    {#if error}<p class="mt-6 rounded border border-red-500/40 bg-red-500/5 p-4 text-red-200" role="alert">{error}</p>{/if}
  </div>
</main>
