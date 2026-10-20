<script lang="ts">
import type { cleanup } from "../../wailsjs/go/models";
import DataTable, { type Column } from "../components/DataTable.svelte";
import Empty from "../components/Empty.svelte";
import FilterBar from "../components/FilterBar.svelte";
import { formatBytes, shortRef } from "../lib/format";
import { scan } from "../lib/scan.svelte";
import { matchesQuery } from "../lib/table";

type C = cleanup.Candidate;

const kindLabel: Record<string, string> = {
  container: "Container",
  image: "Image",
  volume: "Volume",
  network: "Network",
  buildCache: "Build cache",
};

let query = $state("");
let show = $state("all");

const all = $derived(scan.result?.report.candidates ?? []);
const options = $derived([
  { id: "all", label: "All", count: all.length },
  ...Object.entries(kindLabel)
    .map(([id, label]) => ({ id, label, count: all.filter((c) => c.kind === id).length }))
    .filter((o) => o.count > 0),
]);
const visible = $derived(
  all.filter((c) => (show === "all" || c.kind === show) && matchesQuery(query, [c.name, c.project, c.reason, kindLabel[c.kind]])),
);
const sum = (cs: C[]) => cs.reduce((t, c) => t + Math.max(c.size, 0), 0);
const groups = $derived([
  {
    id: "safe",
    title: "Safe",
    hint: "No data is lost. Can be removed together.",
    items: visible.filter((c) => c.verdict === "safe"),
  },
  {
    id: "review",
    title: "Review",
    hint: "Probably waste, but may hold data or belong to a stack stopped on purpose. Removed one at a time.",
    items: visible.filter((c) => c.verdict === "review"),
  },
]);

const columns: Column<C>[] = [
  { id: "kind", label: "Kind", width: 110, sort: (c) => kindLabel[c.kind] ?? c.kind },
  { id: "name", label: "Name", width: 280, sort: (c) => c.name },
  { id: "project", label: "Project", width: 150, sort: (c) => c.project },
  {
    id: "size",
    label: "Frees",
    width: 100,
    align: "right",
    title: "Space this frees. Layers shared with other images are not counted.",
    sort: (c) => c.size,
  },
  { id: "reason", label: "Reason" },
];
</script>

{#if all.length === 0}
  <Empty text="Nothing to clean up." />
{:else}
  <p class="note">Removal from bilge is not available yet. This list shows what would be offered.</p>
  <FilterBar bind:query bind:selected={show} {options} placeholder="Filter by name, project, reason" shown={visible.length} total={all.length} />

  {#each groups as group (group.id)}
    {#if group.items.length > 0}
      <section>
        <h2>{group.title} <span class="sub">{formatBytes(sum(group.items))}</span></h2>
        <p class="sub">{group.hint}</p>
        <DataTable id="cleanup-{group.id}" {columns} rows={group.items} key={(c) => c.kind + c.id} sortBy="size" sortDir="desc">
          {#snippet row(c: C)}
            <td>{kindLabel[c.kind] ?? c.kind}</td>
            <td class="mono" title={c.name}>{shortRef(c.name)}</td>
            <td>{c.project}</td>
            <td class="num">{formatBytes(c.size)}</td>
            <td class="sub wrap">{c.reason}</td>
          {/snippet}
        </DataTable>
      </section>
    {/if}
  {/each}
  {#if visible.length === 0}
    <Empty text="No candidates match the filter." />
  {/if}
{/if}

<style>
  .note {
    margin-top: 0;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    background: var(--chip);
    color: var(--muted);
  }

  section {
    margin-top: 1.25rem;
  }

  h2 {
    margin: 0;
    font-size: 1rem;
  }
</style>
