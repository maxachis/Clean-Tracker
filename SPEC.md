# Clean Tracker

## Overview

A desktop app for timed cleaning sessions. You list the cleaning to do and how
many of each action it takes, then work through it against a countdown. Each
action you finish resets the timer for the next one.

The timer works like Book Tracker's "next page due": the time left in the
session is split evenly across the actions still to do. Working fast earns
more time per action, and falling behind leaves less.

## Technology Stack

Same stack as Book Tracker, so its patterns and tooling carry over directly.

| Component | Technology |
|-----------|------------|
| Framework | Wails v2 |
| Backend | Go (flat layout: `store/`, `model/`, `service/`, `migrations/`) |
| Frontend | Svelte 5 + TypeScript (strict) |
| Database | SQLite via `modernc.org/sqlite` (pure Go) |
| Data location | `os.UserConfigDir()/clean-tracker/clean-tracker.db` |

The timer math belongs in Go `service/`, so it can be tested without a browser.
The frontend only calls Go to get each new snapshot and shows a ticking
display.

## Phases

### 1. Collection

You build the list for a session.

- Add a chore: a **name** ("Pick up boxes") and a **count** (5). The count is a whole number, 1 or more.
- Edit a chore's name or count, remove a chore, or reorder the list.
- Set the **session length** in whole minutes. The last value you used is pre-filled.
- Pick from recent chore names, which come from past sessions, so regular chores are quick to add.
- Start the session. This needs at least one chore and a session length above zero.
- The draft is saved as you edit, so closing the app doesn't lose it.

### 2. Execution

A single large countdown shows the time until the next action is due.

- **Per-action time** = session time left ÷ actions left, where actions left is the sum of what's still outstanding on every chore.
- Each chore shows as a tile with its outstanding count. **Tapping any tile** with something outstanding logs one action for that chore. The counts drop by one, a new deadline is set from the new totals, and the countdown resets.
- **When the countdown reaches zero**, one *miss* is recorded and a new countdown starts right away, again from the current totals. *(Decided: a miss doesn't use up an action. It stays outstanding, so the next countdown is shorter. That's the pressure to catch up.)*
- **Undo** reverses the last logged action (a mis-tap) and resets the countdown. Misses can't be undone.
- **Pause** stops both the session clock and the action countdown, and **Resume** continues them. Paused time doesn't count against the session.
- Alongside the countdown: the session time remaining, the actions left, the number done, and the number missed.
- A sound and a visual cue mark each miss, so you notice it from across the room.

### Session end

A session ends in one of three ways:

| Ending | Trigger |
|--------|---------|
| `completed` | The last outstanding action is logged |
| `time_up` | The session clock reaches zero with actions still outstanding |
| `abandoned` | You end it early |

The summary shows how long it took, the number done out of the total, the
misses, and how far each chore got. From there you can start a new session
with the same list.

## Timer Model

The whole execution state is **replayed** from the session's stored facts:
its length, start time, chores and event log, plus the current time. Nothing
about misses or deadlines is stored. That makes reopening the app after a
close correct for free, and lets the core be one pure, table-tested function.

**Active time.** The session clock counts *active seconds*: wall time since
`started_at` minus time spent paused. Every deadline below is in active
seconds. `L` is the session length in seconds.

**Reset points.** The countdown restarts at a *reset point*: the session
start, every `done` event, every `undo` event, every `resume` event, and
every miss.

**Deadline rule.** At a reset point at active time `t`, with `n` actions
outstanding:

```
per_action = max((L - t) / n, MIN_ACTION_SECONDS)   // MIN_ACTION_SECONDS = 1
deadline   = min(t + per_action, L)
```

**Replay.** Walk the events in time order. Before applying each event, emit
a miss for every deadline that falls before it, treating each miss as a new
reset point with the same `n`. Do the same between the last event and `now`.
Then apply the event:

- `done`: decrements one chore's outstanding count.
- `undo`: reverses the most recent `done` that hasn't already been undone.
- `pause` and `resume`: stop and restart the active clock. The countdown doesn't advance while paused, and resuming starts a fresh countdown.
- `end`: the session is abandoned.

**Termination**, checked in this order:

1. `n` reaches 0: `completed`, at that event's time.
2. Active time reaches `L`: `time_up`. A deadline landing exactly on `L`, where `n` is still above 0, is not counted as a miss, since time-up already covers it.
3. An `end` event: `abandoned`.

**Why the floor exists.** Without it, repeated misses shrink the countdown
geometrically toward `L` without ever reaching it, which means infinitely many
misses. Each miss multiplies the remaining time by `(1 − 1/n)`. The
one-second floor keeps the miss count finite. Near the end of an overrun
session, misses still come fast. That's intended.

**Snapshot.** Replay returns everything the UI needs:

- the status
- each chore's done and outstanding counts
- the totals: done, outstanding and missed
- the session seconds left
- the seconds left on the current countdown and the countdown's full length, for a progress ring
- whether the session is paused
- the `now` the snapshot was computed at

The frontend ticks the display locally from that and asks for a fresh
snapshot when the countdown reaches zero. It plays the miss cue when the
missed total goes up.

## Data Model

### Session

| Field | Type | Notes |
|-------|------|-------|
| id | UUID | |
| length_seconds | Integer | The planned session length. Editable while a draft. |
| created_at | DateTime | |
| started_at | DateTime | Null while a draft |
| ended_at | DateTime | Written once replay finds a terminal status |
| end_status | Enum | `completed`, `time_up` or `abandoned`. Null until it ends. |

The status is derived: `draft` (not yet started), `active` or `paused`
(running), or the stored `end_status`. At most one draft and at most one
running session exist at a time.

### Chore (belongs to a session)

| Field | Type | Notes |
|-------|------|-------|
| id | UUID | |
| session_id | UUID | |
| name | String | Trimmed, not empty |
| target_count | Integer | ≥ 1 |
| position | Integer | List order |

Chores are only editable while their session is a draft.

### Event (a log for the session)

| Field | Type | Notes |
|-------|------|-------|
| id | Integer | Autoincrement. Breaks ties between events with the same timestamp. |
| session_id | UUID | |
| kind | Enum | `done`, `undo`, `pause`, `resume`, `end` |
| chore_id | UUID | Only for `done` |
| at | DateTime | UTC, millisecond precision |

How many of a chore are done is computed from these events, not stored. Undo
writes an `undo` event instead of deleting anything, so the history stays
complete.

### Settings

| Field | Default | Notes |
|-------|---------|-------|
| last_session_minutes | 30 | Pre-fills the session length |
| miss_sound | on | |

## Edge Cases

- **App closed mid-session:** reopening the app replays the log, so the session resumes. Misses that fell due while the app was closed get counted. If the session clock ran out in the meantime, the session ends as `time_up`.
- **Tapping a tile with nothing outstanding:** rejected. Its tile is shown as done.
- **Pausing while paused, or resuming while running:** rejected.
- **Undo with nothing to undo:** rejected.

## Out of Scope (for now)

- Recurring or scheduled sessions
- Statistics beyond the per-session summary
- CSV import and export
- Different times for different chore types, which was option (c) of the timing question
