package service

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/maxachis/clean-tracker/wails/model"
	"github.com/maxachis/clean-tracker/wails/store"
)

// GetDraft returns the draft session with its chores, creating an empty one
// (seeded with the last used length) when none exists.
func (s *Service) GetDraft() (model.Draft, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.draft()
}

func (s *Service) draft() (model.Draft, error) {
	sess, err := s.Store.GetDraft()
	if errors.Is(err, store.ErrNotFound) {
		sess, err = s.createDraft()
	}
	if err != nil {
		return model.Draft{}, err
	}
	return s.loadDraft(sess)
}

func (s *Service) createDraft() (model.Session, error) {
	settings, err := s.Store.GetSettings()
	if err != nil {
		return model.Session{}, fmt.Errorf("seed draft length: %w", err)
	}
	sess := model.Session{
		ID:            uuid.NewString(),
		LengthSeconds: settings.LastSessionMinutes * model.SecondsPerMinute,
		CreatedAt:     s.clock(),
	}
	if err := s.Store.InsertSession(sess); err != nil {
		return model.Session{}, err
	}
	return sess, nil
}

func (s *Service) loadDraft(sess model.Session) (model.Draft, error) {
	chores, err := s.Store.ListChores(sess.ID)
	if err != nil {
		return model.Draft{}, err
	}
	return model.Draft{Session: sess, Chores: chores}, nil
}

func validateChore(name string, count int) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", ErrInvalidChoreName
	}
	if count < 1 {
		return "", ErrInvalidChoreCount
	}
	return trimmed, nil
}

// AddChore appends a chore to the draft.
func (s *Service) AddChore(req model.AddChoreRequest) (model.Draft, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name, err := validateChore(req.Name, req.Count)
	if err != nil {
		return model.Draft{}, err
	}
	d, err := s.draft()
	if err != nil {
		return model.Draft{}, err
	}
	chore := model.Chore{
		ID: uuid.NewString(), SessionID: d.Session.ID,
		Name: name, TargetCount: req.Count, Position: nextPosition(d.Chores),
	}
	if err := s.Store.InsertChore(chore); err != nil {
		return model.Draft{}, err
	}
	return s.loadDraft(d.Session)
}

func nextPosition(chores []model.Chore) int {
	next := 0
	for _, c := range chores {
		next = max(next, c.Position+1)
	}
	return next
}

// UpdateChore edits a draft chore's name and count.
func (s *Service) UpdateChore(req model.UpdateChoreRequest) (model.Draft, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name, err := validateChore(req.Name, req.Count)
	if err != nil {
		return model.Draft{}, err
	}
	sess, err := s.draftSessionOfChore(req.ID)
	if err != nil {
		return model.Draft{}, err
	}
	if err := s.Store.UpdateChore(req.ID, name, req.Count); err != nil {
		return model.Draft{}, err
	}
	return s.loadDraft(sess)
}

// RemoveChore deletes a chore from the draft.
func (s *Service) RemoveChore(id string) (model.Draft, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, err := s.draftSessionOfChore(id)
	if err != nil {
		return model.Draft{}, err
	}
	if err := s.Store.DeleteChore(id); err != nil {
		return model.Draft{}, err
	}
	return s.loadDraft(sess)
}

// draftSessionOfChore returns the chore's session, or ErrNotDraft if that
// session has already started.
func (s *Service) draftSessionOfChore(choreID string) (model.Session, error) {
	chore, err := s.Store.GetChore(choreID)
	if err != nil {
		return model.Session{}, err
	}
	sess, err := s.Store.GetSession(chore.SessionID)
	if err != nil {
		return model.Session{}, err
	}
	if sess.StartedAt != nil {
		return model.Session{}, ErrNotDraft
	}
	return sess, nil
}

// ReorderChores sets the draft's chore order. ids must name every draft chore once.
func (s *Service) ReorderChores(ids []string) (model.Draft, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, err := s.draft()
	if err != nil {
		return model.Draft{}, err
	}
	if !sameIDs(d.Chores, ids) {
		return model.Draft{}, ErrReorderMismatch
	}
	if err := s.Store.SetChorePositions(ids); err != nil {
		return model.Draft{}, err
	}
	return s.loadDraft(d.Session)
}

func sameIDs(chores []model.Chore, ids []string) bool {
	if len(chores) != len(ids) {
		return false
	}
	have := make([]string, 0, len(chores))
	for _, c := range chores {
		have = append(have, c.ID)
	}
	want := slices.Clone(ids)
	slices.Sort(have)
	slices.Sort(want)
	return slices.Equal(have, want)
}

// SetDraftLength sets the draft's session length in whole minutes. Zero is
// allowed while editing; StartSession rejects it.
func (s *Service) SetDraftLength(minutes int) (model.Draft, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if minutes < 0 {
		return model.Draft{}, ErrInvalidLength
	}
	d, err := s.draft()
	if err != nil {
		return model.Draft{}, err
	}
	if err := s.Store.UpdateSessionLength(d.Session.ID, minutes*model.SecondsPerMinute); err != nil {
		return model.Draft{}, err
	}
	d.Session.LengthSeconds = minutes * model.SecondsPerMinute
	return d, nil
}

// RecentChoreNames returns distinct chore names from past sessions, newest first.
func (s *Service) RecentChoreNames() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Store.RecentChoreNames(RecentChoreNamesLimit)
}

// RepeatSession copies a past session's chores, order and length into the
// draft, replacing any chores the draft already has.
func (s *Service) RepeatSession(sessionID string) (model.Draft, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	source, err := s.Store.GetSession(sessionID)
	if err != nil {
		return model.Draft{}, err
	}
	sourceChores, err := s.Store.ListChores(sessionID)
	if err != nil {
		return model.Draft{}, err
	}
	d, err := s.draft()
	if err != nil {
		return model.Draft{}, err
	}
	if d.Session.ID == sessionID {
		return d, nil
	}
	if err := s.replaceDraftContents(d.Session.ID, source.LengthSeconds, sourceChores); err != nil {
		return model.Draft{}, err
	}
	d.Session.LengthSeconds = source.LengthSeconds
	return s.loadDraft(d.Session)
}

func (s *Service) replaceDraftContents(draftID string, lengthSeconds int, chores []model.Chore) error {
	if err := s.Store.DeleteChoresForSession(draftID); err != nil {
		return err
	}
	if err := s.Store.UpdateSessionLength(draftID, lengthSeconds); err != nil {
		return err
	}
	for position, c := range chores {
		copied := model.Chore{
			ID: uuid.NewString(), SessionID: draftID,
			Name: c.Name, TargetCount: c.TargetCount, Position: position,
		}
		if err := s.Store.InsertChore(copied); err != nil {
			return err
		}
	}
	return nil
}
