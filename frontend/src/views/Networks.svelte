<script lang="ts">
import Empty from "../components/Empty.svelte";
import Status from "../components/Status.svelte";
import { formatAge } from "../lib/format";
import { scan } from "../lib/scan.svelte";
import { countBy, statusIndex } from "../lib/status";

const builtIn = new Set(["bridge", "host", "none"]);
const status = $derived(statusIndex(scan.result?.report));
const usedBy = $derived(countBy(scan.result?.inventory.containers, (c) => c.networkIds));
const networks = $derived(
  [...(scan.result?.inventory.networks ?? [])].sort((a, b) => a.name.localeCompare(b.name)),
);
</script>

{#if networks.length === 0}
  <Empty text="No networks." />
{:else}
  <table>
    <thead>
      <tr>
        <th>Name</th>
        <th>Driver</th>
        <th>Scope</th>
        <th class="num">Containers</th>
        <th>Created</th>
        <th>Status</th>
      </tr>
    </thead>
    <tbody>
      {#each networks as n (n.id)}
        <tr>
          <td>{n.name}</td>
          <td>{n.driver}</td>
          <td>{n.scope}</td>
          <td class="num">{usedBy.get(n.id) ?? 0}</td>
          <td class="nowrap">{formatAge(n.created)}</td>
          <td>
            {#if builtIn.has(n.name)}
              <Status value="built-in" />
            {:else}
              <Status value={status("network", n.id)} />
            {/if}
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{/if}
