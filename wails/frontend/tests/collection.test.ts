import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render } from "@testing-library/svelte";
import App from "../src/App.svelte";
import { type View, freshBackend, settle } from "./helpers";

beforeEach(() => {
  vi.useFakeTimers();
  freshBackend();
});
afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

async function openCollection() {
  const view = render(App);
  await settle();
  return view;
}

async function addChore(view: View, name: string, count: number) {
  await fireEvent.input(view.getByLabelText("New chore name"), { target: { value: name } });
  const stepper = view.getByLabelText("count for new chore", { selector: "input" });
  await fireEvent.change(stepper, { target: { value: String(count) } });
  await fireEvent.click(view.getByRole("button", { name: "Add" }));
  await settle();
}

const MINIMUM_LABEL = "Minimum per action in seconds";

async function setMinimum(view: View, value: string) {
  const input = view.getByLabelText(MINIMUM_LABEL);
  await fireEvent.input(input, { target: { value } });
  await fireEvent.change(input);
  await settle();
}

describe("Collection screen", () => {
  // You can't start an empty session; the button must say so by being disabled.
  it("disables Start until there is at least one chore", async () => {
    const view = await openCollection();
    expect((view.getByRole("button", { name: "Start" }) as HTMLButtonElement).disabled).toBe(true);
    await addChore(view, "Dishes", 1);
    expect((view.getByRole("button", { name: "Start" }) as HTMLButtonElement).disabled).toBe(false);
  });

  // The live line is how you judge whether the list fits the time you have.
  it("updates total actions and average time per action when a chore is added", async () => {
    const view = await openCollection();
    await addChore(view, "Pick up boxes", 5);
    await addChore(view, "Dishes", 7);
    expect(view.getByText("12 actions")).toBeTruthy();
    expect(view.getByText(/~2:30 each/)).toBeTruthy();
  });

  it("disables Start when the session length is zero and shows the backend's reason", async () => {
    const view = await openCollection();
    await addChore(view, "Dishes", 1);
    const minutes = view.getByLabelText("Session length in minutes");
    await fireEvent.input(minutes, { target: { value: "0" } });
    await fireEvent.change(minutes);
    await settle();
    expect((view.getByRole("button", { name: "Start" }) as HTMLButtonElement).disabled).toBe(true);
    expect(view.getByRole("alert").textContent).toMatch(/above zero minutes/);
  });

  it("shows a backend rejection for a blank chore name inline", async () => {
    const view = await openCollection();
    await fireEvent.click(view.getByRole("button", { name: "Add" }));
    await settle();
    expect(view.getByRole("alert").textContent).toMatch(/must not be empty/);
  });

  // Regular chores should be one tap to add.
  it("adds a recent chore name with one tap on its suggestion chip", async () => {
    const view = await openCollection();
    await fireEvent.click(view.getByRole("button", { name: "Dishes" }));
    await settle();
    expect(view.getByDisplayValue("Dishes")).toBeTruthy();
  });

  it("reorders chores with the down button", async () => {
    const view = await openCollection();
    await addChore(view, "First", 1);
    await addChore(view, "Second", 1);
    await fireEvent.click(view.getByRole("button", { name: "Move First down" }));
    await settle();
    const names = view.getAllByLabelText("Chore name").map((i) => (i as HTMLInputElement).value);
    expect(names).toEqual(["Second", "First"]);
  });

  it("removes a chore", async () => {
    const view = await openCollection();
    await addChore(view, "Dishes", 1);
    await fireEvent.click(view.getByRole("button", { name: "Remove Dishes" }));
    await settle();
    expect(view.queryByDisplayValue("Dishes")).toBeNull();
  });

  // The minimum is a setting; each new session copies it when it starts.
  it("saves the minimum per action as a setting when it changes", async () => {
    const fake = freshBackend();
    const view = await openCollection();
    expect((view.getByLabelText(MINIMUM_LABEL) as HTMLInputElement).value).toBe("10");
    await setMinimum(view, "25");
    expect((await fake.getSettings()).min_action_seconds).toBe(25);
  });

  it("rejects an out-of-range minimum inline and keeps the saved value", async () => {
    const fake = freshBackend();
    const view = await openCollection();
    await setMinimum(view, "601");
    expect(view.getByRole("alert").textContent).toMatch(/minimum per action/);
    expect((await fake.getSettings()).min_action_seconds).toBe(10);
  });

  // Splitting 30 minutes over 200 actions is 9s each, under the 10s default minimum.
  it("warns when the even split is below the minimum, and not otherwise", async () => {
    const view = await openCollection();
    await addChore(view, "Dishes", 12);
    expect(view.queryByText(/below your/)).toBeNull();
    await addChore(view, "Crumbs", 188);
    expect(view.getByText(/200 actions/)).toBeTruthy();
    expect(view.getByText(/below your 0:10 minimum/)).toBeTruthy();
  });
});
