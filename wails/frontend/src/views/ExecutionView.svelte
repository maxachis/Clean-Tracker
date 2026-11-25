<script lang="ts">
  import { onMount } from "svelte";
  import { api, errorMessage } from "../lib/api";
  import { appState, showSummary } from "../lib/stores/appState.svelte";
  import { formatClock } from "../lib/services/countdown";
  import { projectClock } from "../lib/services/localTick";
  import { keyToCommand } from "../lib/services/keys";
  import { playMissBeep } from "../lib/services/audio";
  import { isTerminal } from "../lib/types";
  import type { SessionView, Snapshot } from "../lib/types";
  import CountdownRing from "../lib/components/CountdownRing.svelte";
  import ChoreTile from "../lib/components/ChoreTile.svelte";
  import StatLine from "../lib/components/StatLine.svelte";
  import MissFlash from "../lib/components/MissFlash.svelte";

  const TICK_INTERVAL_MS = 250;
  const URGENT_FRACTION = 0.2;
  const MAX_HOTKEYS = 9;

  const run = appState.run as SessionView;
  const sessionId = run.session_id;

  let snapshot = $state<Snapshot>(run.snapshot);
  let receivedAtMs = $state(performance.now());
  let nowMs = $state(performance.now());
  let error = $state("");
  let confirmingEnd = $state(false);
  let flashId = $state(0);
  let refreshing = false;

  const clock = $derived(projectClock({ snapshot, receivedAtMs, nowMs }));
  const urgent = $derived(!snapshot.paused && clock.ringFraction <= URGENT_FRACTION);

  onMount(() => {
    const timer = setInterval(tick, TICK_INTERVAL_MS);
    return () => clearInterval(timer);
  });

  function tick() {
    nowMs = performance.now();
    if (clock.expired) void refresh();
  }

  /** Ask Go for the truth: it knows about misses, time-up and completion. */
  async function refresh() {
    if (refreshing) return;
    refreshing = true;
    try {
      const view = await api.getRunningSession();
      if (view) accept(view.snapshot);
      else await showSummary(sessionId);
    } catch (err) {
      error = errorMessage(err);
    } finally {
      refreshing = false;
    }
  }

  function accept(next: Snapshot) {
    const missedMore = next.missed > snapshot.missed;
    snapshot = next;
    receivedAtMs = performance.now();
    nowMs = receivedAtMs;
    if (missedMore) signalMiss();
    if (isTerminal(next.status)) void showSummary(sessionId);
  }

  function signalMiss() {
    flashId++;
    if (appState.settings.miss_sound) playMissBeep();
  }

  async function command(send: () => Promise<Snapshot>) {
    error = "";
    try {
      accept(await send());
    } catch (err) {
      error = errorMessage(err);
      await refresh();
    }
  }

  const logChore = (choreId: string) => command(() => api.logAction(choreId));
  const undo = () => command(() => api.undo());
  const togglePause = () => command(() => (snapshot.paused ? api.resume() : api.pause()));
  async function endSession() {
    confirmingEnd = false;
    await command(() => api.endSession());
  }

  function onKeydown(event: KeyboardEvent) {
    const target = event.target as HTMLElement | null;
    if (target && ["INPUT", "TEXTAREA"].includes(target.tagName)) return;
    const cmd = keyToCommand(event);
    if (!cmd) return;
    event.preventDefault();
    if (cmd.kind === "undo") void undo();
    else if (cmd.kind === "togglePause") void togglePause();
    else {
      const chore = snapshot.chores[cmd.index];
      if (chore && chore.outstanding > 0) void logChore(chore.id);
    }
  }

  // A focused button would also click on Space keyup; Space is ours.
  function onKeyup(event: KeyboardEvent) {
    if (event.key === " ") event.preventDefault();
  }
</script>

<svelte:window onkeydown={onKeydown} onkeyup={onKeyup} />

{#key flashId}
  {#if flashId > 0}<MissFlash />{/if}
{/key}

<section class="run">
  <div class="dial">
    <CountdownRing
      clock={formatClock(clock.actionSecondsLeft)}
      fraction={clock.ringFraction}
      paused={snapshot.paused}
      {urgent}
    />
    <dl class="stats">
      <StatLine label="Session left" value={formatClock(clock.sessionSecondsLeft)} />
      <StatLine label="Actions left" value={snapshot.outstanding} />
      <StatLine label="Done" value={snapshot.done} />
      <StatLine label="Missed" value={snapshot.missed} alert={snapshot.missed > 0} />
    </dl>
  </div>

  <div class="work">
    <div class="tiles">
      {#each snapshot.chores as chore, i (chore.id)}
        <ChoreTile
          {chore}
          hotkey={i < MAX_HOTKEYS ? i + 1 : null}
          onlog={() => logChore(chore.id)}
        />
      {/each}
    </div>

    {#if error}<p class="error" role="alert">{error}</p>{/if}

    <div class="controls">
      <button type="button" class="btn" onclick={undo} disabled={snapshot.done === 0}>Undo</button>
      <button type="button" class="btn" onclick={togglePause}>{snapshot.paused ? "Resume" : "Pause"}</button>
      {#if confirmingEnd}
        <span class="confirm" role="alertdialog" aria-label="Confirm end session">
          End this session?
          <button type="button" class="btn danger" onclick={endSession}>End session</button>
          <button type="button" class="btn quiet" onclick={() => (confirmingEnd = false)}>Keep going</button>
        </span>
      {:else}
        <button type="button" class="btn quiet" onclick={() => (confirmingEnd = true)}>End</button>
      {/if}
    </div>
  </div>
</section>

<style>
  .run {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: var(--space-5);
    align-items: start;
  }
  .dial { display: flex; flex-direction: column; gap: var(--space-4); align-items: center; }
  .stats {
    display: grid; grid-template-columns: repeat(2, 1fr); gap: var(--space-2) var(--space-4);
    margin: 0; width: 100%;
  }
  .work { display: flex; flex-direction: column; gap: var(--space-4); min-width: 0; }
  .tiles { display: grid; grid-template-columns: repeat(auto-fill, minmax(230px, 1fr)); gap: var(--space-3); }
  .controls { display: flex; flex-wrap: wrap; gap: var(--space-3); align-items: center; }
  .confirm { display: flex; flex-wrap: wrap; gap: var(--space-2); align-items: center; font-weight: 700; }

  @media (max-width: 800px) {
    .run { grid-template-columns: 1fr; --ring-size: min(70vw, 340px); }
    .dial { align-items: center; }
  }
  @media (min-width: 801px) {
    .run { --ring-size: min(44vw, 62vh, 400px); }
  }
</style>
