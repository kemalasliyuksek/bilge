<script lang="ts">
import type { inventory } from "../../wailsjs/go/models";
import DataTable, { type Column } from "../components/DataTable.svelte";
import Empty from "../components/Empty.svelte";
import FilterBar from "../components/FilterBar.svelte";
import Status from "../components/Status.svelte";
import { formatAge } from "../lib/format";
import { scan } from "../lib/scan.svelte";
import { countBy, statusIndex } from "../lib/status";
import { matchesQuery } from "../lib/table";

type N = inventory.Network;

const builtIn = new Set(["bridge", "host", "none"]);

let query = $state("");
let show = $state("all");

const verdict = $derived(statusIndex(scan.result?.report));
const status = (n: N) => (builtIn.has(n.name) ? "built-in" : verdict("network", n.id));
const usedBy = $derived(countBy(scan.result?.inventory.containers, (c) => c.networkIds));
const all = $derived(scan.result?.inventory.networks ?? []);
const of = (s: string) => all.filter((n) => status(n) === s).length;
const options = $derived([
  { id: "all", label: "All", count: all.length },
  { id: "in use", label: "In use", count: of("in use") },
  { id: "safe", label: "Unused", count: of("safe") },
  { id: "built-in", label: "Built-in", count: of("built-in") },
]);
const rows = $derived(
  all.filter((n) => (show === "all" || status(n) === show) && matchesQuery(query, [n.name, n.driver, n.scope])),
);

const columns: Column<N>[] = [
  { id: "name", label: "Name", width: 260, sort: (n) => n.name },
  { id: "driver", label: "Driver", width: 110, sort: (n) => n.driver },
  { id: "scope", label: "Scope", width: 90, sort: (n) => n.scope },
  { id: "containers", label: "Containers", width: 100, align: "right", sort: (n) => usedBy.get(n.id) ?? 0 },
  { id: "created", label: "Created", width: 130, sort: (n) => n.created },
  { id: "status", label: "Status", sort: (n) => status(n) },
];
</script>

{#if all.length === 0}
  <Empty text="No networks." />
{:else}
  <FilterBar bind:query bind:selected={show} {options} placeholder="Filter by name or driver" shown={rows.length} total={all.length} />
  <DataTable id="networks" {columns} {rows} key={(n) => n.id} sortBy="name">
    {#snippet row(n: N)}
      <td title={n.name}>{n.name}</td>
      <td>{n.driver}</td>
      <td>{n.scope}</td>
      <td class="num">{usedBy.get(n.id) ?? 0}</td>
      <td>{formatAge(n.created)}</td>
      <td><Status value={status(n)} /></td>
    {/snippet}
  </DataTable>
{/if}
