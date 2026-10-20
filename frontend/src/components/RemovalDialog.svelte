<script lang="ts">
import { onMount } from "svelte";
import type { cleanup, sweep } from "../../wailsjs/go/models";
import { formatBytes, shortRef } from "../lib/format";
import { kindLabel, needsTypedConfirm, requestFor, resultKey, summarize } from "../lib/removal";
import { scan } from "../lib/scan.svelte";

let { items, onclose }: { items: cleanup.Candidate[]; onclose: () => void } = $props();

let dialog: HTMLDialogElement;
let phase = $state<"confirm" | "running" | "done">("confirm");
let typed = $state("");
let error = $state("");

const single = $derived(items.length === 1 ? items[0] : undefined);
const mustType = $derived(single !== undefined && needsTypedConfirm(single));
const canConfirm = $derived(!mustType || typed === single?.name);
const total = $derived(items.reduce((t, c) => t + Math.max(c.size, 0), 0));
const byKey = $derived(new Map(scan.progress.map((r) => [resultKey(r), r] as const)));
const summary = $derived(summarize(scan.progress));

onMount(() => dialog.showModal());

function close() {
  if (phase === "running") return;
  dialog.close();
  onclose();
}

async function confirm() {
  if (!canConfirm) return;
  phase = "running";
  try {
    await scan.remove(items.map(requestFor));
  } catch (err) {
    error = err instanceof Error ? err.message : String(err);
  }
  phase = "done";
}

const statusText = (r: sweep.Result | undefined) => {
  if (!r) return phase === "running" ? "waiting" : "";
  return r.status;
};
</script>

<dialog
  bind:this={dialog}
  oncancel={(e) => {
    // Escape must not hide a removal that is still running.
    e.preventDefault();
    close();
  }}
>
  <h2>
    {#if phase === "confirm"}
      {single ? `Remove ${kindLabel[single.kind]?.toLowerCase() ?? single.kind}?` : `Remove ${items.length} items?`}
    {:else if phase === "running"}
      Removing…
    {:else}
      Done
    {/if}
  </h2>

  {#if phase === "confirm"}
    <p>
      This frees about <strong>{formatBytes(total)}</strong>. Removal cannot be undone.
      {#if !single}Each item is checked again against the engine right before it is removed.{/if}
    </p>
    {#if single}
      <p class="reason">{single.reason}</p>
    {/if}
  {/if}

  <ul class="items">
    {#each items as c (resultKey(c))}
      {@const r = byKey.get(resultKey(c))}
      <li>
        <span class="kind">{kindLabel[c.kind] ?? c.kind}</span>
        <span class="name mono" title={c.name}>{shortRef(c.name)}</span>
        <span class="num">{formatBytes(c.size)}</span>
        {#if phase !== "confirm"}
          <span class="state" data-status={r?.status ?? "waiting"} title={r?.message}>{statusText(r)}</span>
        {/if}
      </li>
      {#if r?.message}
        <li class="message">{r.message}</li>
      {/if}
    {/each}
  </ul>

  {#if phase === "confirm" && mustType && single}
    <label class="typed">
      <span>This is a named volume and may hold data. Type <code>{single.name}</code> to confirm.</span>
      <input bind:value={typed} autocomplete="off" spellcheck="false" />
    </label>
  {/if}

  {#if phase === "done"}
    {#if error}
      <p class="error" role="alert">{error}</p>
    {:else}
      <p>
        Removed {summary.removed}, freed {formatBytes(summary.freed)}.
        {#if summary.skipped}Skipped {summary.skipped}.{/if}
        {#if summary.failed}Failed {summary.failed}.{/if}
      </p>
    {/if}
  {/if}

  <div class="buttons">
    {#if phase === "confirm"}
      <button type="button" onclick={close}>Cancel</button>
      <button type="button" class="danger" disabled={!canConfirm} onclick={confirm}>
        {single ? "Remove" : `Remove ${items.length} items`}
      </button>
    {:else}
      <button type="button" class="primary" disabled={phase === "running"} onclick={close}>Close</button>
    {/if}
  </div>
</dialog>

<style>
  dialog {
    width: min(40rem, 90vw);
    padding: 1.25rem;
    border: 1px solid var(--line);
    border-radius: 10px;
    background: var(--bg);
    color: var(--fg);
  }

  dialog::backdrop {
    background: rgb(0 0 0 / 0.45);
  }

  h2 {
    margin: 0 0 0.75rem;
    font-size: 1.1rem;
  }

  .reason {
    color: var(--muted);
  }

  .items {
    max-height: 40vh;
    overflow: auto;
    margin: 0.75rem 0;
    padding: 0;
    list-style: none;
    border-top: 1px solid var(--line);
  }

  .items li {
    display: grid;
    grid-template-columns: 6.5rem 1fr 5.5rem 5rem;
    gap: 0.5rem;
    align-items: center;
    padding: 0.35rem 0;
    border-bottom: 1px solid var(--line);
  }

  .items li.message {
    display: block;
    padding-left: 7rem;
    color: var(--muted);
    font-size: 0.8rem;
  }

  .kind {
    color: var(--muted);
  }

  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .state {
    font-size: 0.8rem;
    color: var(--muted);
  }

  .state[data-status="removed"] {
    color: var(--ok);
  }

  .state[data-status="skipped"] {
    color: var(--warn);
  }

  .state[data-status="failed"] {
    color: var(--bad);
  }

  .typed {
    display: grid;
    gap: 0.4rem;
    margin: 0.75rem 0;
  }

  .typed input {
    padding: 0.4rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--bg);
    color: inherit;
    font: inherit;
    font-family: ui-monospace, "JetBrains Mono", Menlo, monospace;
  }

  .error {
    color: var(--bad);
  }

  .buttons {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }

  .danger {
    border-color: var(--bad);
    background: var(--bad);
    color: var(--bg);
  }
</style>
