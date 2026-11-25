import type { ChoreProgress } from "../types";

const SECONDS_PER_MINUTE = 60;

export interface PaceOptions {
  totalActions: number;
  minutes: number;
}

/** Seconds available per action if the whole list is split evenly; null if undefined. */
export function averageSecondsPerAction(options: PaceOptions): number | null {
  const { totalActions, minutes } = options;
  if (totalActions <= 0 || minutes <= 0) return null;
  return (minutes * SECONDS_PER_MINUTE) / totalActions;
}

export function sumCounts(chores: { count: number }[]): number {
  return chores.reduce((sum, c) => sum + c.count, 0);
}

/** 0..1 completion of one chore, for summary bars. */
export function choreCompletion(chore: Pick<ChoreProgress, "done" | "target">): number {
  if (chore.target <= 0) return 0;
  return Math.min(1, chore.done / chore.target);
}
