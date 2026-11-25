export type KeyCommand =
  | { kind: "log"; index: number }
  | { kind: "togglePause" }
  | { kind: "undo" };

export interface KeyInput {
  key: string;
  ctrlKey?: boolean;
  metaKey?: boolean;
}

const FIRST_DIGIT_KEY = 1;
const LAST_DIGIT_KEY = 9;

/** Map a keypress to an Execution command: 1-9 log, Space pause/resume, U / Ctrl+Z undo. */
export function keyToCommand(input: KeyInput): KeyCommand | null {
  const { key, ctrlKey = false, metaKey = false } = input;
  const lower = key.toLowerCase();
  if ((ctrlKey || metaKey) && lower === "z") return { kind: "undo" };
  if (ctrlKey || metaKey) return null;
  if (key === " ") return { kind: "togglePause" };
  if (lower === "u") return { kind: "undo" };
  const digit = Number(key);
  if (Number.isInteger(digit) && digit >= FIRST_DIGIT_KEY && digit <= LAST_DIGIT_KEY) {
    return { kind: "log", index: digit - FIRST_DIGIT_KEY };
  }
  return null;
}
