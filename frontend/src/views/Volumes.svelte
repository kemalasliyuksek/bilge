<script lang="ts">
import Empty from "../components/Empty.svelte";
import Status from "../components/Status.svelte";
import { formatAge, formatBytes } from "../lib/format";
import { scan } from "../lib/scan.svelte";
import { countBy, statusIndex } from "../lib/status";

const status = $derived(statusIndex(scan.result?.report));
const usedBy = $derived(countBy(scan.result?.inventory.containers, (c) => c.volumes));
const volumes = $derived([...(scan.result?.inventory.volumes ?? [])].sort((a, b) => b.size - a.size));
</script>

{#if volumes.length === 0}
  <Empty text="No volumes." />
{:else}
  <table>
    <thead>
      <tr>
        <th>Name</th>
        <th>Project</th>
        <th class="num">Size</th>
        <th class="num">Containers</th>
        <th>Created</th>
        <th>Status</th>
      </tr>
    </thead>
    <tbody>
      {#each volumes as v (v.name)}
        <tr>
          <td class="mono">{v.name}</td>
          <td>{v.labels?.["com.docker.compose.project"] ?? ""}</td>
          <td class="num">{formatBytes(v.size)}</td>
          <td class="num">{usedBy.get(v.name) ?? 0}</td>
          <td class="nowrap">{formatAge(v.created)}</td>
          <td><Status value={status("volume", v.name)} /></td>
        </tr>
      {/each}
    </tbody>
  </table>
{/if}
