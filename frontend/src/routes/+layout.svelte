<script lang="ts">
  import '../app.css';
  import AuthShell from '$lib/auth/AuthShell.svelte';

  let { children } = $props();
</script>

<AuthShell>
  {@render children()}
</AuthShell>
