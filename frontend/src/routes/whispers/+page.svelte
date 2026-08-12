<script lang="ts">
  import { onMount } from 'svelte';
  import { createApiClient } from '$lib/api/client';
  import FuzzySelect from '$lib/forms/FuzzySelect.svelte';

  type Group = { id: number; name: string; players: string[] };
  type Player = { id: string; label: string; detail?: string };
  type Message = { id: number; message: string; enabled: boolean };
  type Data = { groups: Group[]; players: Player[]; messages: Message[] };
  let data = $state<Data | null>(null);
  let error = $state<string | null>(null);
  let busy = $state(false);
  let newGroup = $state('');
  let newMessage = $state('');
  let pendingPlayer = $state<Record<number, string>>({});

  async function load() {
    try { data = await createApiClient().get<Data>('/api/v1/whisper'); error = null; }
    catch (cause) { data = null; error = cause instanceof Error ? cause.message : 'Could not load whisper settings'; }
  }
  async function createGroup(event: SubmitEvent) {
    event.preventDefault(); if (!newGroup.trim()) return; busy = true;
    try { await createApiClient().post('/api/v1/whisper/groups', { body: JSON.stringify({ name: newGroup }), headers: { 'Content-Type': 'application/json' } }); newGroup = ''; await load(); }
    catch (cause) { error = cause instanceof Error ? cause.message : 'Could not create group'; } finally { busy = false; }
  }
  async function deleteGroup(id: number) {
    if (!confirm('Delete this twin group?')) return; busy = true;
    try { await createApiClient().delete(`/api/v1/whisper/groups/${id}`); await load(); }
    catch (cause) { error = cause instanceof Error ? cause.message : 'Could not delete group'; } finally { busy = false; }
  }
  async function mutateMember(groupID: number, playerID: string, add: boolean) {
    busy = true;
    try {
      const api = createApiClient();
      await (add ? api.post(`/api/v1/whisper/groups/${groupID}/members`, { body: JSON.stringify({ player_id: playerID }), headers: { 'Content-Type': 'application/json' } }) : api.delete(`/api/v1/whisper/groups/${groupID}/members`, { body: JSON.stringify({ player_id: playerID }), headers: { 'Content-Type': 'application/json' } }));
      pendingPlayer[groupID] = '';
      await load();
    } catch (cause) { error = cause instanceof Error ? cause.message : 'Could not update group membership'; } finally { busy = false; }
  }
  async function createMessage(event: SubmitEvent) {
    event.preventDefault(); if (!newMessage.trim()) return; busy = true;
    try { await createApiClient().post('/api/v1/whisper/messages', { body: JSON.stringify({ message: newMessage }), headers: { 'Content-Type': 'application/json' } }); newMessage = ''; await load(); }
    catch (cause) { error = cause instanceof Error ? cause.message : 'Could not add doubt message'; } finally { busy = false; }
  }
  async function updateMessage(message: Message) {
    busy = true;
    try { await createApiClient().put(`/api/v1/whisper/messages/${message.id}`, { body: JSON.stringify(message), headers: { 'Content-Type': 'application/json' } }); await load(); }
    catch (cause) { error = cause instanceof Error ? cause.message : 'Could not update doubt message'; } finally { busy = false; }
  }
  async function deleteMessage(id: number) {
    if (!confirm('Remove this doubt message from the pool?')) return; busy = true;
    try { await createApiClient().delete(`/api/v1/whisper/messages/${id}`); await load(); }
    catch (cause) { error = cause instanceof Error ? cause.message : 'Could not remove doubt message'; } finally { busy = false; }
  }
  function playerOptions(group: Group) {
    return (data?.players ?? []).filter((player) => !group.players.includes(player.id)).map((player) => ({ value: player.id, label: player.label, detail: player.detail }));
  }
  function playerLabel(id: string) { return data?.players.find((player) => player.id === id)?.label ?? 'Unknown player'; }
  function playerDetail(id: string) {
    const detail = data?.players.find((player) => player.id === id)?.detail;
    return detail === 'Discord member' ? '' : detail;
  }
  onMount(load);
</script>

<svelte:head><title>Whispers | Betrayal Admin</title></svelte:head>
<main class="min-h-screen bg-[#080b12] p-6 text-slate-100 sm:p-10"><div class="mx-auto max-w-6xl">
  <header class="border-b border-slate-800 pb-8"><p class="text-xs font-semibold uppercase tracking-[0.2em] text-violet-300">Social mechanics</p><h1 class="mt-3 text-4xl font-semibold">Whispers</h1><p class="mt-3 max-w-2xl text-slate-400">Manage symmetric twin groups and the 2% doubt-message pool. A whisper sent to any member reaches every member in their group.</p></header>
  {#if error}<p role="alert" class="mt-6 rounded border border-red-500/40 bg-red-500/5 p-4 text-red-200">{error}</p>{/if}
  {#if !data && !error}<p role="status" class="py-12 text-slate-300">Loading whisper settings…</p>{:else if data}
    <section class="mt-8 rounded border border-slate-800 bg-[#0d111b] p-5"><div class="flex flex-wrap items-end justify-between gap-4"><div><p class="text-xs uppercase tracking-wide text-slate-500">Twin groups</p><h2 class="mt-2 text-2xl font-semibold">Linked players</h2></div><form class="flex w-full gap-2 sm:w-auto" onsubmit={createGroup}><input bind:value={newGroup} required maxlength="100" placeholder="Group name" aria-label="New group name" class="min-w-0 flex-1 bg-slate-900 p-2 sm:w-56" /><button disabled={busy} class="border border-violet-400 px-4 py-2">Create group</button></form></div>
      {#if data.groups.length === 0}<p class="mt-6 rounded border border-dashed border-slate-700 p-6 text-sm text-slate-400">No twin groups yet. Create one above, then choose players from the member selector.</p>{/if}
      <div class="mt-5 grid gap-4 md:grid-cols-2">{#each data.groups as group (group.id)}<article class="rounded border border-slate-700 p-4"><div class="flex justify-between gap-4"><div><h3 class="font-medium">{group.name}</h3><p class="mt-1 text-xs text-slate-500">{group.players.length} linked player{group.players.length === 1 ? '' : 's'}</p></div><button class="text-sm text-red-300" disabled={busy} onclick={() => deleteGroup(group.id)}>Delete</button></div>
        <div class="mt-4 flex flex-wrap gap-2">{#if group.players.length === 0}<span class="text-sm text-amber-200">No members assigned</span>{:else}{#each group.players as player}<span class="rounded-full border border-violet-400/50 bg-violet-400/10 px-3 py-1 text-sm text-violet-100"><span>{playerLabel(player)}</span>{#if playerDetail(player)} <span class="text-xs text-violet-300">({playerDetail(player)})</span>{/if}<button type="button" class="ml-2 text-violet-300 hover:text-white" aria-label={`Remove ${playerLabel(player)}`} onclick={() => mutateMember(group.id, player, false)}>×</button></span>{/each}{/if}</div>
        <div class="mt-5 border-t border-slate-800 pt-4"><FuzzySelect label="Add a player" placeholder="Type a name or player ID…" options={playerOptions(group)} bind:value={pendingPlayer[group.id]} disabled={busy} emptyText="Every player is already in this group" /><button type="button" class="mt-3 border border-violet-400 px-3 py-2 text-sm text-violet-200 disabled:opacity-50" disabled={busy || !pendingPlayer[group.id]} onclick={() => mutateMember(group.id, pendingPlayer[group.id], true)}>Add selected player</button></div>
      </article>{/each}</div>
    </section>
    <section class="mt-8 rounded border border-slate-800 bg-[#0d111b] p-5"><div class="flex flex-wrap items-end justify-between gap-4"><div><p class="text-xs uppercase tracking-wide text-slate-500">Suspicion pool</p><h2 class="mt-2 text-2xl font-semibold">Doubt messages</h2></div><form class="flex min-w-72 flex-1 gap-2 sm:max-w-xl" onsubmit={createMessage}><input bind:value={newMessage} required maxlength="1000" placeholder="Add a doubt message" aria-label="New doubt message" class="min-w-0 flex-1 bg-slate-900 p-2" /><button disabled={busy} class="border border-amber-400 px-4 py-2">Add message</button></form></div>
      {#if data.messages.length === 0}<p class="mt-6 rounded border border-dashed border-slate-700 p-6 text-sm text-amber-200">No doubt messages configured. Add at least one active message for the 2% suspicion effect to have anything to send.</p>{/if}
      <div class="mt-5 space-y-3">{#each data.messages as message (message.id)}<article class="flex flex-wrap items-center gap-3 rounded border border-slate-700 p-3"><input bind:value={message.message} maxlength="1000" aria-label={`Doubt message ${message.id}`} class="min-w-60 flex-1 bg-slate-900 p-2" /><label class="flex items-center gap-2 text-sm text-slate-400"><input type="checkbox" bind:checked={message.enabled} /> Enabled</label><button disabled={busy} class="border border-slate-600 px-3 py-2 text-sm" onclick={() => updateMessage(message)}>Save</button><button disabled={busy} class="text-sm text-red-300" onclick={() => deleteMessage(message.id)}>Remove</button></article>{/each}</div>
    </section>
  {/if}
</div></main>
