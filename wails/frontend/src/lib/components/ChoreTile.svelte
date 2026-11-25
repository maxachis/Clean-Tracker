<script lang="ts">
  import type { ChoreProgress } from "../types";

  interface Props {
    chore: ChoreProgress;
    hotkey: number | null;
    onlog: () => void;
  }
  let { chore, hotkey, onlog }: Props = $props();

  const finished = $derived(chore.outstanding === 0);
</script>

<button type="button" class="tile" class:finished disabled={finished} onclick={onlog} data-testid="tile">
  <span class="count num" aria-label="{chore.outstanding} left">{finished ? "✓" : chore.outstanding}</span>
  <span class="name">{chore.name}</span>
  {#if hotkey !== null && !finished}<kbd>{hotkey}</kbd>{/if}
</button>

<style>
  .tile {
    position: relative;
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-height: 84px;
    padding: var(--space-3) var(--space-4);
    border: var(--border);
    border-radius: var(--radius);
    background: var(--panel);
    box-shadow: var(--shadow);
    text-align: left;
    transition: transform 80ms, box-shadow 80ms, background 120ms;
  }
  .tile:hover:not(:disabled) { transform: translate(-1px, -1px); background: var(--signal-wash); }
  .tile:active:not(:disabled) { transform: translate(3px, 3px); box-shadow: var(--shadow-pressed); }
  .count {
    font-size: 2.8rem;
    font-weight: 800;
    font-variation-settings: "wdth" var(--wdth-wide);
    min-width: 2ch;
    text-align: center;
    line-height: 1;
  }
  .name { font-weight: 700; font-size: 1.25rem; flex: 1; min-width: 0; overflow-wrap: anywhere; }
  kbd {
    font: inherit; font-weight: 800; font-size: 0.8rem;
    border: 2px solid var(--ink); border-radius: 4px; padding: 0 6px;
  }
  .finished {
    background: var(--done-wash);
    box-shadow: none;
    color: var(--ink-soft);
    opacity: 0.85;
  }
  .finished .name { text-decoration: line-through; text-decoration-thickness: 2px; }
</style>
