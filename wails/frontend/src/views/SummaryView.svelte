<script lang="ts">
  import { appState, repeatFromSummary, showCollection } from "../lib/stores/appState.svelte";
  import { errorMessage } from "../lib/api";
  import { formatClock } from "../lib/services/countdown";
  import { choreCompletion } from "../lib/services/pace";
  import { endingLabel, sessionDuration } from "../lib/services/summaryText";
  import type { Summary } from "../lib/types";

  const summary = $derived(appState.summary as Summary);
  const snapshot = $derived(summary.snapshot);
  const total = $derived(snapshot.chores.reduce((sum, c) => sum + c.target, 0));
  let error = $state("");

  async function guarded(action: () => Promise<void>) {
    error = "";
    try {
      await action();
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<section class="summary">
  <h1 data-testid="ending">{endingLabel(snapshot.status)}</h1>

  <dl class="facts">
    <div><dt>Took</dt><dd class="num">{formatClock(sessionDuration(summary))}</dd></div>
    <div><dt>Done</dt><dd class="num">{snapshot.done} of {total}</dd></div>
    <div><dt>Missed</dt><dd class="num">{snapshot.missed}</dd></div>
  </dl>

  <ul class="bars">
    {#each snapshot.chores as chore (chore.id)}
      <li>
        <div class="bar-head"><span>{chore.name}</span><span class="num">{chore.done}/{chore.target}</span></div>
        <div class="track"><div class="fill" style:width="{choreCompletion(chore) * 100}%"></div></div>
      </li>
    {/each}
  </ul>

  {#if error}<p class="error" role="alert">{error}</p>{/if}

  <div class="actions">
    <button type="button" class="btn primary" onclick={() => guarded(() => repeatFromSummary(summary.session.id))}>Repeat this list</button>
    <button type="button" class="btn" onclick={() => guarded(showCollection)}>New list</button>
  </div>
</section>

<style>
  .summary { display: flex; flex-direction: column; gap: var(--space-5); max-width: 640px; margin: 0 auto; width: 100%; }
  h1 { font-size: clamp(2.4rem, 8vw, 4rem); font-weight: 800; font-variation-settings: "wdth" var(--wdth-wide); line-height: 1; }
  .facts { display: flex; gap: var(--space-5); margin: 0; flex-wrap: wrap; }
  .facts div { border-left: var(--border); padding-left: var(--space-3); }
  dt { font-size: 0.9rem; font-weight: 600; color: var(--ink-soft); }
  dd { margin: 0; font-size: 2.2rem; font-weight: 800; font-variation-settings: "wdth" var(--wdth-wide); }
  .bars { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: var(--space-3); }
  .bar-head { display: flex; justify-content: space-between; font-weight: 700; margin-bottom: var(--space-1); }
  .track { height: 18px; border: var(--border); border-radius: var(--radius); background: var(--panel); overflow: hidden; }
  .fill { height: 100%; background: var(--signal); }
  .actions { display: flex; gap: var(--space-3); flex-wrap: wrap; }
</style>
