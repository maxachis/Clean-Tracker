import type { Snapshot } from "../types";
import { isTerminal } from "../types";

const MS_PER_SECOND = 1000;

export interface ProjectedClock {
  actionSecondsLeft: number;
  sessionSecondsLeft: number;
  /** 1 = fresh countdown, 0 = out of time. Drives the progress ring. */
  ringFraction: number;
  /** The local countdown (or session) hit zero: ask Go for a fresh snapshot. */
  expired: boolean;
}

export interface ProjectOptions {
  snapshot: Snapshot;
  /** performance.now() when the snapshot arrived. */
  receivedAtMs: number;
  /** performance.now() now. Never compare against Go's wall clock. */
  nowMs: number;
}

/**
 * Advance a snapshot's countdowns by the time elapsed since it arrived,
 * measured on the JS monotonic clock. A paused or finished snapshot is frozen.
 */
export function projectClock(options: ProjectOptions): ProjectedClock {
  const { snapshot, receivedAtMs, nowMs } = options;
  const frozen = snapshot.paused || isTerminal(snapshot.status);
  const elapsed = frozen ? 0 : Math.max(0, (nowMs - receivedAtMs) / MS_PER_SECOND);
  const actionSecondsLeft = Math.max(0, snapshot.action_seconds_left - elapsed);
  const sessionSecondsLeft = Math.max(0, snapshot.session_seconds_left - elapsed);
  return {
    actionSecondsLeft,
    sessionSecondsLeft,
    ringFraction: ringFraction(actionSecondsLeft, snapshot.action_seconds_total),
    expired: !frozen && (actionSecondsLeft <= 0 || sessionSecondsLeft <= 0),
  };
}

function ringFraction(left: number, total: number): number {
  if (total <= 0) return 0;
  return Math.min(1, Math.max(0, left / total));
}
