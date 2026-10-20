<script lang="ts">
import { formatBytes } from "../lib/format";
import { scan } from "../lib/scan.svelte";

let { open }: { open: (view: string) => void } = $props();

const sum = (xs: { size: number }[] | null | undefined) =>
  (xs ?? []).reduce((total, x) => total + Math.max(x.size, 0), 0);

const inv = $derived(scan.result?.inventory);
const candidates = $derived(scan.result?.report.candidates ?? []);
const review = $derived(candidates.filter((c) => c.verdict === "review"));
const running = $derived((inv?.containers ?? []).filter((c) => c.state === "running").length);
</script>

{#if inv}
  <section class="tiles">
    <button type="button" class="tile" onclick={() => open("containers")}>
      <span class="label">Containers</span>
      <span class="value">{inv.containers?.length ?? 0}</span>
      <span class="sub">{running} running</span>
    </button>
    <button type="button" class="tile" onclick={() => open("images")}>
      <span class="label">Images</span>
      <span class="value">{inv.images?.length ?? 0}</span>
      <span class="sub">{formatBytes(sum(inv.images))}</span>
    </button>
    <button type="button" class="tile" onclick={() => open("volumes")}>
      <span class="label">Volumes</span>
      <span class="value">{inv.volumes?.length ?? 0}</span>
      <span class="sub">{formatBytes(sum(inv.volumes))}</span>
    </button>
    <button type="button" class="tile" onclick={() => open("networks")}>
      <span class="label">Networks</span>
      <span class="value">{inv.networks?.length ?? 0}</span>
      <span class="sub">&nbsp;</span>
    </button>
    <div class="tile">
      <span class="label">Build cache</span>
      <span class="value">{formatBytes(sum(inv.buildCache))}</span>
      <span class="sub">{inv.buildCache?.length ?? 0} records</span>
    </div>
  </section>

  <section class="reclaim">
    <h2>Reclaimable</h2>
    <p>
      <strong>{formatBytes(scan.result?.safeSize ?? 0)}</strong> can be freed safely.
      {#if review.length > 0}
        {review.length} more {review.length === 1 ? "item needs" : "items need"} review
        ({formatBytes(sum(review))}).
      {/if}
    </p>
    <button type="button" class="primary" onclick={() => open("cleanup")}>Open cleanup</button>
  </section>
{/if}

<style>
  .tiles {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(10rem, 1fr));
    gap: 0.75rem;
  }

  .tile {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    padding: 1rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
    color: inherit;
    font: inherit;
    text-align: left;
  }

  button.tile {
    cursor: pointer;
  }

  button.tile:hover {
    border-color: var(--accent);
  }

  .label {
    color: var(--muted);
    font-size: 0.8rem;
  }

  .value {
    font-size: 1.5rem;
    font-variant-numeric: tabular-nums;
  }

  .reclaim {
    margin-top: 1.5rem;
    padding: 1rem;
    border: 1px solid var(--line);
    border-radius: 8px;
  }

  .reclaim h2 {
    margin: 0 0 0.5rem;
    font-size: 1rem;
  }
</style>
