package service

import (
	"math"
	"testing"
	"time"

	"github.com/maxachis/clean-tracker/wails/model"
)

const (
	floatTolerance = 1e-6
	sessionLength  = 600
	choreA         = "chore-a"
	choreB         = "chore-b"
)

var replayBase = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func atSeconds(sec float64) time.Time {
	return replayBase.Add(time.Duration(sec * float64(time.Second)))
}

func standardChores() []model.Chore {
	return []model.Chore{
		{ID: choreA, Name: "Pick up boxes", TargetCount: 3, Position: 0},
		{ID: choreB, Name: "Wipe counters", TargetCount: 2, Position: 1},
	}
}

func singleChore(count int) []model.Chore {
	return []model.Chore{{ID: choreA, Name: "Pick up boxes", TargetCount: count}}
}

func doneEvent(id int64, chore string, sec float64) model.Event {
	return model.Event{ID: id, Kind: model.EventDone, ChoreID: chore, At: atSeconds(sec)}
}

func plainEvent(id int64, kind model.EventKind, sec float64) model.Event {
	return model.Event{ID: id, Kind: kind, At: atSeconds(sec)}
}

type wantSnapshot struct {
	status      model.SessionStatus
	done        int
	outstanding int
	missed      int
	sessionLeft float64
	actionLeft  float64
	actionTotal float64
	paused      bool
	endedAtSec  *float64
	perChore    map[string]int // outstanding per chore, checked when set
}

func secPtr(v float64) *float64 { return &v }

func assertClose(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > floatTolerance {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

func assertSnapshot(t *testing.T, got model.Snapshot, want wantSnapshot) {
	t.Helper()
	if got.Status != want.status {
		t.Errorf("status = %q, want %q", got.Status, want.status)
	}
	if got.Done != want.done || got.Outstanding != want.outstanding || got.Missed != want.missed {
		t.Errorf("done/outstanding/missed = %d/%d/%d, want %d/%d/%d",
			got.Done, got.Outstanding, got.Missed, want.done, want.outstanding, want.missed)
	}
	if got.Paused != want.paused {
		t.Errorf("paused = %v, want %v", got.Paused, want.paused)
	}
	assertClose(t, "session seconds left", got.SessionSecondsLeft, want.sessionLeft)
	assertClose(t, "action seconds left", got.ActionSecondsLeft, want.actionLeft)
	assertClose(t, "action seconds total", got.ActionSecondsTotal, want.actionTotal)
	assertEndedAt(t, got.EndedAt, want.endedAtSec)
	for id, wantOut := range want.perChore {
		if gotOut := outstandingOf(got, id); gotOut != wantOut {
			t.Errorf("chore %s outstanding = %d, want %d", id, gotOut, wantOut)
		}
	}
}

func assertEndedAt(t *testing.T, got *time.Time, wantSec *float64) {
	t.Helper()
	if wantSec == nil {
		if got != nil {
			t.Errorf("ended_at = %v, want nil", got)
		}
		return
	}
	if got == nil {
		t.Fatalf("ended_at = nil, want %v", atSeconds(*wantSec))
	}
	assertClose(t, "ended_at seconds", got.Sub(replayBase).Seconds(), *wantSec)
}

func outstandingOf(s model.Snapshot, choreID string) int {
	for _, c := range s.Chores {
		if c.ID == choreID {
			return c.Outstanding
		}
	}
	return -1
}

func TestReplay(t *testing.T) {
	tests := []struct {
		name   string
		length int
		chores []model.Chore
		events []model.Event
		nowSec float64
		want   wantSnapshot
	}{
		{
			// With nothing logged, the whole session is divided evenly by the actions to do.
			name:   "fresh session splits L evenly (L=600, 5 actions gives a 120s countdown)",
			length: sessionLength, chores: standardChores(), nowSec: 0,
			want: wantSnapshot{status: model.StatusActive, outstanding: 5, sessionLeft: 600, actionLeft: 120, actionTotal: 120},
		},
		{
			name:   "a done resets the countdown using the new totals",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{doneEvent(1, choreA, 60)}, nowSec: 60,
			want: wantSnapshot{status: model.StatusActive, done: 1, outstanding: 4, sessionLeft: 540, actionLeft: 135, actionTotal: 135,
				perChore: map[string]int{choreA: 2, choreB: 2}},
		},
		{
			name:   "the countdown keeps running down after a done",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{doneEvent(1, choreA, 60)}, nowSec: 100,
			want: wantSnapshot{status: model.StatusActive, done: 1, outstanding: 4, sessionLeft: 500, actionLeft: 95, actionTotal: 135},
		},
		{
			// 590s over 4 remaining actions beats the 120s an on-pace first action would leave.
			name:   "working fast earns more time per action",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{doneEvent(1, choreA, 10)}, nowSec: 10,
			want: wantSnapshot{status: model.StatusActive, done: 1, outstanding: 4, sessionLeft: 590, actionLeft: 147.5, actionTotal: 147.5},
		},
		{
			name:   "missing a deadline counts one miss and leaves the action outstanding with a shorter next countdown",
			length: sessionLength, chores: standardChores(), nowSec: 121,
			want: wantSnapshot{status: model.StatusActive, outstanding: 5, missed: 1, sessionLeft: 479, actionLeft: 95, actionTotal: 96},
		},
		{
			// Misses at 120, 216, 292.8 and 354.24; the fifth deadline is 403.392.
			name:   "several misses while the app was closed are all counted on replay",
			length: sessionLength, chores: standardChores(), nowSec: 400,
			want: wantSnapshot{status: model.StatusActive, outstanding: 5, missed: 4, sessionLeft: 200, actionLeft: 3.392, actionTotal: 49.152},
		},
		{
			name:   "a miss before a done is counted and the done still resets the countdown",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{doneEvent(1, choreA, 130)}, nowSec: 130,
			want: wantSnapshot{status: model.StatusActive, done: 1, outstanding: 4, missed: 1, sessionLeft: 470, actionLeft: 117.5, actionTotal: 117.5},
		},
		{
			// L=10, 2 actions: misses at 5, 7.5, 8.75, 9.75; then the 1s floor puts the next deadline past L.
			name:   "the 1-second floor keeps misses finite near L",
			length: 10, chores: singleChore(2), nowSec: 9.8,
			want: wantSnapshot{status: model.StatusActive, outstanding: 2, missed: 4, sessionLeft: 0.2, actionLeft: 0.2, actionTotal: 0.25},
		},
		{
			name:   "after the floor-limited misses the session ends time_up at L with no further misses",
			length: 10, chores: singleChore(2), nowSec: 1000,
			want: wantSnapshot{status: model.StatusTimeUp, outstanding: 2, missed: 4, endedAtSec: secPtr(10)},
		},
		{
			// The single deadline lands exactly on L, which time-up already covers.
			name:   "reaching L with work outstanding is time_up with no extra miss at L",
			length: 60, chores: singleChore(1), nowSec: 100,
			want: wantSnapshot{status: model.StatusTimeUp, outstanding: 1, missed: 0, endedAtSec: secPtr(60)},
		},
		{
			name:   "logging the last action is completed even mid-countdown",
			length: sessionLength, chores: singleChore(1),
			events: []model.Event{doneEvent(1, choreA, 30)}, nowSec: 500,
			want: wantSnapshot{status: model.StatusCompleted, done: 1, sessionLeft: 570, endedAtSec: secPtr(30)},
		},
		{
			name:   "undo restores a count and resets the countdown",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{doneEvent(1, choreA, 60), plainEvent(2, model.EventUndo, 90)}, nowSec: 90,
			want: wantSnapshot{status: model.StatusActive, outstanding: 5, sessionLeft: 510, actionLeft: 102, actionTotal: 102,
				perChore: map[string]int{choreA: 3, choreB: 2}},
		},
		{
			name:   "undo after undo reverses the next-older done",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{
				doneEvent(1, choreA, 10), doneEvent(2, choreB, 20),
				plainEvent(3, model.EventUndo, 30), plainEvent(4, model.EventUndo, 40),
			}, nowSec: 40,
			want: wantSnapshot{status: model.StatusActive, outstanding: 5, sessionLeft: 560, actionLeft: 112, actionTotal: 112,
				perChore: map[string]int{choreA: 3, choreB: 2}},
		},
		{
			name:   "a single undo reverses only the most recent done",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{
				doneEvent(1, choreA, 10), doneEvent(2, choreB, 20), plainEvent(3, model.EventUndo, 30),
			}, nowSec: 30,
			want: wantSnapshot{status: model.StatusActive, done: 1, outstanding: 4, sessionLeft: 570, actionLeft: 142.5, actionTotal: 142.5,
				perChore: map[string]int{choreA: 2, choreB: 2}},
		},
		{
			name:   "pause freezes both clocks (no misses however long it stays paused)",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{plainEvent(1, model.EventPause, 30)}, nowSec: 10000,
			want: wantSnapshot{status: model.StatusPaused, outstanding: 5, sessionLeft: 570, actionLeft: 90, actionTotal: 120, paused: true},
		},
		{
			name:   "resume starts a fresh countdown from remaining time",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{plainEvent(1, model.EventPause, 30), plainEvent(2, model.EventResume, 100)}, nowSec: 100,
			want: wantSnapshot{status: model.StatusActive, outstanding: 5, sessionLeft: 570, actionLeft: 114, actionTotal: 114},
		},
		{
			name:   "paused time does not count against the session after resume",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{plainEvent(1, model.EventPause, 30), plainEvent(2, model.EventResume, 100)}, nowSec: 110,
			want: wantSnapshot{status: model.StatusActive, outstanding: 5, sessionLeft: 560, actionLeft: 104, actionTotal: 114},
		},
		{
			// After resuming at wall 100 (active 30) the deadline is active 144; wall 230 is active 160.
			name:   "misses are counted in active time after a pause",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{plainEvent(1, model.EventPause, 30), plainEvent(2, model.EventResume, 100)}, nowSec: 230,
			want: wantSnapshot{status: model.StatusActive, outstanding: 5, missed: 1, sessionLeft: 440, actionLeft: 75.2, actionTotal: 91.2},
		},
		{
			name:   "end event abandons the session at the end time",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{plainEvent(1, model.EventEnd, 50)}, nowSec: 900,
			want: wantSnapshot{status: model.StatusAbandoned, outstanding: 5, sessionLeft: 550, endedAtSec: secPtr(50)},
		},
		{
			// Sorted by ID the undo (nothing to undo yet) precedes the done; sorted wrongly the done would be undone.
			name:   "same-timestamp events order by ID",
			length: sessionLength, chores: standardChores(),
			events: []model.Event{doneEvent(2, choreA, 10), plainEvent(1, model.EventUndo, 10)}, nowSec: 10,
			want: wantSnapshot{status: model.StatusActive, done: 1, outstanding: 4, sessionLeft: 590, actionLeft: 147.5, actionTotal: 147.5},
		},
		{
			name:   "events after the session has ended are ignored",
			length: sessionLength, chores: singleChore(1),
			events: []model.Event{doneEvent(1, choreA, 30), plainEvent(2, model.EventEnd, 40)}, nowSec: 50,
			want: wantSnapshot{status: model.StatusCompleted, done: 1, sessionLeft: 570, endedAtSec: secPtr(30)},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Replay(ReplayInput{
				LengthSeconds: tc.length,
				StartedAt:     replayBase,
				Chores:        tc.chores,
				Events:        tc.events,
				Now:           atSeconds(tc.nowSec),
			})
			assertSnapshot(t, got, tc.want)
			if !got.AsOf.Equal(atSeconds(tc.nowSec)) {
				t.Errorf("as_of = %v, want now", got.AsOf)
			}
		})
	}
}
