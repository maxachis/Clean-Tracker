// DEV/TEST ONLY. An in-memory stand-in for the Go backend, used when the app
// runs in a plain browser (vite dev) or under vitest. It mirrors the service's
// validation messages and carries a minimal port of the replay timer
// (SPEC.md "Timer Model") so countdown and miss flow behave realistically.
// The real timer lives in Go; do not grow this file.
import type {
  Backend, Chore, ChoreEdit, ChoreInput, ChoreProgress, Draft, Session, SessionView,
  Settings, Snapshot, Summary,
} from "./types";
import { SessionStatus } from "./types";

const MS_PER_SECOND = 1000;
const SECONDS_PER_MINUTE = 60;
const DEFAULT_MIN_ACTION_SECONDS = 10;
const MIN_MIN_ACTION_SECONDS = 1;
const MAX_MIN_ACTION_SECONDS = 600;
const DEFAULT_MINUTES = 30;
const SEED_RECENT_NAMES = [
  "Wipe counters", "Pick up boxes", "Dishes", "Vacuum living room", "Take out trash", "Fold laundry",
];

const ERR = {
  choreName: "chore name must not be empty",
  choreCount: "chore count must be a whole number of 1 or more",
  length: "session length must be above zero minutes",
  minAction: "minimum per action must be a whole number of seconds from 1 to 600",
  noChores: "add at least one chore before starting",
  running: "another session is already running",
  noRunning: "no session is running",
  ended: "the session has already ended",
  unknownChore: "chore is not part of the running session",
  nothingOutstanding: "nothing is outstanding for that chore",
  nothingToUndo: "there is no logged action to undo",
  alreadyPaused: "the session is already paused",
  notPaused: "the session is not paused",
  notFound: "not found",
} as const;

type EventKind = "done" | "undo" | "pause" | "resume" | "end";
interface FakeEvent { kind: EventKind; choreId?: string; atMs: number }
interface FakeSession {
  session: Session;
  chores: Chore[];
  events: FakeEvent[];
  startedMs?: number;
  settled: boolean;
}

export interface FakeOptions {
  /** Epoch milliseconds. Defaults to Date.now. */
  now?: () => number;
  recentNames?: string[];
}

export function createFakeBackend(options: FakeOptions = {}): Backend {
  const now = options.now ?? (() => Date.now());
  const sessions = new Map<string, FakeSession>();
  let settings: Settings = { last_session_minutes: DEFAULT_MINUTES, miss_sound: true, min_action_seconds: DEFAULT_MIN_ACTION_SECONDS };
  let draftId = "";
  let nextId = 1;
  const seedNames = options.recentNames ?? SEED_RECENT_NAMES;
  const id = (prefix: string) => `${prefix}-${nextId++}`;
  const iso = (ms: number) => new Date(ms).toISOString();

  function newDraft(): FakeSession {
    const sessionId = id("s");
    const fake: FakeSession = {
      session: {
        id: sessionId,
        length_seconds: settings.last_session_minutes * SECONDS_PER_MINUTE,
        min_action_seconds: settings.min_action_seconds,
        created_at: iso(now()),
      },
      chores: [], events: [], settled: false,
    };
    sessions.set(sessionId, fake);
    draftId = sessionId;
    return fake;
  }

  const draft = (): FakeSession => sessions.get(draftId) ?? newDraft();
  const draftView = (): Draft => {
    const d = draft();
    return { session: { ...d.session }, chores: d.chores.map((c) => ({ ...c })) };
  };
  const reposition = (d: FakeSession) => d.chores.forEach((c, i) => (c.position = i));

  function validateChore(input: ChoreInput): string {
    const name = input.name.trim();
    if (!name) throw new Error(ERR.choreName);
    if (!Number.isInteger(input.count) || input.count < 1) throw new Error(ERR.choreCount);
    return name;
  }

  function liveSession(): FakeSession | undefined {
    return [...sessions.values()].find((s) => s.startedMs !== undefined && !s.settled);
  }

  function settle(fake: FakeSession, snap: Snapshot): void {
    if (!snap.status || fake.settled) return;
    if (snap.status === SessionStatus.Completed || snap.status === SessionStatus.TimeUp || snap.status === SessionStatus.Abandoned) {
      fake.settled = true;
      fake.session.end_status = snap.status;
      fake.session.ended_at = snap.ended_at;
    }
  }

  function snapshotOf(fake: FakeSession): Snapshot {
    return replay(fake, now());
  }

  function runCommand(kind: EventKind, check: (s: Snapshot) => void, choreId?: string): Snapshot {
    const fake = liveSession();
    if (!fake) throw new Error(ERR.noRunning);
    const before = snapshotOf(fake);
    settle(fake, before);
    if (before.ended_at) throw new Error(ERR.ended);
    check(before);
    fake.events.push({ kind, choreId, atMs: now() });
    const after = snapshotOf(fake);
    settle(fake, after);
    return after;
  }

  const backend: Backend = {
    async getDraft() { return draftView(); },
    async addChore(req) {
      const name = validateChore(req);
      const d = draft();
      d.chores.push({ id: id("c"), session_id: d.session.id, name, target_count: req.count, position: d.chores.length });
      return draftView();
    },
    async updateChore(req: ChoreEdit) {
      const name = validateChore(req);
      const chore = draft().chores.find((c) => c.id === req.id);
      if (!chore) throw new Error(ERR.notFound);
      chore.name = name;
      chore.target_count = req.count;
      return draftView();
    },
    async removeChore(choreId) {
      const d = draft();
      d.chores = d.chores.filter((c) => c.id !== choreId);
      reposition(d);
      return draftView();
    },
    async reorderChores(ids) {
      const d = draft();
      const byId = new Map(d.chores.map((c) => [c.id, c]));
      if (ids.length !== d.chores.length || ids.some((i) => !byId.has(i))) {
        throw new Error("reorder must list every draft chore exactly once");
      }
      d.chores = ids.map((i) => byId.get(i)!);
      reposition(d);
      return draftView();
    },
    async setDraftLength(minutes) {
      if (!Number.isInteger(minutes) || minutes < 1) throw new Error(ERR.length);
      draft().session.length_seconds = minutes * SECONDS_PER_MINUTE;
      return draftView();
    },
    async recentChoreNames() {
      const past = [...sessions.values()].filter((s) => s.startedMs !== undefined).reverse()
        .flatMap((s) => s.chores.map((c) => c.name));
      return [...new Set([...past, ...seedNames])];
    },
    async startSession() {
      const d = draft();
      if (d.chores.length === 0) throw new Error(ERR.noChores);
      if (d.session.length_seconds <= 0) throw new Error(ERR.length);
      const live = liveSession();
      if (live) {
        const snap = snapshotOf(live);
        settle(live, snap);
        if (!live.settled) throw new Error(ERR.running);
      }
      d.session.min_action_seconds = settings.min_action_seconds;
      d.startedMs = now();
      d.session.started_at = iso(d.startedMs);
      settings = { ...settings, last_session_minutes: d.session.length_seconds / SECONDS_PER_MINUTE };
      newDraft();
      return snapshotOf(d);
    },
    async getRunningSession() {
      const fake = liveSession();
      if (!fake) return null;
      const snapshot = snapshotOf(fake);
      settle(fake, snapshot);
      const view: SessionView = { session_id: fake.session.id, session: { ...fake.session }, snapshot };
      return view;
    },
    async logAction(choreId) {
      return runCommand("done", (s) => {
        const chore = s.chores.find((c) => c.id === choreId);
        if (!chore) throw new Error(ERR.unknownChore);
        if (chore.outstanding === 0) throw new Error(ERR.nothingOutstanding);
      }, choreId);
    },
    async undo() {
      return runCommand("undo", (s) => { if (s.done === 0) throw new Error(ERR.nothingToUndo); });
    },
    async pause() {
      return runCommand("pause", (s) => { if (s.paused) throw new Error(ERR.alreadyPaused); });
    },
    async resume() {
      return runCommand("resume", (s) => { if (!s.paused) throw new Error(ERR.notPaused); });
    },
    async endSession() { return runCommand("end", () => undefined); },
    async getSummary(sessionId): Promise<Summary> {
      const fake = sessions.get(sessionId);
      if (!fake || fake.startedMs === undefined) throw new Error(ERR.notFound);
      const snapshot = snapshotOf(fake);
      settle(fake, snapshot);
      return { session: { ...fake.session }, snapshot };
    },
    async repeatSession(sessionId) {
      const source = sessions.get(sessionId);
      if (!source) throw new Error(ERR.notFound);
      const d = draft();
      d.session.length_seconds = source.session.length_seconds;
      d.chores = source.chores.map((c, i) => ({ ...c, id: id("c"), session_id: d.session.id, position: i }));
      return draftView();
    },
    async getSettings() { return { ...settings }; },
    async updateSettings(req) {
      if (req.last_session_minutes < 1) throw new Error(ERR.length);
      const m = req.min_action_seconds;
      if (!Number.isInteger(m) || m < MIN_MIN_ACTION_SECONDS || m > MAX_MIN_ACTION_SECONDS) throw new Error(ERR.minAction);
      settings = { ...req };
      return { ...settings };
    },
  };
  return backend;
}

// ---- minimal replay port (SPEC.md "Timer Model") ----

function replay(fake: FakeSession, nowMs: number): Snapshot {
  const length = fake.session.length_seconds;
  const minAction = fake.session.min_action_seconds;
  const startMs = fake.startedMs!;
  const outstanding = new Map(fake.chores.map((c) => [c.id, c.target_count]));
  const doneStack: string[] = [];
  let left = [...outstanding.values()].reduce((a, b) => a + b, 0);
  let active = 0;
  let wallMs = startMs;
  let paused = false;
  let missed = 0;
  let resetAt = 0;
  let deadline = deadlineFor(0, left, length, minAction);
  let status: SessionStatus = SessionStatus.Active;
  let endedMs: number | undefined;

  const reset = () => { resetAt = active; deadline = deadlineFor(active, left, length, minAction); };
  const advanceTo = (targetMs: number): boolean => {
    if (!paused) {
      const target = active + (targetMs - wallMs) / MS_PER_SECOND;
      while (deadline < target && deadline < length) {
        active = deadline;
        missed++;
        reset();
      }
      active = Math.min(target, length);
      if (target >= length) {
        status = SessionStatus.TimeUp;
        endedMs = startMs + length * MS_PER_SECOND;
        wallMs = targetMs;
        return false;
      }
    }
    wallMs = targetMs;
    return true;
  };

  for (const ev of fake.events) {
    if (!advanceTo(ev.atMs)) break;
    if (ev.kind === "done" && ev.choreId) {
      outstanding.set(ev.choreId, (outstanding.get(ev.choreId) ?? 0) - 1);
      doneStack.push(ev.choreId);
      left--;
      if (left === 0) { status = SessionStatus.Completed; endedMs = ev.atMs; break; }
      reset();
    } else if (ev.kind === "undo") {
      const last = doneStack.pop();
      if (last) { outstanding.set(last, (outstanding.get(last) ?? 0) + 1); left++; }
      reset();
    } else if (ev.kind === "pause") {
      paused = true;
    } else if (ev.kind === "resume") {
      paused = false;
      reset();
    } else if (ev.kind === "end") {
      status = SessionStatus.Abandoned; endedMs = ev.atMs; break;
    }
  }
  if (status === SessionStatus.Active) {
    advanceTo(nowMs);
    if (status === SessionStatus.Active && paused) status = SessionStatus.Paused;
  }

  const finished = status !== SessionStatus.Active && status !== SessionStatus.Paused;
  const chores: ChoreProgress[] = fake.chores.map((c) => {
    const out = outstanding.get(c.id) ?? 0;
    return { id: c.id, name: c.name, target: c.target_count, done: c.target_count - out, outstanding: out };
  });
  return {
    status,
    chores,
    done: chores.reduce((a, c) => a + c.done, 0),
    outstanding: left,
    missed,
    session_seconds_left: Math.max(0, length - active),
    action_seconds_left: finished ? 0 : Math.max(0, deadline - active),
    action_seconds_total: deadline - resetAt,
    paused: status === SessionStatus.Paused,
    ended_at: finished && endedMs !== undefined ? new Date(endedMs).toISOString() : undefined,
    as_of: new Date(nowMs).toISOString(),
  };
}

function deadlineFor(active: number, left: number, length: number, minAction: number): number {
  const perAction = Math.max((length - active) / Math.max(left, 1), minAction);
  return Math.min(active + perAction, length);
}
