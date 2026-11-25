const SECONDS_PER_MINUTE = 60;
const SECONDS_PER_HOUR = 3600;

const pad2 = (n: number) => String(n).padStart(2, "0");

/**
 * Format a countdown as MM:SS (H:MM:SS from one hour up). Partial seconds
 * round up so the display only reads 00:00 once time has truly run out.
 */
export function formatClock(seconds: number): string {
  const whole = Math.max(0, Math.ceil(seconds));
  const hours = Math.floor(whole / SECONDS_PER_HOUR);
  const minutes = Math.floor((whole % SECONDS_PER_HOUR) / SECONDS_PER_MINUTE);
  const secs = whole % SECONDS_PER_MINUTE;
  if (hours > 0) return `${hours}:${pad2(minutes)}:${pad2(secs)}`;
  return `${pad2(minutes)}:${pad2(secs)}`;
}

/** Compact m:ss with no leading zero on minutes, for "~2:30 each". */
export function formatShortClock(seconds: number): string {
  const whole = Math.max(0, Math.round(seconds));
  const minutes = Math.floor(whole / SECONDS_PER_MINUTE);
  return `${minutes}:${pad2(whole % SECONDS_PER_MINUTE)}`;
}
