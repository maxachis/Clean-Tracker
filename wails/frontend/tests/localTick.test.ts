import { describe, expect, it } from "vitest";
import { projectClock } from "../src/lib/services/localTick";
import { SessionStatus } from "../src/lib/types";
import type { Snapshot } from "../src/lib/types";

const snap = (over: Partial<Snapshot> = {}): Snapshot => ({
  status: SessionStatus.Active,
  chores: [],
  done: 0,
  outstanding: 4,
  missed: 0,
  session_seconds_left: 120,
  action_seconds_left: 30,
  action_seconds_total: 30,
  paused: false,
  as_of: "2026-01-01T00:00:00Z",
  ...over,
});

describe("projectClock", () => {
  // Elapsed time is measured on the JS monotonic clock from receipt, so it
  // never depends on the JS wall clock agreeing with Go's.
  it("counts down both clocks by the time since the snapshot arrived", () => {
    const c = projectClock({ snapshot: snap(), receivedAtMs: 1000, nowMs: 11000 });
    expect(c.actionSecondsLeft).toBe(20);
    expect(c.sessionSecondsLeft).toBe(110);
    expect(c.expired).toBe(false);
  });

  it("drives the ring as the fraction of the action countdown remaining", () => {
    const c = projectClock({ snapshot: snap(), receivedAtMs: 0, nowMs: 7500 });
    expect(c.ringFraction).toBeCloseTo(0.75);
  });

  // At zero the UI must fetch a fresh snapshot rather than guess what happens next.
  it("reports expiry when the action countdown reaches zero", () => {
    const c = projectClock({ snapshot: snap(), receivedAtMs: 0, nowMs: 31000 });
    expect(c.actionSecondsLeft).toBe(0);
    expect(c.expired).toBe(true);
  });

  it("reports expiry when the session clock runs out first", () => {
    const s = snap({ session_seconds_left: 5, action_seconds_left: 30 });
    expect(projectClock({ snapshot: s, receivedAtMs: 0, nowMs: 6000 }).expired).toBe(true);
  });

  // Paused time doesn't count against the session.
  it("freezes while paused", () => {
    const c = projectClock({ snapshot: snap({ paused: true, status: SessionStatus.Paused }), receivedAtMs: 0, nowMs: 60000 });
    expect(c.actionSecondsLeft).toBe(30);
    expect(c.expired).toBe(false);
  });

  it("freezes once the session is over", () => {
    const c = projectClock({ snapshot: snap({ status: SessionStatus.TimeUp, action_seconds_left: 0 }), receivedAtMs: 0, nowMs: 60000 });
    expect(c.expired).toBe(false);
  });

  it("does not produce a ring fraction from an empty countdown", () => {
    const c = projectClock({ snapshot: snap({ action_seconds_total: 0, action_seconds_left: 0 }), receivedAtMs: 0, nowMs: 0 });
    expect(c.ringFraction).toBe(0);
  });
});
