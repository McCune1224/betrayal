<script lang="ts">
  import { onMount } from 'svelte';
  import { createApiClient } from '$lib/api/client';

  const confirmationPhrase = 'RESET BETRAYAL GAME';
  let summary = $state<Record<string, number> | null>(null);
  let error = $state('');
  let status = $state('');
  let confirm = $state('');
  let understand = $state(false);
  let busy = $state(false);
  let canReset = $derived(confirm === confirmationPhrase && understand && !busy);

  onMount(async () => {
    try {
      summary = (await createApiClient().get<{ summary: Record<string, number> }>('/api/v1/admin/reset')).summary;
    } catch (cause) {
      error = cause instanceof Error ? cause.message : 'Could not load reset preview';
    }
  });

  async function reset() {
    if (!canReset) {
      error = `Type ${confirmationPhrase} exactly and acknowledge the reset.`;
      return;
    }
    busy = true;
    error = '';
    status = '';
    try {
      await createApiClient().post('/api/v1/admin/reset', {
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ confirm, understand })
      });
      status = 'Game reset completed.';
      confirm = '';
      understand = false;
    } catch (cause) {
      error = cause instanceof Error ? cause.message : 'Reset failed';
    } finally {
      busy = false;
    }
  }
</script>

<svelte:head><title>Reset | Betrayal Admin</title></svelte:head>
<main class="min-h-screen bg-slate-950 p-6 text-slate-100">
  <div class="mx-auto max-w-3xl">
    <h1 class="text-3xl font-semibold">Reset game</h1>
    <p class="mt-3 text-red-300">Destructive operation. This permanently clears the current game data; production remains enabled and server confirmation is authoritative.</p>
    {#if summary}<pre class="mt-6 overflow-x-auto border border-slate-700 p-4">{JSON.stringify(summary, null, 2)}</pre>{/if}
    <label class="mt-6 block text-sm text-slate-300">Type {confirmationPhrase} exactly
      <input aria-label={confirmationPhrase} class="mt-2 block w-full bg-slate-900 p-3" bind:value={confirm} autocomplete="off" />
    </label>
    <label class="mt-4 flex items-start gap-2 text-sm text-slate-300">
      <input type="checkbox" aria-label="I understand this permanently clears the current game data" bind:checked={understand} />
      <span>I understand this permanently clears the current game data.</span>
    </label>
    <button class="mt-4 border border-red-500 px-4 py-2 disabled:cursor-not-allowed disabled:opacity-50" disabled={!canReset} onclick={reset}>{busy ? 'Resetting…' : 'Execute reset'}</button>
    {#if status}<p role="status" class="mt-4 text-emerald-300">{status}</p>{/if}
    {#if error}<p role="alert" class="mt-4 text-red-300">{error}</p>{/if}
  </div>
</main>
