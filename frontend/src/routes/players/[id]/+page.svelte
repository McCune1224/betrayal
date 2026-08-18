<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { createApiClient } from '$lib/api/client';
  import type { PlayerDetail } from '$lib/player-types';

  type CatalogItem = { id: number; name: string; description: string; rarity: string; cost: number };
  type Member = { id: string; username: string; nickname?: string; bot: boolean };
  let player = $state<PlayerDetail | null>(null);
  let catalogItems = $state<CatalogItem[]>([]);
  let error = $state('');
  let loading = $state(true);
  let mutating = $state(false);
  let mutationError = $state('');
  let playerLabel = $state('');
  let confirmPhrase = $state('');
  const id = $derived(page.params.id);

  async function load() {
    loading = true;
    error = '';
    try {
      const api = createApiClient();
      const [detail, items, resources] = await Promise.all([
        api.get<PlayerDetail>(`/api/v1/players/${id}`),
        api.get<CatalogItem[]>('/api/v1/catalog/items').catch(() => [] as CatalogItem[]),
        api.get<{ members: Member[] }>('/api/v1/discord/resources').catch(() => ({ members: [] }))
      ]);
      player = detail;
      catalogItems = items;
      const member = resources.members.find((candidate) => candidate.id === detail.id);
      playerLabel = member?.nickname || member?.username || `Player ${detail.id}`;
    } catch (cause) {
      error = cause instanceof Error ? cause.message : 'Could not load player';
    } finally {
      loading = false;
    }
  }

  async function mutate(path: string, body: unknown) {
    mutating = true;
    mutationError = '';
    try {
      player = await createApiClient().post<PlayerDetail>(`/api/v1/players/${id}/${path}`, {
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify(body)
      });
    } catch (cause) {
      mutationError = cause instanceof Error ? cause.message : 'Mutation failed';
    } finally {
      mutating = false;
    }
  }

  function submitItem(event: SubmitEvent) {
    event.preventDefault();
    const form = new FormData(event.currentTarget as HTMLFormElement);
    const name = String(form.get('name') ?? '');
    if (name) void mutate('items/add', { name, quantity: Number(form.get('quantity') || 1) });
  }

  function submitNote(event: SubmitEvent) {
    event.preventDefault();
    const form = new FormData(event.currentTarget as HTMLFormElement);
    const info = String(form.get('info') ?? '').trim();
    const position = Number(form.get('position') || 1);
    if (info) void mutate('notes/add', { info, position });
    (event.currentTarget as HTMLFormElement).reset();
  }

  async function removePlayer() {
    if (confirmPhrase !== playerLabel) return;
    mutating = true;
    mutationError = '';
    try {
      await createApiClient().delete(`/api/v1/players/${id}`);
      await goto('/players');
    } catch (cause) {
      mutationError = cause instanceof Error ? cause.message : 'Could not remove player';
      mutating = false;
    }
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
        <div><p class="eyebrow">Player profile</p><h1 class="display-title">{playerLabel}</h1><p class="lede">{player.role} · {player.alive ? 'Active in the game' : 'Eliminated from the game'} · {player.alignment}</p></div>
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
        <form onsubmit={submitItem} class="inline-form">
          <select name="name" aria-label="Item to add" required>
            <option value="">Select an item…</option>
            {#each catalogItems as item (item.id)}<option value={item.name}>{item.name}</option>{/each}
          </select>
          <input name="quantity" type="number" min="1" value="1" aria-label="Quantity" />
          <button class="btn btn-primary" disabled={mutating}>Add item</button>
        </form>
      </section>
      <section class="profile-section">
        <div class="section-heading"><div><p class="eyebrow">Game log</p><h2>Notes</h2></div><span>{player.notes.length} notes</span></div>
        <ul aria-label="Notes" class="resource-list">
          {#each player.notes as note (note.id)}<li><span>{note.info}</span><button class="btn" disabled={mutating} onclick={() => mutate('notes/remove', { note_id: note.id })}>Remove</button></li>{:else}<li class="empty-row">No notes yet.</li>{/each}
        </ul>
        <form onsubmit={submitNote} class="inline-form">
          <input name="info" aria-label="Note text" placeholder="Add a note" required />
          <input name="position" type="number" min="1" value="1" aria-label="Note position" />
          <button class="btn btn-primary" disabled={mutating}>Add note</button>
        </form>
      </section>
      <section class="profile-section profile-section-danger" aria-label="Remove player">
        <div class="section-heading"><div><p class="eyebrow">Danger zone</p><h2>Remove player</h2></div></div>
        <p class="danger-copy">Permanently removes {playerLabel} from the roster. Inventory, confessional, notes, votes, and whisper membership are deleted. The pinned lifeboard must be set again afterwards.</p>
        <div class="inline-form">
          <input aria-label="Type player label to confirm" bind:value={confirmPhrase} placeholder={`Type ${playerLabel} to confirm`} />
          <button type="button" class="btn btn-danger" disabled={mutating || confirmPhrase !== playerLabel} onclick={removePlayer}>Remove player</button>
        </div>
      </section>
    {/if}
  </div>
</main>
