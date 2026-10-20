<script lang="ts">
import type { inventory } from "../../wailsjs/go/models";
import DataTable, { type Column } from "../components/DataTable.svelte";
import Empty from "../components/Empty.svelte";
import FilterBar from "../components/FilterBar.svelte";
import Status from "../components/Status.svelte";
import { formatAge, formatBytes } from "../lib/format";
import { composeLabel } from "../lib/overview";
import { scan } from "../lib/scan.svelte";
import { countBy, statusIndex } from "../lib/status";
import { matchesQuery } from "../lib/table";

type V = inventory.Volume;

const project = (v: V) => v.labels?.[composeLabel] ?? "";

let query = $state("");
let show = $state("all");

const status = $derived(statusIndex(scan.result?.report));
const usedBy = $derived(countBy(scan.result?.inventory.containers, (c) => c.volumes));
const all = $derived(scan.result?.inventory.volumes ?? []);
const of = (s: string) => all.filter((v) => status("volume", v.name) === s).length;
const options = $derived([
  { id: "all", label: "All", count: all.length },
  { id: "in use", label: "In use", count: of("in use") },
  { id: "review", label: "Review", count: of("review") },
  { id: "safe", label: "Safe", count: of("safe") },
]);
const rows = $derived(
  all.filter(
    (v) => (show === "all" || status("volume", v.name) === show) && matchesQuery(query, [v.name, project(v), v.driver]),
  ),
);

const columns: Column<V>[] = [
  { id: "name", label: "Name", width: 300, sort: (v) => v.name },
  { id: "project", label: "Project", width: 160, sort: (v) => project(v) },
  { id: "size", label: "Size", width: 100, align: "right", sort: (v) => v.size },
  { id: "containers", label: "Containers", width: 100, align: "right", sort: (v) => usedBy.get(v.name) ?? 0 },
  { id: "created", label: "Created", width: 130, sort: (v) => v.created },
  { id: "status", label: "Status", sort: (v) => status("volume", v.name) },
];
</script>

{#if all.length === 0}
  <Empty text="No volumes." />
{:else}
  <FilterBar bind:query bind:selected={show} {options} placeholder="Filter by name or project" shown={rows.length} total={all.length} />
  <DataTable id="volumes" {columns} {rows} key={(v) => v.name} sortBy="size" sortDir="desc">
    {#snippet row(v: V)}
      <td class="mono" title={v.name}>{v.name}</td>
      <td>{project(v)}</td>
      <td class="num">{formatBytes(v.size)}</td>
      <td class="num">{usedBy.get(v.name) ?? 0}</td>
      <td>{formatAge(v.created)}</td>
      <td><Status value={status("volume", v.name)} /></td>
    {/snippet}
  </DataTable>
{/if}
