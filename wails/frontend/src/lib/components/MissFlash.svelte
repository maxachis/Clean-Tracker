<script lang="ts">
  /** Full-window red-orange pulse; re-keyed by the parent on every miss. */
</script>

<div class="flash" data-testid="miss-flash" aria-hidden="true"></div>

<style>
  .flash {
    position: fixed; inset: 0; pointer-events: none; z-index: 10;
    background: var(--miss);
    animation: pulse var(--flash-ms) ease-out forwards;
  }
  @keyframes pulse { 0% { opacity: 0.75; } 100% { opacity: 0; } }
</style>
