import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render } from "@testing-library/svelte";
import App from "../src/App.svelte";
import { type View, elapse, freshBackend, settle, startRunning } from "./helpers";
import type { Backend } from "../src/lib/types";

let fake: Backend;

beforeEach(async () => {
  vi.useFakeTimers();
  fake = freshBackend();
  await startRunning(fake, [{ name: "Dishes", count: 2 }, { name: "Trash", count: 1 }]);
});
afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

async function openExecution() {
  const view = render(App);
  await settle();
  return view;
}

const stat = (view: View, id: string) => view.getByTestId(`stat-${id}`).textContent;
const tileCounts = (view: View) =>
  view.getAllByTestId("tile").map((t) => t.querySelector(".count")!.textContent);

describe("Execution screen", () => {
  it("resumes a running session on app load", async () => {
    const view = await openExecution();
    expect(stat(view, "actions-left")).toBe("3");
    expect(view.getByRole("timer").textContent).toBe("03:20");
  });

  // Tapping a tile is the whole interaction: one tap, one action done.
  it("decrements a chore's outstanding count when its tile is tapped", async () => {
    const view = await openExecution();
    await fireEvent.click(view.getAllByTestId("tile")[0]);
    await settle();
    expect(tileCounts(view)).toEqual(["1", "1"]);
    expect(stat(view, "done")).toBe("1");
  });

  it("logs the first chore when 1 is pressed", async () => {
    const view = await openExecution();
    await fireEvent.keyDown(window, { key: "1" });
    await settle();
    expect(tileCounts(view)).toEqual(["1", "1"]);
  });

  it("undoes the last action with U", async () => {
    const view = await openExecution();
    await fireEvent.keyDown(window, { key: "2" });
    await settle();
    expect(tileCounts(view)).toEqual(["2", "✓"]);
    await fireEvent.keyDown(window, { key: "u" });
    await settle();
    expect(tileCounts(view)).toEqual(["2", "1"]);
  });

  it("shows finished chores as done and disabled", async () => {
    const view = await openExecution();
    await fireEvent.click(view.getAllByTestId("tile")[1]);
    await settle();
    expect((view.getAllByTestId("tile")[1] as HTMLButtonElement).disabled).toBe(true);
  });

  it("ticks the countdown down locally between snapshots", async () => {
    const view = await openExecution();
    await elapse(20);
    expect(view.getByRole("timer").textContent).toBe("03:00");
  });

  // Paused time doesn't count against the session: the display must hold still.
  it("freezes the countdown while paused and continues on resume", async () => {
    const view = await openExecution();
    await elapse(10);
    await fireEvent.keyDown(window, { key: " " });
    await settle();
    const frozen = view.getByRole("timer").textContent;
    await elapse(60);
    expect(view.getByRole("timer").textContent).toBe(frozen);
    expect(view.getByRole("button", { name: "Resume" })).toBeTruthy();
    expect(stat(view, "missed")).toBe("0");
  });

  // A miss must be unmissable and counted: the local clock hits zero, Go is asked, missed goes up.
  it("increments the missed counter and flashes when the countdown reaches zero", async () => {
    const view = await openExecution();
    await elapse(201);
    expect(stat(view, "missed")).toBe("1");
    expect(view.getByTestId("miss-flash")).toBeTruthy();
    expect(stat(view, "actions-left")).toBe("3");
  });

  it("asks for confirmation before ending the session", async () => {
    const view = await openExecution();
    await fireEvent.click(view.getByRole("button", { name: "End" }));
    expect(view.getByText("End this session?")).toBeTruthy();
    await fireEvent.click(view.getByRole("button", { name: "Keep going" }));
    expect(view.queryByText("End this session?")).toBeNull();
    expect(view.getByRole("button", { name: "End" })).toBeTruthy();
  });

  it("shows the Summary when the last action is logged", async () => {
    const view = await openExecution();
    await fireEvent.keyDown(window, { key: "2" });
    await settle();
    await fireEvent.keyDown(window, { key: "1" });
    await settle();
    await fireEvent.keyDown(window, { key: "1" });
    await settle();
    expect(view.getByTestId("ending").textContent).toBe("Completed");
  });

  it("shows the Summary as 'Time's up' when the session clock runs out", async () => {
    const view = await openExecution();
    await elapse(601);
    expect(view.getByTestId("ending").textContent).toBe("Time's up");
  });

  it("shows the Summary as 'Ended early' after confirming End", async () => {
    const view = await openExecution();
    await fireEvent.click(view.getByRole("button", { name: "End" }));
    await fireEvent.click(view.getByRole("button", { name: "End session" }));
    await settle();
    expect(view.getByTestId("ending").textContent).toBe("Ended early");
  });
});
