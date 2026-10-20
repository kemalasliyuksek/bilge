<script lang="ts">
import Empty from "../components/Empty.svelte";
import Status from "../components/Status.svelte";
import { formatAge, formatBytes, shortRef } from "../lib/format";
import { scan } from "../lib/scan.svelte";

const containers = $derived(
  [...(scan.result?.inventory.containers ?? [])].sort((a, b) => a.name.localeCompare(b.name)),
);
const project = (labels: Record<string, string> | null) =>
  labels?.["com.docker.compose.project"] ?? "";
</script>

{#if containers.length === 0}
  <Empty text="No containers." />
{:else}
  <table>
    <thead>
      <tr>
        <th>Name</th>
        <th>Image</th>
        <th>State</th>
        <th class="num">Writable layer</th>
        <th>Created</th>
        <th></th>
      </tr>
    </thead>
    <tbody>
      {#each containers as c (c.id)}
        <tr>
          <td>
            <div>{c.name}</div>
            {#if project(c.labels)}<div class="sub">{project(c.labels)}</div>{/if}
          </td>
          <td class="mono" title={c.image}>{shortRef(c.image)}</td>
          <td>
            <Status value={c.state} />
            <div class="sub">{c.status}</div>
          </td>
          <td class="num">{formatBytes(c.sizeRw)}</td>
          <td class="nowrap">{formatAge(c.created)}</td>
          <td class="actions">
            {#if c.state === "running"}
              <button type="button" disabled={scan.busy !== ""} onclick={() => scan.containerAction("stop", c.id)}>Stop</button>
              <button type="button" disabled={scan.busy !== ""} onclick={() => scan.containerAction("restart", c.id)}>Restart</button>
            {:else if c.state === "exited" || c.state === "created"}
              <button type="button" disabled={scan.busy !== ""} onclick={() => scan.containerAction("start", c.id)}>Start</button>
            {/if}
            {#if scan.busy === c.id}<span class="sub">working…</span>{/if}
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{/if}
