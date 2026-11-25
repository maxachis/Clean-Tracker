<script lang="ts">
  interface Props {
    clock: string;
    fraction: number;
    paused: boolean;
    urgent: boolean;
  }
  let { clock, fraction, paused, urgent }: Props = $props();

  const RADIUS = 90;
  const CIRCUMFERENCE = 2 * Math.PI * RADIUS;
  const offset = $derived(CIRCUMFERENCE * (1 - fraction));
</script>

<div class="ring" class:paused class:urgent>
  <svg viewBox="0 0 200 200" aria-hidden="true">
    <circle class="track" cx="100" cy="100" r={RADIUS} />
    <circle
      class="arc"
      cx="100" cy="100" r={RADIUS}
      stroke-dasharray={CIRCUMFERENCE}
      stroke-dashoffset={offset}
      transform="rotate(-90 100 100)"
    />
  </svg>
  <div class="face">
    <span class="clock num" role="timer" aria-label="Time until the next action is due">{clock}</span>
    <span class="caption">{paused ? "Paused" : "until next action"}</span>
  </div>
</div>

<style>
  .ring {
    position: relative;
    width: var(--ring-size);
    aspect-ratio: 1;
    container-type: inline-size;
    flex: none;
  }
  svg { width: 100%; height: 100%; display: block; }
  circle { fill: none; stroke-width: var(--ring-stroke); }
  .track { stroke: var(--done-wash); }
  .arc {
    stroke: var(--signal);
    stroke-linecap: butt;
    transition: stroke-dashoffset 260ms linear, stroke 200ms;
  }
  .urgent .arc { stroke: var(--miss); }
  .paused .arc { stroke: var(--ink-soft); }
  .face {
    position: absolute; inset: 0;
    display: flex; flex-direction: column; align-items: center; justify-content: center;
  }
  .clock {
    font-size: 23cqw;
    font-weight: 800;
    font-variation-settings: "wdth" var(--wdth-wide);
    line-height: 1;
    letter-spacing: -0.02em;
    color: var(--signal);
  }
  .urgent .clock { color: var(--miss); }
  .paused .clock { color: var(--ink-soft); animation: blink 1.4s steps(2, jump-none) infinite; }
  .caption { font-weight: 700; font-size: 5cqw; color: var(--ink-soft); margin-top: 2cqw; }
  @keyframes blink { 50% { opacity: 0.45; } }
</style>
