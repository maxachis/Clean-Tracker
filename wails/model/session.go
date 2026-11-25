package model

import "time"

// Session is a timed cleaning session. StartedAt is nil while it is a draft;
// EndedAt and EndStatus are written once replay finds a terminal status.
type Session struct {
	ID            string         `json:"id"`
	LengthSeconds int            `json:"length_seconds"`
	CreatedAt     time.Time      `json:"created_at"`
	StartedAt     *time.Time     `json:"started_at,omitempty"`
	EndedAt       *time.Time     `json:"ended_at,omitempty"`
	EndStatus     *SessionStatus `json:"end_status,omitempty"`
}

// Chore is one named job in a session and how many times it must be done.
type Chore struct {
	ID          string `json:"id"`
	SessionID   string `json:"session_id"`
	Name        string `json:"name"`
	TargetCount int    `json:"target_count"`
	Position    int    `json:"position"`
}

// Event is one fact in a session's log. ID breaks ties between events that
// share a timestamp. ChoreID is set only for EventDone.
type Event struct {
	ID      int64     `json:"id"`
	Kind    EventKind `json:"kind"`
	ChoreID string    `json:"chore_id,omitempty"`
	At      time.Time `json:"at"`
}

// Draft is the session being assembled, with its chores in list order.
type Draft struct {
	Session Session `json:"session"`
	Chores  []Chore `json:"chores"`
}

// SessionView is a started session paired with its current snapshot.
type SessionView struct {
	SessionID string   `json:"session_id"`
	Session   Session  `json:"session"`
	Snapshot  Snapshot `json:"snapshot"`
}

// Summary is a started session with its final (or current) snapshot.
type Summary struct {
	Session  Session  `json:"session"`
	Snapshot Snapshot `json:"snapshot"`
}
