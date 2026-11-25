<script lang="ts">
  import { onMount } from "svelte";
  import { appState, boot, Screen, setMissSound } from "./lib/stores/appState.svelte";
  import { errorMessage } from "./lib/api";
  import CollectionView from "./views/CollectionView.svelte";
  import ExecutionView from "./views/ExecutionView.svelte";
  import SummaryView from "./views/SummaryView.svelte";

  let bootError = $state("");

  onMount(() => {
    boot().catch((err) => (bootError = errorMessage(err)));
  });
</script>

<div class="shell">
  <header class="bar">
    <span class="brand">Clean Tracker</span>
    <label class="sound">
      <input
        type="checkbox"
        checked={appState.settings.miss_sound}
        onchange={(e) => setMissSound(e.currentTarget.checked)}
      />
      Miss sound
    </label>
  </header>

  <main class="stage">
    {#if bootError}
      <p class="error" role="alert">{bootError}</p>
    {:else if appState.screen === Screen.Collection}
      <CollectionView />
    {:else if appState.screen === Screen.Execution && appState.run}
      <ExecutionView />
    {:else if appState.screen === Screen.Summary && appState.summary}
      <SummaryView />
    {/if}
  </main>
</div>

<style>
  .shell {
    min-height: 100%;
    display: flex;
    flex-direction: column;
  }
  .bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-3) var(--gutter);
    border-bottom: var(--border);
    background: var(--panel);
  }
  .brand {
    font-weight: 800;
    font-variation-settings: "wdth" var(--wdth-wide);
    font-size: 1.1rem;
    letter-spacing: 0.01em;
  }
  .sound {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    font-size: 0.9rem;
    font-weight: 600;
    cursor: pointer;
  }
  .sound input { width: 18px; height: 18px; accent-color: var(--ink); }
  .stage {
    flex: 1;
    padding: var(--gutter);
    width: 100%;
    max-width: 880px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
  }
</style>
