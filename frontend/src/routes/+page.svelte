<script lang="ts">
  let status = $state('frontend wired');
</script>

<svelte:head>
  <title>Betrayal Admin</title>
  <meta name="description" content="Betrayal Discord administration frontend" />
</svelte:head>

<main class="grid min-h-screen place-items-center bg-slate-950 p-6 text-slate-100">
  <p class="text-sm font-medium tracking-wide">{status}</p>
</main>
