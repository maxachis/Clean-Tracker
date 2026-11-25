# Tests

Run with `npm test`. Component tests run the real `App` against the in-memory fake backend (`src/lib/fakeBackend.ts`) under fake timers; `helpers.ts` holds the shared setup.

## countdown.test.ts

- **formatClock**: Verifies the big display reads MM:SS (H:MM:SS from an hour), rounds partial seconds up so 00:00 only appears at true zero, and never goes negative.
- **formatShortClock**: Verifies the compact m:ss form used in "~2:30 each".

## localTick.test.ts

- **projectClock**: Verifies the local countdown advances from the moment a snapshot was received (monotonic clock, not Go's wall clock), drives the ring fraction, reports expiry at zero (action or session), and freezes when paused or finished.

## pace.test.ts

- **averageSecondsPerAction / sumCounts**: Verifies the live "N actions · ~m:ss each" math, including no result for an empty list or zero minutes.
- **isBelowMinimum**: Verifies the pace warning fires only when the even split is under the per-action minimum, never with no average.
- **choreCompletion**: Verifies summary bar fractions (done / target, capped at 1).

## keys.test.ts

- **keyToCommand**: Verifies 1-9 map to the Nth chore, Space to pause/resume, U and Ctrl+Z to undo, and that modified digits and other keys are ignored.

## fakeBackend.test.ts

- **fake backend replay**: Pins the minimal timer port to SPEC.md: even split per action, a miss leaves the action outstanding and shortens the next countdown, time_up at session end, completion on the last action, pause freezes the clock, undo restores an action; the countdown is held at the session's minimum (copied from the setting at start, unaffected by later changes) and out-of-range minimums are rejected.

## collection.test.ts

- **Collection screen**: Start is disabled without chores or with zero minutes; adding chores updates total actions and average time live; backend rejections show inline; suggestion chips add in one tap; chores can be reordered and removed; the "Minimum per action" input saves the setting on change, shows an out-of-range rejection inline, and the pace line warns "below your 0:10 minimum" only when the even split is under it.

## execution.test.ts

- **Execution screen**: Resumes a running session on load; tapping a tile or pressing 1 logs an action; U undoes; finished tiles are disabled; the countdown ticks locally and freezes while paused; a miss bumps the Missed counter and flashes; End needs confirmation; completion, time-up and ending early each show the Summary.

## summary.test.ts

- **Summary screen**: A session that ended while the app was closed opens on its summary; "Repeat this list" refills the draft with the same chores and length; "New list" opens an empty draft.
