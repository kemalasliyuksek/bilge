<script lang="ts">
export type Option = { id: string; label: string; count: number };

let {
  query = $bindable(""),
  selected = $bindable("all"),
  options,
  placeholder = "Filter",
  shown,
  total,
}: {
  query?: string;
  selected?: string;
  options: Option[];
  placeholder?: string;
  shown: number;
  total: number;
} = $props();
</script>

<div class="bar">
  <input type="search" bind:value={query} {placeholder} aria-label={placeholder} />
  <fieldset class="chips" aria-label="Show">
    {#each options as o (o.id)}
      <button
        type="button"
        aria-pressed={selected === o.id}
        class:on={selected === o.id}
        onclick={() => (selected = o.id)}
      >
        {o.label} <span class="n">{o.count}</span>
      </button>
    {/each}
  </fieldset>
  {#if shown !== total}
    <span class="sub">{shown} of {total}</span>
  {/if}
</div>

<style>
  .bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem 0.75rem;
    margin-bottom: 0.75rem;
  }

  input {
    width: 20rem;
    padding: 0.35rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--bg);
    color: inherit;
    font: inherit;
  }

  input:focus {
    outline: 2px solid var(--accent-bg);
    border-color: var(--accent);
  }

  .chips {
    margin: 0;
    padding: 0;
    border: 0;
    display: flex;
    gap: 0.25rem;
  }

  .chips button {
    padding: 0.25rem 0.6rem;
    border-radius: 999px;
    font-size: 0.8rem;
  }

  .chips button.on {
    border-color: var(--accent);
    background: var(--accent-bg);
    color: var(--accent);
  }

  .n {
    color: var(--muted);
    font-variant-numeric: tabular-nums;
  }
</style>
