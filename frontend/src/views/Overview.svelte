<script lang="ts">
import DataTable, { type Column } from "../components/DataTable.svelte";
import Status from "../components/Status.svelte";
import UsageBars from "../components/UsageBars.svelte";
import { formatBytes, shortRef } from "../lib/format";
import { diskBreakdown, type Item, largest, type Project, projects } from "../lib/overview";
import { scan } from "../lib/scan.svelte";

let { open }: { open: (view: string) => void } = $props();

const inv = $derived(scan.result?.inventory);
const eng = $derived(inv?.engine);
const breakdown = $derived(diskBreakdown(scan.result));
const onDisk = $derived(breakdown.reduce((t, b) => t + b.kept + b.review + b.safe, 0));
const reviewSize = $derived(breakdown.reduce((t, b) => t + b.review, 0));
const reviewCount = $derived((scan.result?.report.candidates ?? []).filter((c) => c.verdict === "review").length);
const containers = $derived(inv?.containers ?? []);
const running = $derived(containers.filter((c) => c.state === "running").length);
const projectRows = $derived(projects(scan.result));
const top = $derived(largest(scan.result, 5));

const projectColumns: Column<Project>[] = [
  { id: "name", label: "Project", width: 200, sort: (p) => p.name },
  { id: "containers", label: "Containers", width: 120, align: "right", sort: (p) => p.containers },
  { id: "volumes", label: "Volumes", width: 150, align: "right", sort: (p) => p.volumeSize },
  { id: "networks", label: "Networks", width: 100, align: "right", sort: (p) => p.networks },
  { id: "reclaimable", label: "Reclaimable", sort: (p) => p.reclaimable },
];

const topColumns: Column<Item>[] = [
  { id: "kind", label: "Kind", width: 90 },
  { id: "name", label: "Name", width: 320 },
  { id: "size", label: "Size", width: 100, align: "right" },
  { id: "status", label: "Status" },
];
</script>

{#if inv && eng}
  <section class="hero">
    <div>
      <div class="label">Docker data on disk</div>
      <div class="big">{formatBytes(onDisk)}</div>
    </div>
    <div>
      <div class="label">Safe to remove</div>
      <div class="big safe-text">{formatBytes(scan.result?.safeSize ?? 0)}</div>
    </div>
    <div>
      <div class="label">Needs review</div>
      <div class="big">{formatBytes(reviewSize)}</div>
      <div class="sub">{reviewCount} {reviewCount === 1 ? "item" : "items"}</div>
    </div>
    <div class="cta">
      <button type="button" class="primary" onclick={() => open("cleanup")}>Open cleanup</button>
    </div>
  </section>

  <section class="tiles">
    <button type="button" class="tile" onclick={() => open("containers")}>
      <span class="label">Containers</span>
      <span class="value">{containers.length}</span>
      <span class="sub">{running} running · {containers.length - running} not running</span>
    </button>
    <button type="button" class="tile" onclick={() => open("images")}>
      <span class="label">Images</span>
      <span class="value">{inv.images?.length ?? 0}</span>
      <span class="sub">{formatBytes(breakdown[0].kept + breakdown[0].review + breakdown[0].safe)}</span>
    </button>
    <button type="button" class="tile" onclick={() => open("volumes")}>
      <span class="label">Volumes</span>
      <span class="value">{inv.volumes?.length ?? 0}</span>
      <span class="sub">{formatBytes(breakdown[2].kept + breakdown[2].review + breakdown[2].safe)}</span>
    </button>
    <button type="button" class="tile" onclick={() => open("networks")}>
      <span class="label">Networks</span>
      <span class="value">{inv.networks?.length ?? 0}</span>
      <span class="sub">{projectRows.length} Compose {projectRows.length === 1 ? "project" : "projects"}</span>
    </button>
  </section>

  <div class="grid">
    <section class="panel">
      <h2>Disk usage</h2>
      <UsageBars rows={breakdown} />
    </section>

    <section class="panel">
      <h2>Engine</h2>
      <dl>
        <dt>Docker Engine</dt><dd>{eng.version} <span class="sub">API {eng.apiVersion}</span></dd>
        <dt>Host</dt><dd>{eng.os}</dd>
        <dt>Platform</dt><dd>{eng.osType}/{eng.architecture}</dd>
        <dt>Kernel</dt><dd class="mono">{eng.kernelVersion}</dd>
        <dt>Storage driver</dt><dd>{eng.storageDriver}</dd>
        <dt>Data root</dt><dd class="mono">{eng.rootDir}</dd>
        <dt>Resources</dt><dd>{eng.cpus} CPUs · {formatBytes(eng.memoryBytes)} memory</dd>
      </dl>
    </section>
  </div>

  {#if projectRows.length > 0}
    <section class="block">
      <h2>Compose projects</h2>
      <DataTable id="overview-projects" columns={projectColumns} rows={projectRows} key={(p) => p.name} sortBy="name">
        {#snippet row(p: Project)}
          <td>{p.name}</td>
          <td class="num">{p.running}/{p.containers} running</td>
          <td class="num">{p.volumes} · {formatBytes(p.volumeSize)}</td>
          <td class="num">{p.networks}</td>
          <td>
            {#if p.reclaimable > 0}
              {formatBytes(p.reclaimable)}
            {:else}
              <span class="sub">nothing</span>
            {/if}
            {#if p.containers === 0}<span class="sub"> · no containers left</span>{/if}
          </td>
        {/snippet}
      </DataTable>
    </section>
  {/if}

  <section class="block">
    <h2>Largest images and volumes</h2>
    <DataTable id="overview-largest" columns={topColumns} rows={top} key={(i) => i.kind + i.name}>
      {#snippet row(i: Item)}
        <td>{i.kind}</td>
        <td class="mono" title={i.name}>{shortRef(i.name)}</td>
        <td class="num">{formatBytes(i.size)}</td>
        <td><Status value={i.status} /></td>
      {/snippet}
    </DataTable>
  </section>
{/if}

<style>
  h2 {
    margin: 0 0 0.75rem;
    font-size: 0.95rem;
  }

  .label {
    color: var(--muted);
    font-size: 0.8rem;
  }

  .hero {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    gap: 1rem 2.5rem;
    margin-bottom: 1.25rem;
  }

  .big {
    font-size: 2rem;
    font-weight: 600;
  }

  .safe-text {
    color: var(--safe);
  }

  .cta {
    margin-left: auto;
  }

  .tiles {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(11rem, 1fr));
    gap: 0.75rem;
    margin-bottom: 1rem;
  }

  .tile {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    padding: 0.85rem 1rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
    text-align: left;
  }

  .tile:hover {
    border-color: var(--accent);
  }

  .value {
    font-size: 1.4rem;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(22rem, 1fr));
    gap: 1rem;
  }

  .panel {
    padding: 1rem;
    border: 1px solid var(--line);
    border-radius: 8px;
  }

  dl {
    display: grid;
    grid-template-columns: max-content 1fr;
    gap: 0.35rem 1rem;
    margin: 0;
  }

  dt {
    color: var(--muted);
  }

  dd {
    margin: 0;
    overflow-wrap: anywhere;
  }

  .block {
    margin-top: 1.5rem;
  }
</style>
