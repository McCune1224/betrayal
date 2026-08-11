<script lang="ts">
  import { onMount } from 'svelte';

  import { createApiClient } from '$lib/api/client';

  type Player = {
    id: number;
    alive: boolean;
    coins: number;
    luck: number;
    item_limit: number;
    alignment: string;
    role: string;
  };

  let players = $state<Player[] | null>(null);
  let loadError = $state<string | null>(null);
  let loading = $state(true);

  onMount(async () => {
    try {
      players = await createApiClient().get<Player[]>('/api/v1/players');
    } catch (error) {
      loadError = error instanceof Error ? error.message : 'Could not load players';
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>Players | Betrayal Admin</title>
  <meta name="description" content="Betrayal game players" />
</svelte:head>

<main class="min-h-screen bg-slate-950 p-6 text-slate-100">
  <header class="mx-auto flex max-w-5xl items-end justify-between border-b border-slate-700 pb-5">
    <div><p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Betrayal Admin</p>
    <h1 class="mt-2 text-3xl font-semibold">Players</h1></div>
    <a href="/players/new" class="rounded border border-teal-400 px-4 py-2 text-sm text-teal-300 hover:bg-teal-400/10">New player</a>
  </header>

  {#if loading}
    <p class="mx-auto max-w-5xl py-10 text-slate-300" role="status">Loading players…</p>
  {:else if loadError}
    <p class="mx-auto max-w-5xl py-10 text-red-300" role="alert">{loadError}</p>
  {:else if players && players.length === 0}
    <p class="mx-auto max-w-5xl py-10 text-slate-300">No players are available.</p>
  {:else if players}
    <div class="mx-auto max-w-5xl overflow-x-auto py-8">
      <table class="w-full border-collapse text-left" aria-label="Players">
        <thead class="border-b border-slate-700 text-sm uppercase tracking-wide text-slate-400">
          <tr>
            <th class="p-3">ID</th>
            <th class="p-3">Role</th>
            <th class="p-3">Alignment</th>
            <th class="p-3">State</th>
            <th class="p-3">Coins</th>
            <th class="p-3">Luck</th>
            <th class="p-3">Item limit</th>
          </tr>
        </thead>
        <tbody>
          {#each players as player (player.id)}
            <tr class="border-b border-slate-800">
              <td class="p-3 font-mono"><a class="text-teal-300 hover:underline" href={`/players/${player.id}`}>{player.id}</a></td>
              <td class="p-3">{player.role}</td>
              <td class="p-3">{player.alignment}</td>
              <td class="p-3">{player.alive ? 'Alive' : 'Dead'}</td>
              <td class="p-3">{player.coins}</td>
              <td class="p-3">{player.luck}</td>
              <td class="p-3">{player.item_limit}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else}
    <p class="mx-auto max-w-5xl py-10 text-slate-300">No players are available.</p>
  {/if}
</main>
