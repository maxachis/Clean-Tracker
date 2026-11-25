<script lang="ts">
  interface Props {
    label: string;
    value: string | number;
    alert?: boolean;
  }
  let { label, value, alert = false }: Props = $props();
</script>

<div class="stat" class:alert>
  <dt>{label}</dt>
  <dd class="num" data-testid="stat-{label.toLowerCase().replace(/\s+/g, '-')}">{value}</dd>
</div>

<style>
  .stat { display: flex; flex-direction: column; gap: 2px; padding: var(--space-2) var(--space-3); border-left: var(--border); }
  dt { font-size: 0.85rem; font-weight: 600; color: var(--ink-soft); }
  dd { margin: 0; font-size: 1.7rem; font-weight: 800; font-variation-settings: "wdth" var(--wdth-wide); line-height: 1.1; }
  .alert dd { color: var(--miss); animation: bump 500ms ease-out; }
  @keyframes bump { 0% { transform: scale(1.5); } 100% { transform: scale(1); } }
</style>
