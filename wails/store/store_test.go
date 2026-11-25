package store

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/maxachis/clean-tracker/wails/model"
)

var testTime = time.Date(2026, 3, 4, 5, 6, 7, 123456789, time.UTC)

func openMemory(t *testing.T) *Store {
	t.Helper()
	s, err := Open(memoryDSN)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func mustInsertSession(t *testing.T, s *Store, id string, created time.Time) model.Session {
	t.Helper()
	sess := model.Session{ID: id, LengthSeconds: 600, CreatedAt: created}
	if err := s.InsertSession(sess); err != nil {
		t.Fatalf("insert session: %v", err)
	}
	return sess
}

func mustInsertChore(t *testing.T, s *Store, id, session, name string, position int) {
	t.Helper()
	c := model.Chore{ID: id, SessionID: session, Name: name, TargetCount: 2, Position: position}
	if err := s.InsertChore(c); err != nil {
		t.Fatalf("insert chore: %v", err)
	}
}

func TestOpen_SeedsDefaultSettings(t *testing.T) {
	s := openMemory(t)
	got, err := s.GetSettings()
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	want := model.Settings{LastSessionMinutes: model.DefaultLastSessionMinutes, MissSound: model.DefaultMissSound}
	if got != want {
		t.Fatalf("settings = %+v, want %+v", got, want)
	}
}

func TestUpdateSettings_PersistsFields(t *testing.T) {
	s := openMemory(t)
	if err := s.UpdateSettings(model.Settings{LastSessionMinutes: 12, MissSound: false}); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ := s.GetSettings()
	if got.LastSessionMinutes != 12 || got.MissSound {
		t.Fatalf("settings = %+v", got)
	}
}

func TestSession_RoundTripAndLifecycleQueries(t *testing.T) {
	// The draft is the unstarted session, the running one is started and not
	// ended; ending a session removes it from the running query.
	s := openMemory(t)
	mustInsertSession(t, s, "a", testTime)

	draft, err := s.GetDraft()
	if err != nil || draft.ID != "a" || !draft.CreatedAt.Equal(testTime) {
		t.Fatalf("draft = %+v, %v", draft, err)
	}
	if _, err := s.GetRunning(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("running before start: %v, want ErrNotFound", err)
	}

	if err := s.MarkSessionStarted("a", testTime.Add(time.Minute)); err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := s.GetDraft(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("draft after start: %v, want ErrNotFound", err)
	}
	running, err := s.GetRunning()
	if err != nil || running.StartedAt == nil || !running.StartedAt.Equal(testTime.Add(time.Minute)) {
		t.Fatalf("running = %+v, %v", running, err)
	}

	if err := s.MarkSessionEnded("a", testTime.Add(time.Hour), model.StatusTimeUp); err != nil {
		t.Fatalf("end: %v", err)
	}
	if _, err := s.GetRunning(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("running after end: %v, want ErrNotFound", err)
	}
	got, _ := s.GetSession("a")
	if got.EndStatus == nil || *got.EndStatus != model.StatusTimeUp || got.EndedAt == nil {
		t.Fatalf("ended session = %+v", got)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	s := openMemory(t)
	if _, err := s.GetSession("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestUpdateSessionLength_PersistsAndRejectsMissing(t *testing.T) {
	s := openMemory(t)
	mustInsertSession(t, s, "a", testTime)
	if err := s.UpdateSessionLength("a", 90); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ := s.GetSession("a")
	if got.LengthSeconds != 90 {
		t.Fatalf("length = %d", got.LengthSeconds)
	}
	if err := s.UpdateSessionLength("nope", 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing err = %v", err)
	}
}

func TestChores_CRUDAndReorder(t *testing.T) {
	s := openMemory(t)
	mustInsertSession(t, s, "a", testTime)
	mustInsertChore(t, s, "c1", "a", "One", 0)
	mustInsertChore(t, s, "c2", "a", "Two", 1)

	if err := s.SetChorePositions([]string{"c2", "c1"}); err != nil {
		t.Fatalf("reorder: %v", err)
	}
	if err := s.UpdateChore("c1", "Uno", 5); err != nil {
		t.Fatalf("update: %v", err)
	}
	chores, err := s.ListChores("a")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(chores) != 2 || chores[0].ID != "c2" || chores[1].Name != "Uno" || chores[1].TargetCount != 5 {
		t.Fatalf("chores = %+v", chores)
	}

	if err := s.DeleteChore("c2"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetChore("c2"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("get deleted: %v", err)
	}
}

func TestInsertChore_RejectsZeroCount(t *testing.T) {
	s := openMemory(t)
	mustInsertSession(t, s, "a", testTime)
	err := s.InsertChore(model.Chore{ID: "c", SessionID: "a", Name: "x", TargetCount: 0})
	if err == nil {
		t.Fatal("expected error for zero target count")
	}
}

func TestEvents_ListedInInsertionOrderWithChoreIDs(t *testing.T) {
	s := openMemory(t)
	mustInsertSession(t, s, "a", testTime)
	mustInsertChore(t, s, "c1", "a", "One", 0)
	if err := s.InsertEvent("a", model.EventDone, "c1", testTime); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := s.InsertEvent("a", model.EventPause, "", testTime); err != nil {
		t.Fatalf("insert: %v", err)
	}
	events, err := s.ListEvents("a")
	if err != nil || len(events) != 2 {
		t.Fatalf("events = %+v, %v", events, err)
	}
	if events[0].Kind != model.EventDone || events[0].ChoreID != "c1" || !events[0].At.Equal(testTime) {
		t.Fatalf("first event = %+v", events[0])
	}
	if events[1].Kind != model.EventPause || events[1].ChoreID != "" || events[1].ID <= events[0].ID {
		t.Fatalf("second event = %+v", events[1])
	}
}

func TestDeleteSession_CascadesToChoresAndEvents(t *testing.T) {
	// Foreign keys must be enforced on every pooled connection for cascades to work.
	s := openMemory(t)
	mustInsertSession(t, s, "a", testTime)
	mustInsertChore(t, s, "c1", "a", "One", 0)
	_ = s.InsertEvent("a", model.EventDone, "c1", testTime)
	if _, err := s.DB.Exec(`DELETE FROM sessions WHERE id = 'a'`); err != nil {
		t.Fatalf("delete: %v", err)
	}
	chores, _ := s.ListChores("a")
	events, _ := s.ListEvents("a")
	if len(chores) != 0 || len(events) != 0 {
		t.Fatalf("orphans: %d chores, %d events", len(chores), len(events))
	}
}

func TestRecentChoreNames_DistinctNewestFirstStartedOnly(t *testing.T) {
	s := openMemory(t)
	mustInsertSession(t, s, "old", testTime)
	mustInsertSession(t, s, "new", testTime.Add(time.Hour))
	mustInsertSession(t, s, "draft", testTime.Add(2*time.Hour))
	mustInsertChore(t, s, "o1", "old", "Dishes", 0)
	mustInsertChore(t, s, "o2", "old", "Vacuum", 1)
	mustInsertChore(t, s, "n1", "new", "Dishes", 0)
	mustInsertChore(t, s, "n2", "new", "Mop", 1)
	mustInsertChore(t, s, "d1", "draft", "Draft only", 0)
	for _, id := range []string{"old", "new"} {
		if err := s.MarkSessionStarted(id, testTime); err != nil {
			t.Fatalf("start: %v", err)
		}
	}

	got, err := s.RecentChoreNames(10)
	if err != nil {
		t.Fatalf("recent: %v", err)
	}
	want := []string{"Dishes", "Mop", "Vacuum"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("names = %v, want %v", got, want)
	}
	limited, _ := s.RecentChoreNames(1)
	if len(limited) != 1 {
		t.Fatalf("limit ignored: %v", limited)
	}
}

func TestOpen_FileBacked_PersistsAcrossReopen(t *testing.T) {
	// The in-memory tests skip the real driver path and the pragma DSN.
	path := filepath.Join(t.TempDir(), "ct.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	mustInsertSession(t, s, "a", testTime)
	mustInsertChore(t, s, "c1", "a", "One", 0)
	_ = s.InsertEvent("a", model.EventDone, "c1", testTime)
	_ = s.Close()

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()
	events, err := reopened.ListEvents("a")
	if err != nil || len(events) != 1 || !events[0].At.Equal(testTime) {
		t.Fatalf("events after reopen = %+v, %v", events, err)
	}
}

func TestDefaultDBPath_UsesConfigDir(t *testing.T) {
	cfg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfg)
	t.Setenv("HOME", cfg)
	path, err := DefaultDBPath()
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if filepath.Base(path) != dbFile || filepath.Base(filepath.Dir(path)) != dbSubdir {
		t.Fatalf("path = %s", path)
	}
}
