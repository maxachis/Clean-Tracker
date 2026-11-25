import { vi } from "vitest";
import { setBackend } from "../src/lib/api";
import { createFakeBackend } from "../src/lib/fakeBackend";
import { appState, Screen } from "../src/lib/stores/appState.svelte";
import type { Backend, ChoreInput } from "../src/lib/types";

export const SESSION_MINUTES = 10;

/** Fresh fake backend wired in as the app's backend, with app state reset. */
export function freshBackend(): Backend {
  const fake = createFakeBackend();
  setBackend(fake);
  appState.screen = Screen.Loading;
  appState.draft = null;
  appState.run = null;
  appState.summary = null;
  return fake;
}

/** Build a draft and start it so the app boots straight into Execution. */
export async function startRunning(fake: Backend, chores: ChoreInput[]): Promise<void> {
  await fake.setDraftLength(SESSION_MINUTES);
  for (const chore of chores) await fake.addChore(chore);
  await fake.startSession();
}

/** Let pending promises (and any timers due) run under fake timers. */
export async function settle(): Promise<void> {
  for (let i = 0; i < 5; i++) await vi.advanceTimersByTimeAsync(0);
}

/** Move time forward in 250 ms steps so the ticker fires as in real life. */
export async function elapse(seconds: number): Promise<void> {
  await vi.advanceTimersByTimeAsync(seconds * 1000);
  await settle();
}

/** The slice of the testing-library result that helper functions need. */
export interface View {
  getByTestId(id: string): HTMLElement;
  getAllByTestId(id: string): HTMLElement[];
  getByLabelText(label: string, options?: { selector?: string }): HTMLElement;
  getByRole(role: string, options?: { name?: string }): HTMLElement;
}
