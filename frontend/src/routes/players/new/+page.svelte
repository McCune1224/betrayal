<script lang="ts">
  import { onMount } from 'svelte';
  import { createApiClient } from '$lib/api/client';
  import { goto } from '$app/navigation';

  type Member = { id: string; username: string; nickname?: string; bot: boolean };
  type Role = { id: number; name: string };
  let members = $state<Member[]>([]);
  let roles = $state<Role[]>([]);
  let error = $state('');
  let saving = $state(false);
  let loading = $state(true);

  onMount(async () => {
    try {
      const api = createApiClient();
      const [discord, catalog] = await Promise.all([
        api.get<{ members: Member[] }>('/api/v1/discord/resources'),
        api.get<Role[]>('/api/v1/catalog/roles')
      ]);
      members = discord.members.filter((member) => !member.bot);
      roles = catalog;
    } catch (cause) {
      error = cause instanceof Error ? cause.message : 'Could not load Discord members and roles';
    } finally { loading = false; }
  });

  async function submit(event: SubmitEvent) {
    event.preventDefault(); saving = true; error = '';
    const form = new FormData(event.currentTarget as HTMLFormElement);
    try {
      const player = await createApiClient().post<{ id: number }>('/api/v1/players', {
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ id: String(form.get('member_id')), role: form.get('role') })
      });
      await goto(`/players/${player.id}/edit`);
    } catch (cause) { error = cause instanceof Error ? cause.message : 'Could not create player'; }
    finally { saving = false; }
  }
</script>

<svelte:head><title>New player | Betrayal Admin</title></svelte:head>
<main class="min-h-screen bg-[#080b12] p-6 text-slate-100 sm:p-10">
  <div class="mx-auto max-w-xl">
    <p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Player administration</p>
    <h1 class="mt-3 text-4xl font-semibold">Add player</h1>
    <p class="mt-3 text-slate-400">Choose a real Discord member and assign the starting role. Discord snowflake IDs stay as exact strings.</p>
    {#if error}<p role="alert" class="mt-6 rounded border border-red-500/40 p-4 text-red-200">{error}</p>{/if}
    {#if loading}<p role="status" class="mt-8 text-slate-400">Loading Discord members…</p>{:else}<form onsubmit={submit} class="mt-8 space-y-5 rounded border border-slate-800 bg-[#0d111b] p-6">
      <label class="block text-sm text-slate-300">Discord member<select name="member_id" required class="mt-2 w-full rounded bg-slate-900 px-3 py-3"><option value="">Select a member</option>{#each members as member}<option value={member.id}>{member.nickname ? `${member.nickname} · ` : ''}{member.username} ({member.id})</option>{/each}</select></label>
      <label class="block text-sm text-slate-300">Starting role<select name="role" required class="mt-2 w-full rounded bg-slate-900 px-3 py-3"><option value="">Select a role</option>{#each roles as role}<option value={role.name}>{role.name}</option>{/each}</select></label>
      {#if members.length === 0}<p class="text-sm text-amber-200">No Discord members were returned. Make sure Discord is connected and the bot can view guild members.</p>{/if}
      <button disabled={saving || members.length === 0 || roles.length === 0} class="rounded border border-teal-400 px-4 py-2 text-teal-300 disabled:opacity-50">{saving ? 'Creating…' : 'Create player'}</button>
    </form>{/if}
  </div>
</main>
