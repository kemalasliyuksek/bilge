<script lang="ts">
import type { cleanup } from "../../wailsjs/go/models";
import DataTable, { type Column } from "../components/DataTable.svelte";
import Empty from "../components/Empty.svelte";
import FilterBar from "../components/FilterBar.svelte";
import RemovalDialog from "../components/RemovalDialog.svelte";
import { formatBytes, shortRef } from "../lib/format";
import { isSelected, kindLabel, resultKey, selectedByDefault } from "../lib/removal";
import { scan } from "../lib/scan.svelte";
import { matchesQuery } from "../lib/table";

type C = cleanup.Candidate;

let query = $state("");
let show = $state("all");
// Items whose selection the user flipped; see isSelected.
let toggled = $state(new Set<string>());
let pending = $state<C[] | undefined>();

const all = $derived(scan.result?.report.candidates ?? []);
const options = $derived([
  { id: "all", label: "All", count: all.length },
  ...Object.entries(kindLabel)
    .map(([id, label]) => ({ id, label, count: all.filter((c) => c.kind === id).length }))
    .filter((o) => o.count > 0),
]);
const visible = $derived(
  all.filter(
    (c) =>
      (show === "all" || c.kind === show) &&
      matchesQuery(query, [c.name, c.project, c.reason, kindLabel[c.kind]]),
  ),
);
const safe = $derived(visible.filter((c) => c.verdict === "safe"));
const review = $derived(visible.filter((c) => c.verdict === "review"));
const selected = $derived(safe.filter((c) => isSelected(c, toggled)));
const sum = (cs: C[]) => cs.reduce((t, c) => t + Math.max(c.size, 0), 0);

function toggle(c: C) {
  const k = resultKey(c);
  const next = new Set(toggled);
  if (next.has(k)) next.delete(k);
  else next.add(k);
  toggled = next;
}

function selectAll(on: boolean) {
  const next = new Set(toggled);
  for (const c of safe) {
    if (selectedByDefault(c) === on) next.delete(resultKey(c));
    else next.add(resultKey(c));
  }
  toggled = next;
}

const base: Column<C>[] = [
  { id: "kind", label: "Kind", width: 110, sort: (c) => kindLabel[c.kind] ?? c.kind },
  { id: "name", label: "Name", width: 280, sort: (c) => c.name },
  { id: "project", label: "Project", width: 140, sort: (c) => c.project },
  {
    id: "size",
    label: "Frees",
    width: 100,
    align: "right",
    title: "Space this frees. Layers shared with other images are not counted.",
    sort: (c) => c.size,
  },
];
const safeColumns: Column<C>[] = [{ id: "pick", label: "", width: 44, fixed: true }, ...base, { id: "reason", label: "Reason" }];
// Reason takes the remaining width so the Remove button never scrolls away.
const reviewColumns: Column<C>[] = [
  ...base,
  { id: "reason", label: "Reason" },
  { id: "act", label: "", width: 110, fixed: true },
];
</script>

{#snippet cells(c: C)}
  <td>{kindLabel[c.kind] ?? c.kind}</td>
  <td class="mono" title={c.name}>{shortRef(c.name)}</td>
  <td>{c.project}</td>
  <td class="num">{formatBytes(c.size)}</td>
  <td class="sub" title={c.reason}>{c.reason}</td>
{/snippet}

{#if all.length === 0}
  <Empty text="Nothing to clean up." />
{:else}
  <FilterBar
    bind:query
    bind:selected={show}
    {options}
    placeholder="Filter by name, project, reason"
    shown={visible.length}
    total={all.length}
  />

  {#if safe.length > 0}
    <section>
      <header>
        <div>
          <h2>Safe <span class="sub">{formatBytes(sum(safe))}</span></h2>
          <p class="sub">No data is lost. Selected items are removed together after one confirmation.</p>
        </div>
        <div class="bulk">
          <button type="button" onclick={() => selectAll(selected.length !== safe.length)}>
            {selected.length === safe.length ? "Select none" : "Select all"}
          </button>
          <button
            type="button"
            class="danger"
            disabled={selected.length === 0 || scan.removing}
            onclick={() => (pending = selected)}
          >
            Remove {selected.length} · {formatBytes(sum(selected))}
          </button>
        </div>
      </header>
      <DataTable id="cleanup-safe" columns={safeColumns} rows={safe} key={resultKey} sortBy="size" sortDir="desc">
        {#snippet row(c: C)}
          <td>
            <input
              type="checkbox"
              checked={isSelected(c, toggled)}
              onchange={() => toggle(c)}
              aria-label="Select {c.name}"
            />
          </td>
          {@render cells(c)}
        {/snippet}
      </DataTable>
    </section>
  {/if}

  {#if review.length > 0}
    <section>
      <h2>Review <span class="sub">{formatBytes(sum(review))}</span></h2>
      <p class="sub">
        Probably waste, but may hold data or belong to a stack stopped on purpose. Removed one at a time.
      </p>
      <DataTable id="cleanup-review" columns={reviewColumns} rows={review} key={resultKey} sortBy="size" sortDir="desc">
        {#snippet row(c: C)}
          {@render cells(c)}
          <td class="actions">
            <button type="button" aria-label="Remove {c.name}" disabled={scan.removing} onclick={() => (pending = [c])}>Remove…</button>
          </td>
        {/snippet}
      </DataTable>
    </section>
  {/if}

  {#if visible.length === 0}
    <Empty text="No candidates match the filter." />
  {/if}
{/if}

{#if pending}
  <RemovalDialog items={pending} onclose={() => (pending = undefined)} />
{/if}

<style>
  section {
    margin-top: 1.25rem;
  }

  header {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 1rem;
  }

  h2 {
    margin: 0;
    font-size: 1rem;
  }

  .bulk {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }

  .danger {
    border-color: var(--bad);
    background: var(--bad);
    color: var(--bg);
  }

  .danger:disabled {
    background: transparent;
    color: var(--muted);
    border-color: var(--line);
  }

  input[type="checkbox"] {
    accent-color: var(--accent);
  }
</style>
