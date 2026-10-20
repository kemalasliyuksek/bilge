<script lang="ts">
import Empty from "../components/Empty.svelte";
import { formatBytes, shortRef } from "../lib/format";
import { scan } from "../lib/scan.svelte";

const kindLabel: Record<string, string> = {
  container: "Container",
  image: "Image",
  volume: "Volume",
  network: "Network",
  buildCache: "Build cache",
};

const candidates = $derived(
  [...(scan.result?.report.candidates ?? [])].sort((a, b) => b.size - a.size),
);
const safe = $derived(candidates.filter((c) => c.verdict === "safe"));
const review = $derived(candidates.filter((c) => c.verdict === "review"));
</script>

{#if candidates.length === 0}
  <Empty text="Nothing to clean up." />
{:else}
  <p class="note">Removal from bilge is not available yet. This list shows what would be offered.</p>

  {#each [{ title: "Safe", hint: "No data is lost. Can be removed together.", items: safe }, { title: "Review", hint: "Probably waste, but may hold data or belong to a stack stopped on purpose. Removed one at a time.", items: review }] as group}
    {#if group.items.length > 0}
      <section>
        <h2>{group.title} <span class="sub">{formatBytes(group.items.reduce((t, c) => t + Math.max(c.size, 0), 0))}</span></h2>
        <p class="sub">{group.hint}</p>
        <table>
          <thead>
            <tr>
              <th>Kind</th>
              <th>Name</th>
              <th>Project</th>
              <th class="num" title="Space this frees. Layers shared with other images are not counted.">Frees</th>
              <th>Reason</th>
            </tr>
          </thead>
          <tbody>
            {#each group.items as c (c.kind + c.id)}
              <tr>
                <td>{kindLabel[c.kind] ?? c.kind}</td>
                <td class="mono" title={c.name}>{shortRef(c.name)}</td>
                <td>{c.project}</td>
                <td class="num">{formatBytes(c.size)}</td>
                <td class="sub">{c.reason}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </section>
    {/if}
  {/each}
{/if}

<style>
  .note {
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    background: var(--chip);
    color: var(--muted);
  }

  section {
    margin-top: 1.5rem;
  }

  h2 {
    margin: 0;
    font-size: 1rem;
  }
</style>
