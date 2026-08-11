<script lang="ts">
 import { createApiClient } from '$lib/api/client'; import { goto } from '$app/navigation';
 let error=$state(''); let saving=$state(false);
 async function submit(event:SubmitEvent){event.preventDefault();saving=true;error='';const f=new FormData(event.currentTarget as HTMLFormElement);try{const p=await createApiClient().post<{id:number}>('/api/v1/players',{headers:{'content-type':'application/json'},body:JSON.stringify({id:Number(f.get('id')),role:f.get('role')})});await goto(`/players/${p.id}/edit`)}catch(e){error=e instanceof Error?e.message:'Could not create player'}finally{saving=false}}
</script>
<svelte:head><title>New player | Betrayal Admin</title></svelte:head><main class="min-h-screen bg-slate-950 p-6 text-slate-100"><h1 class="mx-auto max-w-xl text-3xl">New player</h1>{#if error}<p role="alert" class="mx-auto max-w-xl py-4 text-red-300">{error}</p>{/if}<form onsubmit={submit} class="mx-auto mt-8 max-w-xl space-y-4 border border-slate-700 p-6"><label>Player ID<input name="id" type="number" min="1" required /></label><label>Role<input name="role" required /></label><button disabled={saving}>{saving?'Creating…':'Create player'}</button></form></main>
