<script lang="ts">
import { onMount } from "svelte";
import { Version } from "../wailsjs/go/main/App.js";
import icon from "./assets/icon.png";
import { formatAge } from "./lib/format";
import { scan } from "./lib/scan.svelte";
import Cleanup from "./views/Cleanup.svelte";
import Containers from "./views/Containers.svelte";
import Images from "./views/Images.svelte";
import Networks from "./views/Networks.svelte";
import Overview from "./views/Overview.svelte";
import Volumes from "./views/Volumes.svelte";

const views = [
  { id: "overview", title: "Overview" },
  { id: "containers", title: "Containers" },
  { id: "images", title: "Images" },
  { id: "volumes", title: "Volumes" },
  { id: "networks", title: "Networks" },
  { id: "cleanup", title: "Cleanup" },
];

let current = $state("overview");
let version = $state("");
// Re-render the "scanned … ago" line every 30 s without rescanning.
let now = $state(new Date());

const count = (id: string): number | undefined => {
  const inv = scan.result?.inventory;
  switch (id) {
    case "containers":
      return inv?.containers?.length;
    case "images":
      return inv?.images?.length;
    case "volumes":
      return inv?.volumes?.length;
    case "networks":
      return inv?.networks?.length;
    case "cleanup":
      return scan.result?.report.candidates?.length;
  }
};

onMount(() => {
  Version().then((v) => {
    version = v;
  });
  scan.refresh();
  const timer = setInterval(() => {
    now = new Date();
  }, 30_000);
  return () => clearInterval(timer);
});
</script>

<div class="shell">
  <nav>
    <div class="brand"><img src={icon} alt="" width="24" height="24" />bilge</div>
    {#each views as v (v.id)}
      <button type="button" class:active={current === v.id} onclick={() => (current = v.id)}>
        <span>{v.title}</span>
        {#if count(v.id) !== undefined}<span class="count">{count(v.id)}</span>{/if}
      </button>
    {/each}
    <div class="version">{version}</div>
  </nav>

  <main>
    <header>
      <h1>{views.find((v) => v.id === current)?.title}</h1>
      <div class="scan">
        {#if scan.result}
          <span class="sub">scanned {formatAge(scan.result.inventory.takenAt, now)}</span>
        {/if}
        <button type="button" onclick={() => scan.refresh()} disabled={scan.loading}>
          {scan.loading ? "Scanning…" : "Refresh"}
        </button>
      </div>
    </header>

    {#if scan.error}
      <div class="error" role="alert">
        <span>{scan.error}</span>
        <button type="button" onclick={() => (scan.error = "")}>Dismiss</button>
      </div>
    {/if}

    {#if !scan.result}
      <p class="sub">{scan.loading ? "Reading the Docker Engine…" : ""}</p>
    {:else if current === "overview"}
      <Overview open={(id) => (current = id)} />
    {:else if current === "containers"}
      <Containers />
    {:else if current === "images"}
      <Images />
    {:else if current === "volumes"}
      <Volumes />
    {:else if current === "networks"}
      <Networks />
    {:else if current === "cleanup"}
      <Cleanup />
    {/if}
  </main>
</div>

<style>
  .shell {
    display: grid;
    grid-template-columns: 12rem 1fr;
    height: 100vh;
  }

  nav {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
    padding: 1rem 0.5rem;
    border-right: 1px solid var(--line);
    background: var(--panel);
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0 0.75rem 1rem;
    font-weight: 600;
    font-size: 1.1rem;
  }

  nav button {
    display: flex;
    justify-content: space-between;
    padding: 0.45rem 0.75rem;
    border: 0;
    border-radius: 6px;
    background: transparent;
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  nav button:hover {
    background: var(--chip);
  }

  nav button.active {
    background: var(--accent-bg);
    color: var(--accent);
  }

  .count {
    color: var(--muted);
    font-variant-numeric: tabular-nums;
  }

  .version {
    margin-top: auto;
    padding: 0 0.75rem;
    color: var(--muted);
    font-size: 0.75rem;
  }

  main {
    overflow: auto;
    padding: 1rem 1.5rem 2rem;
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1rem;
  }

  h1 {
    margin: 0;
    font-size: 1.25rem;
  }

  .scan {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .error {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 1rem;
    padding: 0.6rem 0.75rem;
    border-radius: 6px;
    background: var(--bad-bg);
    color: var(--bad);
  }
</style>
