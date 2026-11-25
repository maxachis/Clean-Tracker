import type { SessionStatus, Summary } from "../types";
import { SessionStatus as Status } from "../types";

const MS_PER_SECOND = 1000;

const ENDING_LABELS: Partial<Record<SessionStatus, string>> = {
  [Status.Completed]: "Completed",
  [Status.TimeUp]: "Time's up",
  [Status.Abandoned]: "Ended early",
};

export function endingLabel(status: SessionStatus): string {
  return ENDING_LABELS[status] ?? "Finished";
}

/** Seconds from start to end; falls back to the session length if either is missing. */
export function sessionDuration(summary: Summary): number {
  const { session, snapshot } = summary;
  if (!session.started_at || !snapshot.ended_at) return session.length_seconds;
  const ms = Date.parse(snapshot.ended_at) - Date.parse(session.started_at);
  return Math.max(0, ms / MS_PER_SECOND);
}
