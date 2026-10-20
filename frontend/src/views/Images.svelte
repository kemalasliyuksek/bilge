<script lang="ts">
import type { inventory } from "../../wailsjs/go/models";
import DataTable, { type Column } from "../components/DataTable.svelte";
import Empty from "../components/Empty.svelte";
import FilterBar from "../components/FilterBar.svelte";
import Status from "../components/Status.svelte";
import { formatAge, formatBytes, shortId, shortRef } from "../lib/format";
import { scan } from "../lib/scan.svelte";
import { countBy, statusIndex } from "../lib/status";
import { matchesQuery } from "../lib/table";

type I = inventory.Image;

const tags = (i: I) => (i.tags ?? []).filter((t) => t !== "<none>:<none>");

let query = $state("");
let show = $state("all");

const status = $derived(statusIndex(scan.result?.report));
const usedBy = $derived(countBy(scan.result?.inventory.containers, (c) => [c.imageId]));
const all = $derived(scan.result?.inventory.images ?? []);
const of = (s: string) => all.filter((i) => status("image", i.id) === s).length;
const options = $derived([
  { id: "all", label: "All", count: all.length },
  { id: "in use", label: "In use", count: of("in use") },
  { id: "review", label: "Review", count: of("review") },
  { id: "safe", label: "Safe", count: of("safe") },
]);
const rows = $derived(
  all.filter(
    (i) => (show === "all" || status("image", i.id) === show) && matchesQuery(query, [...tags(i), i.id]),
  ),
);

const columns: Column<I>[] = [
  { id: "tags", label: "Tags", width: 300, sort: (i) => tags(i)[0] ?? "~" },
  { id: "id", label: "ID", width: 130, sort: (i) => i.id },
  { id: "size", label: "Size", width: 100, align: "right", sort: (i) => i.size },
  { id: "containers", label: "Containers", width: 100, align: "right", sort: (i) => usedBy.get(i.id) ?? 0 },
  { id: "created", label: "Created", width: 130, sort: (i) => i.created },
  { id: "status", label: "Status", sort: (i) => status("image", i.id) },
];
</script>

{#if all.length === 0}
  <Empty text="No images." />
{:else}
  <FilterBar bind:query bind:selected={show} {options} placeholder="Filter by tag or ID" shown={rows.length} total={all.length} />
  <DataTable id="images" {columns} {rows} key={(i) => i.id} sortBy="size" sortDir="desc">
    {#snippet row(i: I)}
      <td>
        {#each tags(i) as t}<div class="mono" title={t}>{shortRef(t)}</div>{:else}<span class="sub">untagged</span>{/each}
      </td>
      <td class="mono sub" title={i.id}>{shortId(i.id)}</td>
      <td class="num">{formatBytes(i.size)}</td>
      <td class="num">{usedBy.get(i.id) ?? 0}</td>
      <td>{formatAge(i.created)}</td>
      <td><Status value={status("image", i.id)} /></td>
    {/snippet}
  </DataTable>
{/if}
