import { describe, expect, it } from "vitest";
import { formatClock, formatShortClock } from "../src/lib/services/countdown";

describe("formatClock", () => {
  // The big display must be readable at a glance: always two-digit minutes and seconds.
  it("shows minutes and seconds as MM:SS", () => {
    expect(formatClock(150)).toBe("02:30");
    expect(formatClock(5)).toBe("00:05");
  });

  // 00:00 must mean time is genuinely up, not "less than a second left".
  it("rounds partial seconds up so 00:00 appears only at zero", () => {
    expect(formatClock(0.2)).toBe("00:01");
    expect(formatClock(0)).toBe("00:00");
  });

  it("never goes negative", () => {
    expect(formatClock(-3)).toBe("00:00");
  });

  // Session length can exceed an hour; MM:SS alone would read 90:00.
  it("adds an hours part from one hour up", () => {
    expect(formatClock(3600)).toBe("1:00:00");
    expect(formatClock(5405)).toBe("1:30:05");
  });
});

describe("formatShortClock", () => {
  it("drops the leading zero on minutes for the 'about X each' line", () => {
    expect(formatShortClock(150)).toBe("2:30");
    expect(formatShortClock(45)).toBe("0:45");
  });
});
