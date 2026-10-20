<script lang="ts">
import { formatBytes } from "../lib/format";
import type { Breakdown } from "../lib/overview";

let { rows }: { rows: Breakdown[] } = $props();

const series = [
  { id: "kept", label: "Kept" },
  { id: "review", label: "Needs review" },
  { id: "safe", label: "Safe to remove" },
] as const;

type SeriesId = (typeof series)[number]["id"];

const total = (r: Breakdown) => r.kept + r.review + r.safe;
// One shared scale so bars compare across categories.
const max = $derived(Math.max(1, ...rows.map(total)));

let tip = $state<{ x: number; y: number; text: string } | undefined>();

function show(e: PointerEvent, r: Breakdown, s: (typeof series)[number]) {
  const box = (e.currentTarget as HTMLElement).closest(".chart")?.getBoundingClientRect();
  if (!box) return;
  tip = { x: e.clientX - box.left, y: e.clientY - box.top, text: `${r.label} · ${s.label}: ${formatBytes(r[s.id as SeriesId])}` };
}
</script>

<div class="chart">
  <div class="legend">
    {#each series as s (s.id)}
      <span><i class="swatch {s.id}"></i>{s.label}</span>
    {/each}
  </div>

  {#each rows as r (r.id)}
    <div class="row">
      <span class="label">{r.label}</span>
      <div class="track">
        {#if total(r) > 0}
          <div class="bar" style:width="{(total(r) / max) * 100}%">
            {#each series as s (s.id)}
              {#if r[s.id as SeriesId] > 0}
                <span
                  class="seg {s.id}"
                  style:flex-grow={r[s.id as SeriesId]}
                  role="img"
                  aria-label="{r.label}, {s.label}: {formatBytes(r[s.id as SeriesId])}"
                  onpointermove={(e) => show(e, r, s)}
                  onpointerleave={() => (tip = undefined)}
                ></span>
              {/if}
            {/each}
          </div>
        {/if}
        <span class="value">{formatBytes(total(r))}</span>
      </div>
    </div>
  {/each}

  {#if tip}
    <div class="tip" style:left="{tip.x}px" style:top="{tip.y}px">{tip.text}</div>
  {/if}
</div>

<style>
  .chart {
    position: relative;
  }

  .legend {
    display: flex;
    gap: 1rem;
    margin-bottom: 0.75rem;
    color: var(--muted);
    font-size: 0.8rem;
  }

  .legend span {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
  }

  .swatch {
    width: 10px;
    height: 10px;
    border-radius: 3px;
  }

  .row {
    display: grid;
    grid-template-columns: 8.5rem 1fr;
    align-items: center;
    gap: 0.75rem;
    padding: 0.3rem 0;
  }

  .label {
    color: var(--muted);
    font-size: 0.85rem;
  }

  .track {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    /* Leave room for the value label at the bar's tip. */
    padding-right: 4.5rem;
  }

  .bar {
    display: flex;
    gap: 2px;
    height: 20px;
    min-width: 4px;
  }

  .seg {
    min-width: 2px;
    flex-basis: 0;
  }

  /* Square at the baseline, rounded at the data end. */
  .seg:last-child {
    border-radius: 0 4px 4px 0;
  }

  .value {
    font-size: 0.85rem;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  .kept {
    background: var(--series-kept);
  }

  .review {
    background: var(--series-review);
  }

  .safe {
    background: var(--series-safe);
  }

  .tip {
    position: absolute;
    transform: translate(12px, -130%);
    padding: 0.3rem 0.5rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--bg);
    font-size: 0.8rem;
    white-space: nowrap;
    pointer-events: none;
    box-shadow: 0 2px 8px rgb(0 0 0 / 0.15);
  }
</style>
