<script lang="ts">
import { Version } from "../wailsjs/go/main/App.js";

let version = $state("");

Version().then((v) => {
  version = v;
});
</script>

<main>
  <h1>bilge</h1>
  <p>A careful cleaner for Docker.</p>
  {#if version}
    <p class="muted">version {version}</p>
  {/if}
</main>

<style>
  main {
    padding: 2rem;
  }

  .muted {
    color: var(--muted);
  }
</style>
