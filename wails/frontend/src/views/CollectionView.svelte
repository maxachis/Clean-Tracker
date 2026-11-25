<script lang="ts">
  import { onMount } from "svelte";
  import { api, errorMessage } from "../lib/api";
  import { appState, showExecution } from "../lib/stores/appState.svelte";
  import { formatShortClock } from "../lib/services/countdown";
  import { averageSecondsPerAction } from "../lib/services/pace";
  import type { Draft, MoveDirection } from "../lib/types";
  import AddChoreRow from "../lib/components/AddChoreRow.svelte";
  import ChoreRow from "../lib/components/ChoreRow.svelte";

  const SECONDS_PER_MINUTE = 60;

  const draft = $derived(appState.draft as Draft);
  const chores = $derived(draft.chores);
  const totalActions = $derived(chores.reduce((sum, c) => sum + c.target_count, 0));

  let minutesText = $state(String((appState.draft?.session.length_seconds ?? 0) / SECONDS_PER_MINUTE));
  let suggestions = $state<string[]>([]);
  let error = $state("");

  const minutes = $derived(Number(minutesText));
  const average = $derived(averageSecondsPerAction({ totalActions, minutes }));
  const canStart = $derived(chores.length > 0 && Number.isInteger(minutes) && minutes > 0);

  onMount(async () => {
    suggestions = await api.recentChoreNames().catch(() => []);
  });

  async function apply(change: () => Promise<Draft>) {
    error = "";
    try {
      appState.draft = await change();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  const addChore = (chore: { name: string; count: number }) => apply(() => api.addChore(chore));
  const editChore = (id: string, edit: { name: string; count: number }) =>
    apply(() => api.updateChore({ id, ...edit }));
  const removeChore = (id: string) => apply(() => api.removeChore(id));

  function moveChore(index: number, direction: MoveDirection) {
    const ids = chores.map((c) => c.id);
    const target = direction === "up" ? index - 1 : index + 1;
    [ids[index], ids[target]] = [ids[target], ids[index]];
    return apply(() => api.reorderChores(ids));
  }

  const commitMinutes = () => apply(() => api.setDraftLength(minutes));

  async function start() {
    error = "";
    try {
      await api.startSession();
      const view = await api.getRunningSession();
      if (view) showExecution(view);
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<section class="collection">
  <div class="list-col">
    <h1>Today's list</h1>
    {#if chores.length === 0}
      <p class="empty">Nothing here yet. Add the first chore and how many times it needs doing.</p>
    {:else}
      <ol class="chores">
        {#each chores as chore, i (chore.id)}
          <ChoreRow
            {chore}
            isFirst={i === 0}
            isLast={i === chores.length - 1}
            onedit={(edit) => editChore(chore.id, edit)}
            onremove={() => removeChore(chore.id)}
            onmove={(direction) => moveChore(i, direction)}
          />
        {/each}
      </ol>
    {/if}
    <AddChoreRow {suggestions} existingNames={chores.map((c) => c.name)} onadd={addChore} />
  </div>

  <aside class="plan">
    <label class="minutes">
      <span>Session length</span>
      <span class="minutes-field">
        <input
          class="num"
          type="number"
          min="1"
          inputmode="numeric"
          bind:value={minutesText}
          onchange={commitMinutes}
          aria-label="Session length in minutes"
        />
        <span>min</span>
      </span>
    </label>

    <p class="pace num" aria-live="polite">
      <strong>{totalActions} {totalActions === 1 ? "action" : "actions"}</strong>
      {#if average !== null}
        <span> · ~{formatShortClock(average)} each</span>
      {/if}
    </p>

    {#if error}<p class="error" role="alert">{error}</p>{/if}

    <button type="button" class="btn primary start" disabled={!canStart} onclick={start}>Start</button>
  </aside>
</section>

<style>
  .collection { display: grid; grid-template-columns: 1fr 280px; gap: var(--space-5); align-items: start; }
  .list-col { display: flex; flex-direction: column; gap: var(--space-4); min-width: 0; }
  h1 { font-size: 2.2rem; font-weight: 800; font-variation-settings: "wdth" var(--wdth-wide); line-height: 1; }
  .empty { color: var(--ink-soft); font-weight: 500; }
  .chores { display: flex; flex-direction: column; gap: var(--space-2); margin: 0; padding: 0; }
  .plan {
    position: sticky; top: var(--gutter);
    display: flex; flex-direction: column; gap: var(--space-4);
    background: var(--panel); border: var(--border); border-radius: var(--radius);
    box-shadow: var(--shadow); padding: var(--space-4);
  }
  .minutes { display: flex; flex-direction: column; gap: var(--space-2); font-weight: 700; }
  .minutes-field { display: flex; align-items: baseline; gap: var(--space-2); }
  .minutes input {
    width: 100%; border: 0; border-bottom: var(--border); background: transparent;
    font-size: 3rem; font-weight: 800; font-variation-settings: "wdth" var(--wdth-wide); padding: 0;
  }
  .minutes input:focus { outline: none; border-bottom-color: var(--signal); }
  .pace { font-size: 1.05rem; line-height: 1.4; }
  .pace strong { font-weight: 800; }
  .start { font-size: 1.6rem; padding: var(--space-3) var(--space-4); font-variation-settings: "wdth" var(--wdth-wide); font-weight: 800; }

  @media (max-width: 720px) {
    .collection { grid-template-columns: 1fr; gap: var(--space-4); }
    .plan { position: static; order: -1; }
    h1 { font-size: 1.8rem; }
  }
</style>
