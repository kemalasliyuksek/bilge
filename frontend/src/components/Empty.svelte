<script lang="ts">
let { text }: { text: string } = $props();
</script>

<p class="empty">{text}</p>

<style>
  .empty {
    color: var(--muted);
    padding: 2rem 0;
    text-align: center;
  }
</style>
