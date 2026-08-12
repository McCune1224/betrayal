<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import { createApiClient } from '$lib/api/client';
  import type { PlayerDetail } from '$lib/player-types';

  let player = $state<PlayerDetail | null>(null);
  let error = $state('');
  let loading = $state(true);
  let mutating = $state(false);
  let mutationError = $state('');
  const id = $derived(page.params.id);

  async function load() {
    loading = true;
    error = '';
    try { player = await createApiClient().get<PlayerDetail>(`/api/v1/players/${id}`); }
    catch (cause) { error = cause instanceof Error ? cause.message : 'Could not load player'; }
    finally { loading = false; }
  }

  async function mutate(path: string, body: unknown) {
    mutating = true;
    mutationError = '';
    try { player = await createApiClient().post<PlayerDetail>(`/api/v1/players/${id}/${path}`, { headers: { 'content-type': 'application/json' }, body: JSON.stringify(body) }); }
    catch (cause) { mutationError = cause instanceof Error ? cause.message : 'Mutation failed'; }
    finally { mutating = false; }
  }

  onMount(load);
</script>

<svelte:head><title>Player profile | Betrayal Admin</title></svelte:head>

<main class="page-shell">
  <div class="form-page">
    <a href="/players" class="back-link">← Back to roster</a>
    {#if loading}
      <p role="status" class="state">Loading player profile…</p>
    {:else if error}
      <p role="alert" class="state state-error">{error}</p>
    {:else if player}
      <header class="profile-header">
        <div><p class="eyebrow">Player profile</p><h1 class="display-title">{player.role}</h1><p class="lede">{player.alive ? 'Active in the game' : 'Eliminated from the game'} · {player.alignment}</p></div>
        <a class="btn btn-primary" href={`/players/${id}/edit`}>Edit profile</a>
      </header>
      <section class="profile-stats">
        <div><small>State</small><strong>{player.alive ? 'Alive' : 'Dead'}</strong></div>
        <div><small>Coins</small><strong>{player.coins}</strong></div>
        <div><small>Luck</small><strong>{player.luck}</strong></div>
        <div><small>Capacity</small><strong>{player.item_limit}</strong></div>
      </section>
      {#if mutationError}<p role="alert" class="state state-error">{mutationError}</p>{/if}
      <section class="profile-section">
        <div class="section-heading"><div><p class="eyebrow">Resources</p><h2>Inventory</h2></div><span>{player.items.length} items</span></div>
        <ul aria-label="Items" class="resource-list">
          {#each player.items as item (item.id)}<li><span>{item.name} <small>× {item.quantity}</small></span><button class="btn" disabled={mutating} onclick={() => mutate('items/remove', { name: item.name })}>Remove</button></li>{:else}<li class="empty-row">No items assigned.</li>{/each}
        </ul>
        <form onsubmit={(event) => { event.preventDefault(); const form = new FormData(event.currentTarget); mutate('items/add', { name: form.get('name'), quantity: Number(form.get('quantity') || 1) }); }} class="inline-form"><input name="name" aria-label="Item name" placeholder="Add an item" /><input name="quantity" type="number" min="1" value="1" aria-label="Quantity" /><button class="btn btn-primary" disabled={mutating}>Add item</button></form>
      </section>
      <section class="profile-section">
        <div class="section-heading"><div><p class="eyebrow">Game log</p><h2>Notes</h2></div><span>{player.notes.length} notes</span></div>
        <ul aria-label="Notes" class="resource-list">
          {#each player.notes as note (note.id)}<li><span>{note.info}</span><button class="btn" disabled={mutating} onclick={() => mutate('notes/remove', { note_id: note.id })}>Remove</button></li>{:else}<li class="empty-row">No notes yet.</li>{/each}
        </ul>
      </section>
    {/if}
  </div>
</main>
