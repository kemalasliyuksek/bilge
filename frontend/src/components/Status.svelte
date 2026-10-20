<script lang="ts">
let { value }: { value: string } = $props();
</script>

<span class="status" data-value={value}>{value}</span>

<style>
  .status {
    display: inline-block;
    padding: 0.1rem 0.5rem;
    border-radius: 999px;
    font-size: 0.75rem;
    white-space: nowrap;
    background: var(--chip);
    color: var(--muted);
  }

  .status[data-value="running"],
  .status[data-value="in use"] {
    background: var(--ok-bg);
    color: var(--ok);
  }

  .status[data-value="safe"] {
    background: var(--safe-bg);
    color: var(--safe);
  }

  .status[data-value="review"],
  .status[data-value="paused"],
  .status[data-value="restarting"] {
    background: var(--warn-bg);
    color: var(--warn);
  }

  .status[data-value="dead"] {
    background: var(--bad-bg);
    color: var(--bad);
  }
</style>
