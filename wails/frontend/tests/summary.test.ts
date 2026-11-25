import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render } from "@testing-library/svelte";
import App from "../src/App.svelte";
import { freshBackend, SESSION_MINUTES, settle, startRunning } from "./helpers";

const SESSION_MS_PLUS_ONE = (SESSION_MINUTES * 60 + 1) * 1000;

beforeEach(async () => {
  vi.useFakeTimers();
  const fake = freshBackend();
  await startRunning(fake, [{ name: "Dishes", count: 2 }]);
  // The clock runs out while the app is closed; Go settles it on the next load.
  vi.setSystemTime(Date.now() + SESSION_MS_PLUS_ONE);
});
afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

describe("Summary screen", () => {
  // A session that ended while the app was closed is shown once, as a summary.
  it("opens on the summary when the running session already ended", async () => {
    const view = render(App);
    await settle();
    expect(view.getByTestId("ending").textContent).toBe("Time's up");
    expect(view.getByText("0 of 2")).toBeTruthy();
  });

  // "Repeat this list" is the fast path for a regular routine.
  it("repeats the list into a new draft with the same chores and length", async () => {
    const view = render(App);
    await settle();
    await fireEvent.click(view.getByRole("button", { name: "Repeat this list" }));
    await settle();
    expect(view.getByDisplayValue("Dishes")).toBeTruthy();
    expect((view.getByLabelText("Session length in minutes") as HTMLInputElement).value).toBe("10");
  });

  it("starts a new empty list from 'New list'", async () => {
    const view = render(App);
    await settle();
    await fireEvent.click(view.getByRole("button", { name: "New list" }));
    await settle();
    expect(view.queryByDisplayValue("Dishes")).toBeNull();
    expect(view.getByRole("button", { name: "Start" })).toBeTruthy();
  });
});
