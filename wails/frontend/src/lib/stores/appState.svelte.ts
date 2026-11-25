import { api } from "../api";
import { isTerminal } from "../types";
import type { Draft, SessionView, Settings, Summary } from "../types";

export const Screen = {
  Loading: "loading",
  Collection: "collection",
  Execution: "execution",
  Summary: "summary",
} as const;
export type Screen = (typeof Screen)[keyof typeof Screen];

const FALLBACK_SETTINGS: Settings = { last_session_minutes: 30, miss_sound: true };

interface AppState {
  screen: Screen;
  settings: Settings;
  draft: Draft | null;
  run: SessionView | null;
  summary: Summary | null;
}

export const appState: AppState = $state({
  screen: Screen.Loading,
  settings: { ...FALLBACK_SETTINGS },
  draft: null,
  run: null,
  summary: null,
});

/** Pick the first screen: a running session, its summary, or the draft. */
export async function boot(): Promise<void> {
  appState.settings = await api.getSettings().catch(() => ({ ...FALLBACK_SETTINGS }));
  const view = await api.getRunningSession();
  if (!view) return showCollection();
  if (isTerminal(view.snapshot.status)) return showSummary(view.session_id);
  showExecution(view);
}

export async function showCollection(): Promise<void> {
  appState.draft = await api.getDraft();
  appState.screen = Screen.Collection;
}

export function showExecution(view: SessionView): void {
  appState.run = view;
  appState.screen = Screen.Execution;
}

export async function showSummary(sessionId: string): Promise<void> {
  appState.summary = await api.getSummary(sessionId);
  appState.run = null;
  appState.screen = Screen.Summary;
}

export async function repeatFromSummary(sessionId: string): Promise<void> {
  appState.draft = await api.repeatSession(sessionId);
  appState.screen = Screen.Collection;
}

export async function setMissSound(enabled: boolean): Promise<void> {
  const next = { ...appState.settings, miss_sound: enabled };
  appState.settings = next;
  appState.settings = await api.updateSettings(next);
}
