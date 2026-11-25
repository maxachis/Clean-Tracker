package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/maxachis/clean-tracker/wails/model"
	"github.com/maxachis/clean-tracker/wails/service"
	"github.com/maxachis/clean-tracker/wails/store"
)

var appTestStart = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

type appClock struct{ now time.Time }

func (c *appClock) Now() time.Time { return c.now }

func newTestApp(t *testing.T, dsn string, clock *appClock) *App {
	t.Helper()
	st, err := store.Open(dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return NewApp(service.New(st, clock.Now))
}

// must unwraps a call's result and panics (failing the test with the
// message) on error, so multi-step flows read as straight lines.
func must[T any](v T, err error) T {
	if err != nil {
		panic("unexpected error: " + err.Error())
	}
	return v
}

// exerciseEveryEndpoint drives each IPC method once, in a realistic order,
// against whichever database the App was built on.
func exerciseEveryEndpoint(t *testing.T, app *App, clock *appClock) {
	t.Helper()
	draft := must(app.GetDraft())
	if draft.Session.LengthSeconds != model.DefaultLastSessionMinutes*model.SecondsPerMinute {
		t.Fatalf("draft length = %d", draft.Session.LengthSeconds)
	}
	draft = must(app.SetDraftLength(10))
	must(app.AddChore(model.AddChoreRequest{Name: "Boxes", Count: 3}))
	draft = must(app.AddChore(model.AddChoreRequest{Name: "Counters", Count: 2}))
	boxes, counters := draft.Chores[0].ID, draft.Chores[1].ID
	draft = must(app.UpdateChore(model.UpdateChoreRequest{ID: counters, Name: "Counters", Count: 1}))
	must(app.ReorderChores([]string{counters, boxes}))
	draft = must(app.ReorderChores([]string{boxes, counters}))
	if len(draft.Chores) != 2 || draft.Chores[1].TargetCount != 1 {
		t.Fatalf("draft chores = %+v", draft.Chores)
	}
	extra := must(app.AddChore(model.AddChoreRequest{Name: "Extra", Count: 1}))
	must(app.RemoveChore(extra.Chores[2].ID))

	snap := must(app.StartSession())
	if snap.Outstanding != 4 || snap.Status != model.StatusActive {
		t.Fatalf("start snapshot = %+v", snap)
	}
	view := must(app.GetRunningSession())
	if view == nil {
		t.Fatal("expected a running session")
	}

	clock.now = clock.now.Add(10 * time.Second)
	snap = must(app.LogAction(boxes))
	if snap.Done != 1 {
		t.Fatalf("after log = %+v", snap)
	}
	snap = must(app.Undo())
	if snap.Done != 0 {
		t.Fatalf("after undo = %+v", snap)
	}
	if snap = must(app.Pause()); !snap.Paused {
		t.Fatalf("after pause = %+v", snap)
	}
	if snap = must(app.Resume()); snap.Paused {
		t.Fatalf("after resume = %+v", snap)
	}
	snap = must(app.EndSession())
	if snap.Status != model.StatusAbandoned {
		t.Fatalf("after end = %+v", snap)
	}

	summary := must(app.GetSummary(view.SessionID))
	if summary.Snapshot.Status != model.StatusAbandoned || summary.Session.EndedAt == nil {
		t.Fatalf("summary = %+v", summary)
	}
	names := must(app.RecentChoreNames())
	if len(names) != 2 || names[0] != "Boxes" {
		t.Fatalf("recent names = %v", names)
	}
	repeated := must(app.RepeatSession(view.SessionID))
	if len(repeated.Chores) != 2 || repeated.Session.LengthSeconds != 600 {
		t.Fatalf("repeat = %+v", repeated)
	}

	settings := must(app.UpdateSettings(model.UpdateSettingsRequest{LastSessionMinutes: 25, MissSound: false}))
	if got := must(app.GetSettings()); got != settings || got.MissSound {
		t.Fatalf("settings = %+v", got)
	}
}

func TestApp_EveryEndpoint_InMemory(t *testing.T) {
	clock := &appClock{now: appTestStart}
	exerciseEveryEndpoint(t, newTestApp(t, ":memory:", clock), clock)
}

func TestApp_EveryEndpoint_FileBacked(t *testing.T) {
	// The real driver and a file on disk, the way the shipped app runs.
	clock := &appClock{now: appTestStart}
	exerciseEveryEndpoint(t, newTestApp(t, filepath.Join(t.TempDir(), "ct.db"), clock), clock)
}

func TestApp_RejectionsReachTheFrontendAsErrors(t *testing.T) {
	clock := &appClock{now: appTestStart}
	app := newTestApp(t, ":memory:", clock)
	if _, err := app.StartSession(); err == nil {
		t.Fatal("starting an empty draft must be rejected")
	}
	if _, err := app.Undo(); err == nil {
		t.Fatal("undo without a running session must be rejected")
	}
	if _, err := app.AddChore(model.AddChoreRequest{Name: "", Count: 1}); err == nil {
		t.Fatal("a blank chore name must be rejected")
	}
}
