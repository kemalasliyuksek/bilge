<script lang="ts">
import type { inventory } from "../../wailsjs/go/models";
import DataTable, { type Column } from "../components/DataTable.svelte";
import Empty from "../components/Empty.svelte";
import FilterBar from "../components/FilterBar.svelte";
import Status from "../components/Status.svelte";
import { formatAge, formatBytes, shortRef } from "../lib/format";
import { composeLabel } from "../lib/overview";
import { scan } from "../lib/scan.svelte";
import { matchesQuery } from "../lib/table";

type C = inventory.Container;

const stopped = new Set(["exited", "created", "dead"]);
const project = (c: C) => c.labels?.[composeLabel] ?? "";

let query = $state("");
let show = $state("all");

const all = $derived(scan.result?.inventory.containers ?? []);
const groups: Record<string, (c: C) => boolean> = {
  all: () => true,
  running: (c) => c.state === "running",
  stopped: (c) => stopped.has(c.state),
};
const options = $derived([
  { id: "all", label: "All", count: all.length },
  { id: "running", label: "Running", count: all.filter(groups.running).length },
  { id: "stopped", label: "Stopped", count: all.filter(groups.stopped).length },
]);
const rows = $derived(
  all.filter((c) => groups[show](c) && matchesQuery(query, [c.name, c.image, project(c), c.state, c.status])),
);

const columns: Column<C>[] = [
  { id: "name", label: "Name", width: 200, sort: (c) => c.name },
  { id: "image", label: "Image", width: 240, sort: (c) => shortRef(c.image) },
  { id: "state", label: "State", width: 160, sort: (c) => c.state },
  { id: "size", label: "Writable layer", width: 110, align: "right", sort: (c) => c.sizeRw },
  { id: "created", label: "Created", width: 140, sort: (c) => c.created },
  { id: "actions", label: "" },
];
</script>

{#if all.length === 0}
  <Empty text="No containers." />
{:else}
  <FilterBar bind:query bind:selected={show} {options} placeholder="Filter by name, image, project" shown={rows.length} total={all.length} />
  <DataTable id="containers" {columns} {rows} key={(c) => c.id} sortBy="name">
    {#snippet row(c: C)}
      <td title={c.name}>
        <div>{c.name}</div>
        {#if project(c)}<div class="sub">{project(c)}</div>{/if}
      </td>
      <td class="mono" title={c.image}>{shortRef(c.image)}</td>
      <td>
        <Status value={c.state} />
        <div class="sub" title={c.status}>{c.status}</div>
      </td>
      <td class="num">{formatBytes(c.sizeRw)}</td>
      <td>{formatAge(c.created)}</td>
      <td class="actions">
        {#if scan.busy === c.id}<span class="sub">working…</span>{/if}
        {#if c.state === "running"}
          <button type="button" disabled={scan.busy !== ""} onclick={() => scan.containerAction("stop", c.id)}>Stop</button>
          <button type="button" disabled={scan.busy !== ""} onclick={() => scan.containerAction("restart", c.id)}>Restart</button>
        {:else if c.state === "exited" || c.state === "created"}
          <button type="button" disabled={scan.busy !== ""} onclick={() => scan.containerAction("start", c.id)}>Start</button>
        {/if}
      </td>
    {/snippet}
  </DataTable>
{/if}
