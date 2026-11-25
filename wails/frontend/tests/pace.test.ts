import { describe, expect, it } from "vitest";
import { averageSecondsPerAction, choreCompletion, sumCounts } from "../src/lib/services/pace";

describe("averageSecondsPerAction", () => {
  // The collection screen shows "12 actions · ~2:30 each" for 30 minutes.
  it("splits the session evenly across all actions", () => {
    expect(averageSecondsPerAction({ totalActions: 12, minutes: 30 })).toBe(150);
  });

  it("is undefined with no actions or no time", () => {
    expect(averageSecondsPerAction({ totalActions: 0, minutes: 30 })).toBeNull();
    expect(averageSecondsPerAction({ totalActions: 5, minutes: 0 })).toBeNull();
  });
});

describe("sumCounts", () => {
  it("adds up every chore's count", () => {
    expect(sumCounts([{ count: 5 }, { count: 7 }])).toBe(12);
  });
});

describe("choreCompletion", () => {
  it("is done divided by target, capped at 1", () => {
    expect(choreCompletion({ done: 1, target: 4 })).toBe(0.25);
    expect(choreCompletion({ done: 9, target: 4 })).toBe(1);
    expect(choreCompletion({ done: 0, target: 0 })).toBe(0);
  });
});
