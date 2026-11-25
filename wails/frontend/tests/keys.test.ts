import { describe, expect, it } from "vitest";
import { keyToCommand } from "../src/lib/services/keys";

describe("keyToCommand", () => {
  // Digit keys are for reaching chore tiles without the mouse; 1 is the first tile.
  it("maps 1-9 to logging the Nth chore (zero-based index)", () => {
    expect(keyToCommand({ key: "1" })).toEqual({ kind: "log", index: 0 });
    expect(keyToCommand({ key: "9" })).toEqual({ kind: "log", index: 8 });
  });

  it("ignores 0 and non-digit keys", () => {
    expect(keyToCommand({ key: "0" })).toBeNull();
    expect(keyToCommand({ key: "x" })).toBeNull();
  });

  it("maps Space to pause/resume", () => {
    expect(keyToCommand({ key: " " })).toEqual({ kind: "togglePause" });
  });

  it("maps U and Ctrl+Z to undo", () => {
    expect(keyToCommand({ key: "u" })).toEqual({ kind: "undo" });
    expect(keyToCommand({ key: "z", ctrlKey: true })).toEqual({ kind: "undo" });
  });

  // Browser shortcuts like Ctrl+1 (switch tab) shouldn't log an action.
  it("does not treat modified digits as chore keys", () => {
    expect(keyToCommand({ key: "1", ctrlKey: true })).toBeNull();
  });
});
