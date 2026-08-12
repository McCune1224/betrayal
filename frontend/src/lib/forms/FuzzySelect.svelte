<script lang="ts">
  type Option = { value: string; label: string; detail?: string };
  let {
    label,
    placeholder = 'Search…',
    options,
    value = $bindable(''),
    disabled = false,
    emptyText = 'No matches'
  }: { label: string; placeholder?: string; options: Option[]; value?: string; disabled?: boolean; emptyText?: string } = $props();

  let query = $state('');
  let open = $state(false);
  let input: HTMLInputElement;
  let selected = $derived(options.find((option) => option.value === value));
  let filtered = $derived(options.filter((option) => {
    const haystack = `${option.label} ${option.detail ?? ''} ${option.value}`.toLowerCase();
    return !query.trim() || haystack.includes(query.trim().toLowerCase());
  }).slice(0, 100));

  function choose(option: Option) {
    value = option.value;
    query = '';
    open = false;
  }
  function clear() {
    value = '';
    query = '';
    open = false;
  }

  function handleDocumentClick(event: MouseEvent) {
    if (input && !input.parentElement?.parentElement?.contains(event.target as Node)) open = false;
  }
</script>

<svelte:window onclick={handleDocumentClick} />

<div class="relative">
  <label class="block text-sm text-slate-300">{label}
    <div class="relative mt-2">
      <input bind:this={input} value={open ? query : selected?.label ?? ''} placeholder={selected ? selected.label : placeholder} disabled={disabled} autocomplete="off" class="w-full rounded bg-slate-900 px-3 py-3 pr-20" onfocus={() => open = true} oninput={(event) => { query = event.currentTarget.value; open = true; }} onkeydown={(event) => { if (event.key === 'Escape') open = false; if (event.key === 'Enter' && filtered[0]) { event.preventDefault(); choose(filtered[0]); } }} />
      {#if value}<button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 px-2 text-xs text-slate-400 hover:text-white" onclick={clear}>Clear</button>{/if}
    </div>
  </label>
  {#if open && !disabled}
    <div class="absolute z-20 mt-1 max-h-64 w-full overflow-y-auto rounded border border-slate-700 bg-[#0d111b] p-1 shadow-xl" role="listbox">
      {#if filtered.length === 0}<p class="p-3 text-sm text-slate-500">{emptyText}</p>{/if}
      {#each filtered as option (option.value)}
        <button type="button" role="option" aria-selected={option.value === value} class="block w-full rounded px-3 py-2 text-left hover:bg-slate-800" onclick={() => choose(option)}>
          <span class="block text-sm text-slate-100">{option.label}</span>
          {#if option.detail}<span class="block font-mono text-xs text-slate-500">{option.detail}</span>{/if}
        </button>
      {/each}
    </div>
  {/if}
</div>
