// Wire types mirroring wails/model (JSON tags are the contract).

// Mirrors model/status.go. Values arrive from Go as plain strings.
export const SessionStatus = {
  Draft: "draft",
  Active: "active",
  Paused: "paused",
  Completed: "completed",
  TimeUp: "time_up",
  Abandoned: "abandoned",
} as const;
export type SessionStatus = (typeof SessionStatus)[keyof typeof SessionStatus];

const TERMINAL_STATUSES: readonly SessionStatus[] = [
  SessionStatus.Completed,
  SessionStatus.TimeUp,
  SessionStatus.Abandoned,
];

export function isTerminal(status: SessionStatus): boolean {
  return TERMINAL_STATUSES.includes(status);
}

export interface Session {
  id: string;
  length_seconds: number;
  min_action_seconds: number;
  created_at: string;
  started_at?: string;
  ended_at?: string;
  end_status?: SessionStatus;
}

export interface Chore {
  id: string;
  session_id: string;
  name: string;
  target_count: number;
  position: number;
}

export interface ChoreProgress {
  id: string;
  name: string;
  target: number;
  done: number;
  outstanding: number;
}

export interface Snapshot {
  status: SessionStatus;
  chores: ChoreProgress[];
  done: number;
  outstanding: number;
  missed: number;
  session_seconds_left: number;
  action_seconds_left: number;
  action_seconds_total: number;
  paused: boolean;
  ended_at?: string;
  as_of: string;
}

export interface Draft {
  session: Session;
  chores: Chore[];
}

export interface SessionView {
  session_id: string;
  session: Session;
  snapshot: Snapshot;
}

export interface Summary {
  session: Session;
  snapshot: Snapshot;
}

export interface Settings {
  last_session_minutes: number;
  miss_sound: boolean;
  min_action_seconds: number;
}

export interface ChoreInput {
  name: string;
  count: number;
}

export interface ChoreEdit extends ChoreInput {
  id: string;
}

/** The IPC surface, one method per exported Go method on App. */
export interface Backend {
  getDraft(): Promise<Draft>;
  addChore(req: ChoreInput): Promise<Draft>;
  updateChore(req: ChoreEdit): Promise<Draft>;
  removeChore(id: string): Promise<Draft>;
  reorderChores(ids: string[]): Promise<Draft>;
  setDraftLength(minutes: number): Promise<Draft>;
  recentChoreNames(): Promise<string[]>;
  startSession(): Promise<Snapshot>;
  getRunningSession(): Promise<SessionView | null>;
  logAction(choreId: string): Promise<Snapshot>;
  undo(): Promise<Snapshot>;
  pause(): Promise<Snapshot>;
  resume(): Promise<Snapshot>;
  endSession(): Promise<Snapshot>;
  getSummary(sessionId: string): Promise<Summary>;
  repeatSession(sessionId: string): Promise<Draft>;
  getSettings(): Promise<Settings>;
  updateSettings(req: Settings): Promise<Settings>;
}

export type MoveDirection = "up" | "down";
