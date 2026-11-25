// The single seam between components and the Go bindings. Real bindings are
// used when Wails injected `window.go`; otherwise (plain-browser dev, vitest)
// an in-memory fake stands in.
import * as App from "../../wailsjs/go/main/App";
import { createFakeBackend } from "./fakeBackend";
import type { Backend } from "./types";

// Wails types Go's named string types (SessionStatus) as plain `string`; the
// wire shape is the one in ./types, so narrow once here.
const wire = <T>(call: Promise<unknown>) => call as Promise<T>;

const realBackend: Backend = {
  getDraft: () => wire(App.GetDraft()),
  addChore: (req) => wire(App.AddChore(req)),
  updateChore: (req) => wire(App.UpdateChore(req)),
  removeChore: (id) => wire(App.RemoveChore(id)),
  reorderChores: (ids) => wire(App.ReorderChores(ids)),
  setDraftLength: (minutes) => wire(App.SetDraftLength(minutes)),
  recentChoreNames: () => wire(App.RecentChoreNames()),
  startSession: () => wire(App.StartSession()),
  getRunningSession: () => wire(App.GetRunningSession()),
  logAction: (choreId) => wire(App.LogAction(choreId)),
  undo: () => wire(App.Undo()),
  pause: () => wire(App.Pause()),
  resume: () => wire(App.Resume()),
  endSession: () => wire(App.EndSession()),
  getSummary: (sessionId) => wire(App.GetSummary(sessionId)),
  repeatSession: (sessionId) => wire(App.RepeatSession(sessionId)),
  getSettings: () => wire(App.GetSettings()),
  updateSettings: (req) => wire(App.UpdateSettings(req)),
};

function hasWailsRuntime(): boolean {
  return typeof window !== "undefined" && "go" in window;
}

let active: Backend | null = null;

function current(): Backend {
  active ??= hasWailsRuntime() ? realBackend : createFakeBackend();
  return active;
}

/** Swap the backend (tests give each case a fresh fake with its own clock). */
export function setBackend(backend: Backend | null): void {
  active = backend;
}

export const api: Backend = new Proxy({} as Backend, {
  get: (_target, method: string) => (...args: unknown[]) =>
    (current()[method as keyof Backend] as (...a: unknown[]) => unknown)(...args),
});

/** Wails rejects with a plain string; browsers/fakes reject with Error. */
export function errorMessage(err: unknown): string {
  if (err instanceof Error) return err.message;
  return typeof err === "string" ? err : "Something went wrong";
}
