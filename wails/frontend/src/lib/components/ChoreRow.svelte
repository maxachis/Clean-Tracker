<script lang="ts">
  import type { Chore, MoveDirection } from "../types";
  import CountStepper from "./CountStepper.svelte";

  interface Props {
    chore: Chore;
    isFirst: boolean;
    isLast: boolean;
    onedit: (edit: { name: string; count: number }) => void;
    onremove: () => void;
    onmove: (direction: MoveDirection) => void;
  }
  let { chore, isFirst, isLast, onedit, onremove, onmove }: Props = $props();

  function commitName(input: HTMLInputElement) {
    const name = input.value.trim();
    if (name === chore.name) return;
    onedit({ name, count: chore.target_count });
  }
</script>

<li class="row">
  <div class="moves">
    <button type="button" class="mv" aria-label="Move {chore.name} up" disabled={isFirst} onclick={() => onmove("up")}>▲</button>
    <button type="button" class="mv" aria-label="Move {chore.name} down" disabled={isLast} onclick={() => onmove("down")}>▼</button>
  </div>
  <input
    class="name"
    aria-label="Chore name"
    value={chore.name}
    onblur={(e) => commitName(e.currentTarget)}
    onkeydown={(e) => e.key === "Enter" && e.currentTarget.blur()}
  />
  <CountStepper
    value={chore.target_count}
    label="count for {chore.name}"
    onchange={(count) => onedit({ name: chore.name, count })}
  />
  <button type="button" class="rm" aria-label="Remove {chore.name}" onclick={onremove}>×</button>
</li>

<style>
  .row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
    background: var(--panel);
    border: var(--border);
    border-radius: var(--radius);
    list-style: none;
  }
  .moves { display: flex; flex-direction: column; gap: 2px; }
  .mv {
    border: 0; background: var(--done-wash); border-radius: 3px;
    font-size: 0.6rem; padding: 2px 6px; line-height: 1;
  }
  .mv:disabled { opacity: 0.3; }
  .name {
    flex: 1;
    min-width: 0;
    border: 0;
    border-bottom: 2px solid transparent;
    background: transparent;
    font-weight: 700;
    font-size: 1.1rem;
    padding: var(--space-1) 0;
  }
  .name:hover, .name:focus { border-bottom-color: var(--ink); outline: none; }
  .rm {
    border: 0; background: transparent; font-size: 1.6rem; font-weight: 700;
    width: 36px; height: 36px; border-radius: var(--radius); line-height: 1;
  }
  .rm:hover { background: var(--miss); }
</style>
