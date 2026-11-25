package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/maxachis/clean-tracker/wails/model"
	"github.com/maxachis/clean-tracker/wails/store"
)

// fakeClock lets tests control the time the service sees.
type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

func (c *fakeClock) Advance(seconds float64) { c.now = c.now.Add(secondsToDuration(seconds)) }

func newTestService(t *testing.T) (*Service, *fakeClock) {
	t.Helper()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	clock := &fakeClock{now: replayBase}
	return New(st, clock.Now), clock
}

func mustAdd(t *testing.T, svc *Service, name string, count int) model.Draft {
	t.Helper()
	d, err := svc.AddChore(model.AddChoreRequest{Name: name, Count: count})
	if err != nil {
		t.Fatalf("add chore %q: %v", name, err)
	}
	return d
}

// startStandard starts a 10-minute session with chores Boxes x3 and Counters x2
// and returns their IDs.
func startStandard(t *testing.T, svc *Service) (boxes, counters string) {
	t.Helper()
	if _, err := svc.SetDraftLength(10); err != nil {
		t.Fatalf("set length: %v", err)
	}
	mustAdd(t, svc, "Boxes", 3)
	d := mustAdd(t, svc, "Counters", 2)
	if _, err := svc.StartSession(); err != nil {
		t.Fatalf("start: %v", err)
	}
	return d.Chores[0].ID, d.Chores[1].ID
}

func wantErr(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("err = %v, want %v", err, target)
	}
}

func TestGetDraft_CreatesOnceSeededFromLastSessionMinutes(t *testing.T) {
	svc, _ := newTestService(t)
	if _, err := svc.UpdateSettings(model.UpdateSettingsRequest{LastSessionMinutes: 45, MissSound: true}); err != nil {
		t.Fatalf("settings: %v", err)
	}
	first, err := svc.GetDraft()
	if err != nil {
		t.Fatalf("draft: %v", err)
	}
	second, _ := svc.GetDraft()
	if first.Session.ID != second.Session.ID {
		t.Fatal("a second GetDraft must return the same draft, not create another")
	}
	if first.Session.LengthSeconds != 45*model.SecondsPerMinute {
		t.Fatalf("length = %d, want 45 minutes", first.Session.LengthSeconds)
	}
}

func TestAddChore_ValidatesAndTrimsAndAppends(t *testing.T) {
	svc, _ := newTestService(t)
	tests := []struct {
		name string
		req  model.AddChoreRequest
		want error
	}{
		{"empty name is rejected", model.AddChoreRequest{Name: "", Count: 1}, ErrInvalidChoreName},
		{"whitespace-only name is rejected", model.AddChoreRequest{Name: "   ", Count: 1}, ErrInvalidChoreName},
		{"count below one is rejected", model.AddChoreRequest{Name: "Dishes", Count: 0}, ErrInvalidChoreCount},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.AddChore(tc.req)
			wantErr(t, err, tc.want)
		})
	}
	mustAdd(t, svc, "  Dishes  ", 2)
	d := mustAdd(t, svc, "Mop", 1)
	if len(d.Chores) != 2 || d.Chores[0].Name != "Dishes" || d.Chores[1].Position <= d.Chores[0].Position {
		t.Fatalf("chores = %+v", d.Chores)
	}
}

func TestDraftChoreEditing_UpdateRemoveReorder(t *testing.T) {
	svc, _ := newTestService(t)
	mustAdd(t, svc, "One", 1)
	d := mustAdd(t, svc, "Two", 1)
	one, two := d.Chores[0].ID, d.Chores[1].ID

	d, err := svc.UpdateChore(model.UpdateChoreRequest{ID: one, Name: " Uno ", Count: 4})
	if err != nil || d.Chores[0].Name != "Uno" || d.Chores[0].TargetCount != 4 {
		t.Fatalf("update = %+v, %v", d.Chores, err)
	}
	if _, err := svc.UpdateChore(model.UpdateChoreRequest{ID: one, Name: "", Count: 1}); !errors.Is(err, ErrInvalidChoreName) {
		t.Fatalf("blank update err = %v", err)
	}
	d, err = svc.ReorderChores([]string{two, one})
	if err != nil || d.Chores[0].ID != two {
		t.Fatalf("reorder = %+v, %v", d.Chores, err)
	}
	if _, err := svc.ReorderChores([]string{two}); !errors.Is(err, ErrReorderMismatch) {
		t.Fatalf("partial reorder err = %v", err)
	}
	d, err = svc.RemoveChore(two)
	if err != nil || len(d.Chores) != 1 || d.Chores[0].ID != one {
		t.Fatalf("remove = %+v, %v", d.Chores, err)
	}
}

func TestChoresAreEditableOnlyOnADraft(t *testing.T) {
	svc, _ := newTestService(t)
	boxes, _ := startStandard(t, svc)
	if _, err := svc.UpdateChore(model.UpdateChoreRequest{ID: boxes, Name: "X", Count: 1}); !errors.Is(err, ErrNotDraft) {
		t.Fatalf("update err = %v", err)
	}
	if _, err := svc.RemoveChore(boxes); !errors.Is(err, ErrNotDraft) {
		t.Fatalf("remove err = %v", err)
	}
}

func TestSetDraftLength_StoresMinutesAsSeconds(t *testing.T) {
	svc, _ := newTestService(t)
	d, err := svc.SetDraftLength(20)
	if err != nil || d.Session.LengthSeconds != 20*model.SecondsPerMinute {
		t.Fatalf("length = %+v, %v", d.Session, err)
	}
	if _, err := svc.SetDraftLength(-1); !errors.Is(err, ErrInvalidLength) {
		t.Fatalf("negative err = %v", err)
	}
}

func TestStartSession_RequiresChoresAndPositiveLength(t *testing.T) {
	svc, _ := newTestService(t)
	if _, err := svc.StartSession(); !errors.Is(err, ErrNoChores) {
		t.Fatalf("no chores err = %v", err)
	}
	mustAdd(t, svc, "Dishes", 1)
	if _, err := svc.SetDraftLength(0); err != nil {
		t.Fatalf("set zero: %v", err)
	}
	if _, err := svc.StartSession(); !errors.Is(err, ErrInvalidLength) {
		t.Fatalf("zero length err = %v", err)
	}
}

func TestStartSession_SavesLastSessionMinutesAndReturnsFreshSnapshot(t *testing.T) {
	svc, _ := newTestService(t)
	startStandard(t, svc)
	settings, _ := svc.GetSettings()
	if settings.LastSessionMinutes != 10 {
		t.Fatalf("last session minutes = %d, want 10", settings.LastSessionMinutes)
	}
	view, err := svc.GetRunningSession()
	if err != nil || view == nil {
		t.Fatalf("running = %v, %v", view, err)
	}
	assertSnapshot(t, view.Snapshot, wantSnapshot{
		status: model.StatusActive, outstanding: 5, sessionLeft: 600, actionLeft: 120, actionTotal: 120,
	})
}

func TestStartSession_RejectedWhileAnotherSessionIsRunning(t *testing.T) {
	svc, _ := newTestService(t)
	startStandard(t, svc)
	mustAdd(t, svc, "Another", 1)
	if _, err := svc.StartSession(); !errors.Is(err, ErrSessionRunning) {
		t.Fatalf("err = %v, want ErrSessionRunning", err)
	}
}

func TestGetRunningSession_NilWhenNothingRuns(t *testing.T) {
	svc, _ := newTestService(t)
	view, err := svc.GetRunningSession()
	if err != nil || view != nil {
		t.Fatalf("view = %v, err = %v", view, err)
	}
}

func TestLogAction_ReducesOutstandingAndResetsCountdown(t *testing.T) {
	svc, clock := newTestService(t)
	boxes, _ := startStandard(t, svc)
	clock.Advance(60)
	snap, err := svc.LogAction(boxes)
	if err != nil {
		t.Fatalf("log: %v", err)
	}
	assertSnapshot(t, snap, wantSnapshot{
		status: model.StatusActive, done: 1, outstanding: 4, sessionLeft: 540, actionLeft: 135, actionTotal: 135,
	})
}

func TestLogAction_RejectsChoreWithNothingOutstandingOrUnknown(t *testing.T) {
	svc, _ := newTestService(t)
	_, counters := startStandard(t, svc)
	for i := 0; i < 2; i++ {
		if _, err := svc.LogAction(counters); err != nil {
			t.Fatalf("log %d: %v", i, err)
		}
	}
	if _, err := svc.LogAction(counters); !errors.Is(err, ErrNothingOutstanding) {
		t.Fatalf("finished chore err = %v", err)
	}
	if _, err := svc.LogAction("nope"); !errors.Is(err, ErrUnknownChore) {
		t.Fatalf("unknown chore err = %v", err)
	}
}

func TestLogAction_RequiresARunningSession(t *testing.T) {
	svc, _ := newTestService(t)
	if _, err := svc.LogAction("x"); !errors.Is(err, ErrNoRunningSession) {
		t.Fatalf("err = %v", err)
	}
}

func TestCompletingLastActionPersistsEndOnce(t *testing.T) {
	// The summary and history read ended_at/end_status, so a terminal replay
	// must write them, with the completion time rather than a later read time.
	svc, clock := newTestService(t)
	mustAdd(t, svc, "Only", 1)
	d, _ := svc.GetDraft()
	if _, err := svc.StartSession(); err != nil {
		t.Fatalf("start: %v", err)
	}
	clock.Advance(30)
	snap, err := svc.LogAction(d.Chores[0].ID)
	if err != nil || snap.Status != model.StatusCompleted {
		t.Fatalf("snap = %+v, %v", snap, err)
	}
	clock.Advance(500)
	if _, err := svc.GetSummary(d.Session.ID); err != nil {
		t.Fatalf("summary: %v", err)
	}
	stored, _ := svc.Store.GetSession(d.Session.ID)
	if stored.EndStatus == nil || *stored.EndStatus != model.StatusCompleted {
		t.Fatalf("end status = %v", stored.EndStatus)
	}
	if stored.EndedAt == nil || !stored.EndedAt.Equal(replayBase.Add(30*time.Second)) {
		t.Fatalf("ended at = %v, want completion time", stored.EndedAt)
	}
}

func TestTimeUpWhileClosedIsDetectedPersistedAndReportedOnce(t *testing.T) {
	// Reopening after the session clock ran out ends the session as time_up;
	// the first look reports it, later looks find no running session.
	svc, clock := newTestService(t)
	startStandard(t, svc)
	clock.Advance(5000)
	view, err := svc.GetRunningSession()
	if err != nil || view == nil || view.Snapshot.Status != model.StatusTimeUp {
		t.Fatalf("view = %+v, %v", view, err)
	}
	again, err := svc.GetRunningSession()
	if err != nil || again != nil {
		t.Fatalf("second look = %+v, %v", again, err)
	}
	summary, err := svc.GetSummary(view.SessionID)
	if err != nil || summary.Snapshot.Status != model.StatusTimeUp || summary.Session.EndedAt == nil {
		t.Fatalf("summary = %+v, %v", summary, err)
	}
	assertClose(t, "ended at", summary.Session.EndedAt.Sub(replayBase).Seconds(), 600)
}

func TestMissesWhileAppWasClosedAreCountedByANewService(t *testing.T) {
	// A fresh Service over the same store stands in for reopening the app.
	svc, clock := newTestService(t)
	startStandard(t, svc)
	clock.Advance(400)
	reopened := New(svc.Store, clock.Now)
	view, err := reopened.GetRunningSession()
	if err != nil || view == nil {
		t.Fatalf("view = %v, %v", view, err)
	}
	if view.Snapshot.Missed != 4 {
		t.Fatalf("missed = %d, want 4", view.Snapshot.Missed)
	}
}

func TestLogActionAfterTimeRanOutIsRejected(t *testing.T) {
	svc, clock := newTestService(t)
	boxes, _ := startStandard(t, svc)
	clock.Advance(5000)
	if _, err := svc.LogAction(boxes); !errors.Is(err, ErrSessionEnded) {
		t.Fatalf("err = %v, want ErrSessionEnded", err)
	}
}

func TestUndo_RestoresLastActionAndRejectsWhenNothingToUndo(t *testing.T) {
	svc, clock := newTestService(t)
	boxes, _ := startStandard(t, svc)
	if _, err := svc.Undo(); !errors.Is(err, ErrNothingToUndo) {
		t.Fatalf("empty undo err = %v", err)
	}
	clock.Advance(60)
	_, _ = svc.LogAction(boxes)
	clock.Advance(30)
	snap, err := svc.Undo()
	if err != nil {
		t.Fatalf("undo: %v", err)
	}
	assertSnapshot(t, snap, wantSnapshot{
		status: model.StatusActive, outstanding: 5, sessionLeft: 510, actionLeft: 102, actionTotal: 102,
	})
	if _, err := svc.Undo(); !errors.Is(err, ErrNothingToUndo) {
		t.Fatalf("second undo err = %v", err)
	}
}

func TestPauseResume_RejectDoubleAndFreezeTime(t *testing.T) {
	svc, clock := newTestService(t)
	startStandard(t, svc)
	if _, err := svc.Resume(); !errors.Is(err, ErrNotPaused) {
		t.Fatalf("resume while running err = %v", err)
	}
	clock.Advance(30)
	snap, err := svc.Pause()
	if err != nil || !snap.Paused {
		t.Fatalf("pause = %+v, %v", snap, err)
	}
	if _, err := svc.Pause(); !errors.Is(err, ErrAlreadyPaused) {
		t.Fatalf("double pause err = %v", err)
	}
	clock.Advance(9000)
	snap, err = svc.Resume()
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	assertSnapshot(t, snap, wantSnapshot{
		status: model.StatusActive, outstanding: 5, sessionLeft: 570, actionLeft: 114, actionTotal: 114,
	})
}

func TestEndSession_AbandonsAndPersists(t *testing.T) {
	svc, clock := newTestService(t)
	startStandard(t, svc)
	view, _ := svc.GetRunningSession()
	clock.Advance(50)
	snap, err := svc.EndSession()
	if err != nil || snap.Status != model.StatusAbandoned {
		t.Fatalf("end = %+v, %v", snap, err)
	}
	if again, _ := svc.GetRunningSession(); again != nil {
		t.Fatal("an ended session must no longer be running")
	}
	stored, _ := svc.Store.GetSession(view.SessionID)
	if stored.EndStatus == nil || *stored.EndStatus != model.StatusAbandoned {
		t.Fatalf("end status = %v", stored.EndStatus)
	}
	if _, err := svc.EndSession(); !errors.Is(err, ErrNoRunningSession) {
		t.Fatalf("end twice err = %v", err)
	}
}

func TestGetSummary_RejectsDraftsAndUnknownSessions(t *testing.T) {
	svc, _ := newTestService(t)
	d, _ := svc.GetDraft()
	if _, err := svc.GetSummary(d.Session.ID); !errors.Is(err, ErrSessionNotStarted) {
		t.Fatalf("draft summary err = %v", err)
	}
	if _, err := svc.GetSummary("nope"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("unknown summary err = %v", err)
	}
}

func TestRepeatSession_CopiesListAndReplacesExistingDraftChores(t *testing.T) {
	svc, _ := newTestService(t)
	startStandard(t, svc)
	view, _ := svc.GetRunningSession()
	if _, err := svc.EndSession(); err != nil {
		t.Fatalf("end: %v", err)
	}
	stale, _ := svc.GetDraft()
	mustAdd(t, svc, "Stale chore", 9)
	_, _ = svc.SetDraftLength(5)

	d, err := svc.RepeatSession(view.SessionID)
	if err != nil {
		t.Fatalf("repeat: %v", err)
	}
	if d.Session.ID != stale.Session.ID {
		t.Fatal("repeat must reuse the existing draft, keeping one draft at a time")
	}
	if d.Session.LengthSeconds != 600 {
		t.Fatalf("length = %d, want the source's 600", d.Session.LengthSeconds)
	}
	if len(d.Chores) != 2 || d.Chores[0].Name != "Boxes" || d.Chores[0].TargetCount != 3 ||
		d.Chores[1].Name != "Counters" || d.Chores[1].TargetCount != 2 {
		t.Fatalf("chores = %+v", d.Chores)
	}
}

func TestRecentChoreNames_CappedAtLimitNewestFirst(t *testing.T) {
	svc, clock := newTestService(t)
	total := RecentChoreNamesLimit + 5
	for i := 0; i < total; i++ {
		mustAdd(t, svc, fmt.Sprintf("Chore %02d", i), 1)
		if _, err := svc.StartSession(); err != nil {
			t.Fatalf("start %d: %v", i, err)
		}
		if _, err := svc.EndSession(); err != nil {
			t.Fatalf("end %d: %v", i, err)
		}
		clock.Advance(1)
	}
	names, err := svc.RecentChoreNames()
	if err != nil || len(names) != RecentChoreNamesLimit {
		t.Fatalf("names = %v, %v", names, err)
	}
	if names[0] != fmt.Sprintf("Chore %02d", total-1) {
		t.Fatalf("first = %q, want the newest", names[0])
	}
}

func TestUpdateSettings_ValidatesAndPersists(t *testing.T) {
	svc, _ := newTestService(t)
	if _, err := svc.UpdateSettings(model.UpdateSettingsRequest{LastSessionMinutes: 0}); !errors.Is(err, ErrInvalidLength) {
		t.Fatalf("zero minutes err = %v", err)
	}
	got, err := svc.UpdateSettings(model.UpdateSettingsRequest{LastSessionMinutes: 15, MissSound: false})
	if err != nil || got.LastSessionMinutes != 15 || got.MissSound {
		t.Fatalf("settings = %+v, %v", got, err)
	}
	reread, _ := svc.GetSettings()
	if reread != got {
		t.Fatalf("reread = %+v, want %+v", reread, got)
	}
}

func TestFullSession_FileBackedSurvivesReopen(t *testing.T) {
	// The real driver and file path: a session started in one process is
	// still running, with misses counted, after reopening the database.
	path := filepath.Join(t.TempDir(), "ct.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	clock := &fakeClock{now: replayBase}
	svc := New(st, clock.Now)
	boxes, _ := startStandard(t, svc)
	clock.Advance(10)
	if _, err := svc.LogAction(boxes); err != nil {
		t.Fatalf("log: %v", err)
	}
	_ = st.Close()

	reopened, err := store.Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()
	clock.Advance(200)
	view, err := New(reopened, clock.Now).GetRunningSession()
	if err != nil || view == nil {
		t.Fatalf("view = %v, %v", view, err)
	}
	if view.Snapshot.Done != 1 || view.Snapshot.Missed != 1 {
		t.Fatalf("snapshot = %+v, want 1 done and 1 miss", view.Snapshot)
	}
}
