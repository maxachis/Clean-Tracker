<script lang="ts">
  interface Props {
    value: number;
    label: string;
    onchange: (next: number) => void;
  }
  let { value, label, onchange }: Props = $props();

  const MIN_COUNT = 1;
  const clamp = (n: number) => Math.max(MIN_COUNT, Math.floor(n) || MIN_COUNT);
</script>

<div class="stepper" role="group" aria-label={label}>
  <button type="button" class="step" aria-label="Fewer {label}" disabled={value <= MIN_COUNT} onclick={() => onchange(clamp(value - 1))}>−</button>
  <input
    class="num"
    type="number"
    inputmode="numeric"
    min={MIN_COUNT}
    aria-label={label}
    {value}
    onchange={(e) => onchange(clamp(e.currentTarget.valueAsNumber))}
  />
  <button type="button" class="step" aria-label="More {label}" onclick={() => onchange(clamp(value + 1))}>+</button>
</div>

<style>
  .stepper {
    display: inline-flex;
    border: var(--border);
    border-radius: var(--radius);
    background: var(--panel);
    overflow: hidden;
  }
  .step {
    width: 40px;
    border: 0;
    background: var(--ink);
    color: var(--paper);
    font-size: 1.3rem;
    font-weight: 800;
    line-height: 1;
  }
  .step:disabled { opacity: 0.35; }
  input {
    width: 3.2rem;
    border: 0;
    text-align: center;
    font-weight: 800;
    font-size: 1.1rem;
    background: transparent;
    -moz-appearance: textfield;
    appearance: textfield;
  }
  input::-webkit-outer-spin-button,
  input::-webkit-inner-spin-button { -webkit-appearance: none; margin: 0; }
</style>
