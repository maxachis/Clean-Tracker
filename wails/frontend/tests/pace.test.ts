import { describe, expect, it } from "vitest";
import { averageSecondsPerAction, choreCompletion, isBelowMinimum, sumCounts } from "../src/lib/services/pace";

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

describe("isBelowMinimum", () => {
  // The even split can be shorter than the minimum per action; the collection
  // screen warns because the real countdown will be held at the minimum.
  it("is true only when the even split is under the minimum", () => {
    expect(isBelowMinimum({ averageSeconds: 9, minimumSeconds: 10 })).toBe(true);
    expect(isBelowMinimum({ averageSeconds: 10, minimumSeconds: 10 })).toBe(false);
    expect(isBelowMinimum({ averageSeconds: 150, minimumSeconds: 10 })).toBe(false);
  });

  it("is false when there is no average to compare", () => {
    expect(isBelowMinimum({ averageSeconds: null, minimumSeconds: 10 })).toBe(false);
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
