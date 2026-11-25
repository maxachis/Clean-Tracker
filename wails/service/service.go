package service

import (
	"errors"
	"sync"
	"time"

	"github.com/maxachis/clean-tracker/wails/store"
)

// RecentChoreNamesLimit caps how many past chore names are offered as quick picks.
const RecentChoreNamesLimit = 20

// eventPrecision is the resolution event and session timestamps are stored at.
const eventPrecision = time.Millisecond

// Rejections the service returns for requests that break the session rules.
var (
	ErrInvalidChoreName   = errors.New("chore name must not be empty")
	ErrInvalidChoreCount  = errors.New("chore count must be a whole number of 1 or more")
	ErrInvalidLength      = errors.New("session length must be above zero minutes")
	ErrInvalidMinAction   = errors.New("minimum per action must be a whole number of seconds from 1 to 600")
	ErrNotDraft           = errors.New("chores can only be edited while the session is a draft")
	ErrReorderMismatch    = errors.New("reorder must list every draft chore exactly once")
	ErrNoChores           = errors.New("add at least one chore before starting")
	ErrSessionRunning     = errors.New("another session is already running")
	ErrNoRunningSession   = errors.New("no session is running")
	ErrSessionEnded       = errors.New("the session has already ended")
	ErrSessionNotStarted  = errors.New("the session has not been started")
	ErrUnknownChore       = errors.New("chore is not part of the running session")
	ErrNothingOutstanding = errors.New("nothing is outstanding for that chore")
	ErrNothingToUndo      = errors.New("there is no logged action to undo")
	ErrAlreadyPaused      = errors.New("the session is already paused")
	ErrNotPaused          = errors.New("the session is not paused")
)

// Service holds the stateful session logic on top of the store. All public
// methods are serialized, since Wails calls them from concurrent goroutines.
type Service struct {
	Store *store.Store
	now   func() time.Time
	mu    sync.Mutex
}

// New returns a Service reading time from now, so tests control the clock.
func New(s *store.Store, now func() time.Time) *Service {
	return &Service{Store: s, now: now}
}

// clock returns the current time at stored precision.
func (s *Service) clock() time.Time {
	return s.now().UTC().Truncate(eventPrecision)
}
