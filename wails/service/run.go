package service

import (
	"errors"
	"fmt"

	"github.com/maxachis/clean-tracker/wails/model"
	"github.com/maxachis/clean-tracker/wails/store"
)

// StartSession starts the draft: it needs at least one chore and a length
// above zero, and no other session may be running.
func (s *Service) StartSession() (model.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireNoLiveSession(); err != nil {
		return model.Snapshot{}, err
	}
	d, err := s.draft()
	if err != nil {
		return model.Snapshot{}, err
	}
	if err := validateStartable(d); err != nil {
		return model.Snapshot{}, err
	}
	if err := s.Store.MarkSessionStarted(d.Session.ID, s.clock()); err != nil {
		return model.Snapshot{}, err
	}
	if err := s.rememberLength(d.Session.LengthSeconds); err != nil {
		return model.Snapshot{}, err
	}
	started, err := s.Store.GetSession(d.Session.ID)
	if err != nil {
		return model.Snapshot{}, err
	}
	return s.settle(started)
}

func validateStartable(d model.Draft) error {
	if len(d.Chores) == 0 {
		return ErrNoChores
	}
	if d.Session.LengthSeconds <= 0 {
		return ErrInvalidLength
	}
	return nil
}

func (s *Service) rememberLength(lengthSeconds int) error {
	settings, err := s.Store.GetSettings()
	if err != nil {
		return err
	}
	settings.LastSessionMinutes = lengthSeconds / model.SecondsPerMinute
	return s.Store.UpdateSettings(settings)
}

// requireNoLiveSession fails if a started session is still running. A session
// whose time ran out unseen is settled (ended) and does not block.
func (s *Service) requireNoLiveSession() error {
	sess, err := s.Store.GetRunning()
	if errors.Is(err, store.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	snap, err := s.settle(sess)
	if err != nil {
		return err
	}
	if snap.Status.IsTerminal() {
		return nil
	}
	return ErrSessionRunning
}

// GetRunningSession returns the running session and its snapshot, or nil
// when none is running. A session whose end was found only now (time ran out
// or it completed while the app was closed) is returned once with its
// terminal snapshot; after that it is no longer running.
func (s *Service) GetRunningSession() (*model.SessionView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, err := s.Store.GetRunning()
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	snap, err := s.settle(sess)
	if err != nil {
		return nil, err
	}
	return &model.SessionView{SessionID: sess.ID, Session: sess, Snapshot: snap}, nil
}

// settle replays a started session as of now and, if that reveals it has
// ended, writes ended_at and end_status (once).
func (s *Service) settle(sess model.Session) (model.Snapshot, error) {
	snap, err := s.replay(sess)
	if err != nil {
		return model.Snapshot{}, err
	}
	if snap.Status.IsTerminal() && sess.EndStatus == nil {
		if err := s.Store.MarkSessionEnded(sess.ID, *snap.EndedAt, snap.Status); err != nil {
			return model.Snapshot{}, err
		}
	}
	return snap, nil
}

func (s *Service) replay(sess model.Session) (model.Snapshot, error) {
	if sess.StartedAt == nil {
		return model.Snapshot{}, ErrSessionNotStarted
	}
	chores, err := s.Store.ListChores(sess.ID)
	if err != nil {
		return model.Snapshot{}, err
	}
	events, err := s.Store.ListEvents(sess.ID)
	if err != nil {
		return model.Snapshot{}, err
	}
	return Replay(ReplayInput{
		LengthSeconds: sess.LengthSeconds,
		StartedAt:     *sess.StartedAt,
		Chores:        chores,
		Events:        events,
		Now:           s.clock(),
	}), nil
}

// GetSummary returns a started session with its current or final snapshot.
func (s *Service) GetSummary(sessionID string) (model.Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, err := s.Store.GetSession(sessionID)
	if err != nil {
		return model.Summary{}, err
	}
	snap, err := s.settle(sess)
	if err != nil {
		return model.Summary{}, err
	}
	if sess.EndStatus == nil && snap.Status.IsTerminal() {
		sess, err = s.Store.GetSession(sessionID)
		if err != nil {
			return model.Summary{}, err
		}
	}
	return model.Summary{Session: sess, Snapshot: snap}, nil
}

// command describes one event the running session may be asked to record.
type command struct {
	kind    model.EventKind
	choreID string
	check   func(model.Snapshot) error
}

// LogAction records one done action for a chore.
func (s *Service) LogAction(choreID string) (model.Snapshot, error) {
	return s.run(command{kind: model.EventDone, choreID: choreID, check: requireOutstanding(choreID)})
}

func requireOutstanding(choreID string) func(model.Snapshot) error {
	return func(snap model.Snapshot) error {
		for _, c := range snap.Chores {
			if c.ID != choreID {
				continue
			}
			if c.Outstanding == 0 {
				return ErrNothingOutstanding
			}
			return nil
		}
		return ErrUnknownChore
	}
}

// Undo reverses the most recent logged action that is not already undone.
func (s *Service) Undo() (model.Snapshot, error) {
	return s.run(command{kind: model.EventUndo, check: func(snap model.Snapshot) error {
		if snap.Done == 0 {
			return ErrNothingToUndo
		}
		return nil
	}})
}

func (s *Service) Pause() (model.Snapshot, error) {
	return s.run(command{kind: model.EventPause, check: func(snap model.Snapshot) error {
		if snap.Paused {
			return ErrAlreadyPaused
		}
		return nil
	}})
}

func (s *Service) Resume() (model.Snapshot, error) {
	return s.run(command{kind: model.EventResume, check: func(snap model.Snapshot) error {
		if !snap.Paused {
			return ErrNotPaused
		}
		return nil
	}})
}

// EndSession abandons the running session.
func (s *Service) EndSession() (model.Snapshot, error) {
	return s.run(command{kind: model.EventEnd})
}

// run validates a command against the running session's current state,
// appends its event, and returns the resulting snapshot.
func (s *Service) run(cmd command) (model.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, err := s.Store.GetRunning()
	if errors.Is(err, store.ErrNotFound) {
		return model.Snapshot{}, ErrNoRunningSession
	}
	if err != nil {
		return model.Snapshot{}, err
	}
	before, err := s.settle(sess)
	if err != nil {
		return model.Snapshot{}, err
	}
	if before.Status.IsTerminal() {
		return model.Snapshot{}, ErrSessionEnded
	}
	if cmd.check != nil {
		if err := cmd.check(before); err != nil {
			return model.Snapshot{}, err
		}
	}
	if err := s.Store.InsertEvent(sess.ID, cmd.kind, cmd.choreID, s.clock()); err != nil {
		return model.Snapshot{}, fmt.Errorf("record %s: %w", cmd.kind, err)
	}
	return s.settle(sess)
}
