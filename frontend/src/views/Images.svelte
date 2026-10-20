<script lang="ts">
import Empty from "../components/Empty.svelte";
import Status from "../components/Status.svelte";
import { formatAge, formatBytes, shortId, shortRef } from "../lib/format";
import { scan } from "../lib/scan.svelte";
import { countBy, statusIndex } from "../lib/status";

const status = $derived(statusIndex(scan.result?.report));
const usedBy = $derived(countBy(scan.result?.inventory.containers, (c) => [c.imageId]));
const images = $derived([...(scan.result?.inventory.images ?? [])].sort((a, b) => b.size - a.size));
const tags = (t: string[] | null) => (t ?? []).filter((x) => x !== "<none>:<none>");
</script>

{#if images.length === 0}
  <Empty text="No images." />
{:else}
  <table>
    <thead>
      <tr>
        <th>Tags</th>
        <th>ID</th>
        <th class="num">Size</th>
        <th class="num">Containers</th>
        <th>Created</th>
        <th>Status</th>
      </tr>
    </thead>
    <tbody>
      {#each images as i (i.id)}
        <tr>
          <td>
            {#each tags(i.tags) as t}<div class="mono" title={t}>{shortRef(t)}</div>{:else}<span class="sub">untagged</span>{/each}
          </td>
          <td class="mono sub nowrap" title={i.id}>{shortId(i.id)}</td>
          <td class="num">{formatBytes(i.size)}</td>
          <td class="num">{usedBy.get(i.id) ?? 0}</td>
          <td class="nowrap">{formatAge(i.created)}</td>
          <td><Status value={status("image", i.id)} /></td>
        </tr>
      {/each}
    </tbody>
  </table>
{/if}
