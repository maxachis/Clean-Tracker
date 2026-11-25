import { beforeEach, describe, expect, it } from "vitest";
import { createFakeBackend } from "../src/lib/fakeBackend";
import { SessionStatus } from "../src/lib/types";
import type { Backend } from "../src/lib/types";

// The fake is dev/test scaffolding, but the component tests lean on its timer
// matching the spec, so pin the spec's behavior here.
let nowMs = 0;
let fake: Backend;
const advance = (seconds: number) => (nowMs += seconds * 1000);

beforeEach(async () => {
  nowMs = Date.parse("2026-01-01T00:00:00Z");
  fake = createFakeBackend({ now: () => nowMs });
  await fake.setDraftLength(10);
  await fake.addChore({ name: "Dishes", count: 2 });
});

describe("fake backend replay", () => {
  it("splits the session evenly across outstanding actions", async () => {
    const snap = await fake.startSession();
    expect(snap.action_seconds_total).toBe(300);
    expect(snap.action_seconds_left).toBe(300);
  });

  // SPEC: a miss does not use up an action; the next countdown is shorter.
  it("records a miss when the countdown runs out and keeps the action outstanding", async () => {
    await fake.startSession();
    advance(301);
    const view = await fake.getRunningSession();
    expect(view?.snapshot.missed).toBe(1);
    expect(view?.snapshot.outstanding).toBe(2);
    expect(view?.snapshot.action_seconds_total).toBeCloseTo(150);
  });

  // SPEC: per_action = max(even split, M). L=60, 2 actions, M=40: first countdown 40s (not 30s); after the miss the last countdown is capped at L (20s).
  it("holds the countdown at the session's minimum and keeps it when the setting changes", async () => {
    await fake.setDraftLength(1);
    const settings = await fake.getSettings();
    await fake.updateSettings({ ...settings, min_action_seconds: 40 });
    const snap = await fake.startSession();
    expect(snap.action_seconds_total).toBe(40);
    await fake.updateSettings({ ...settings, min_action_seconds: 1 });
    advance(41);
    const view = await fake.getRunningSession();
    expect(view?.session.min_action_seconds).toBe(40);
    expect(view?.snapshot.missed).toBe(1);
    expect(view?.snapshot.action_seconds_total).toBeCloseTo(20); // capped at L, shorter than M
  });

  it("rejects a minimum outside 1-600 seconds", async () => {
    const settings = await fake.getSettings();
    await expect(fake.updateSettings({ ...settings, min_action_seconds: 0 })).rejects.toThrow(/minimum per action/);
    await expect(fake.updateSettings({ ...settings, min_action_seconds: 601 })).rejects.toThrow(/minimum per action/);
  });

  it("ends as time_up when the session clock reaches zero", async () => {
    await fake.startSession();
    advance(601);
    const view = await fake.getRunningSession();
    expect(view?.snapshot.status).toBe(SessionStatus.TimeUp);
    expect(await fake.getRunningSession()).toBeNull();
  });

  it("completes when the last action is logged", async () => {
    const snap = await fake.startSession();
    const id = snap.chores[0].id;
    await fake.logAction(id);
    const last = await fake.logAction(id);
    expect(last.status).toBe(SessionStatus.Completed);
  });

  it("does not run the countdown while paused", async () => {
    await fake.startSession();
    advance(100);
    await fake.pause();
    advance(1000);
    const view = await fake.getRunningSession();
    expect(view?.snapshot.paused).toBe(true);
    expect(view?.snapshot.missed).toBe(0);
    expect(view?.snapshot.session_seconds_left).toBe(500);
  });

  it("undo gives back the last logged action", async () => {
    const snap = await fake.startSession();
    await fake.logAction(snap.chores[0].id);
    const after = await fake.undo();
    expect(after.outstanding).toBe(2);
  });
});
