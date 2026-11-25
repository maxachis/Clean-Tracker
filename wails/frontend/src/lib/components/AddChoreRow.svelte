<script lang="ts">
  import CountStepper from "./CountStepper.svelte";

  interface Props {
    suggestions: string[];
    existingNames: string[];
    onadd: (chore: { name: string; count: number }) => Promise<void>;
  }
  let { suggestions, existingNames, onadd }: Props = $props();

  const DEFAULT_COUNT = 1;
  const MAX_CHIPS = 6;

  let name = $state("");
  let count = $state(DEFAULT_COUNT);
  let nameInput: HTMLInputElement | undefined = $state();

  const chips = $derived(
    suggestions
      .filter((s) => !existingNames.includes(s))
      .filter((s) => s.toLowerCase().includes(name.trim().toLowerCase()))
      .slice(0, MAX_CHIPS),
  );

  async function add(chosenName: string) {
    await onadd({ name: chosenName, count });
    name = "";
    count = DEFAULT_COUNT;
    nameInput?.focus();
  }

  function submit(event: Event) {
    event.preventDefault();
    return add(name);
  }
</script>

<form class="add" onsubmit={submit}>
  <div class="line">
    <input
      bind:this={nameInput}
      class="name"
      aria-label="New chore name"
      placeholder="Add a chore, e.g. Pick up boxes"
      autocomplete="off"
      bind:value={name}
    />
    <CountStepper value={count} label="count for new chore" onchange={(n) => (count = n)} />
    <button type="submit" class="btn">Add</button>
  </div>
  {#if chips.length > 0}
    <ul class="chips" aria-label="Recent chores">
      {#each chips as chip (chip)}
        <li><button type="button" class="chip" onclick={() => add(chip)}>{chip}</button></li>
      {/each}
    </ul>
  {/if}
</form>

<style>
  .add { display: flex; flex-direction: column; gap: var(--space-3); }
  .line { display: flex; gap: var(--space-3); align-items: stretch; flex-wrap: wrap; }
  .name {
    flex: 1 1 200px;
    min-width: 0;
    border: var(--border);
    border-style: dashed;
    border-radius: var(--radius);
    background: transparent;
    padding: var(--space-2) var(--space-3);
    font-weight: 600;
    font-size: 1.05rem;
  }
  .name:focus { border-style: solid; background: var(--panel); outline: none; }
  .chips { display: flex; flex-wrap: wrap; gap: var(--space-2); margin: 0; padding: 0; list-style: none; }
  .chip {
    border: 2px solid var(--ink);
    background: var(--signal-wash);
    border-radius: 999px;
    padding: var(--space-1) var(--space-3);
    font-weight: 600;
    font-size: 0.9rem;
  }
  .chip:hover { background: var(--signal); color: var(--signal-ink); }
</style>
