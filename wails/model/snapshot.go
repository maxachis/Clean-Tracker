package model

import "time"

// MinActionSeconds is the shortest countdown a reset point can set. Without
// it, repeated misses would shrink the countdown toward the session end
// without reaching it, giving infinitely many misses.
const MinActionSeconds = 1.0

// ChoreProgress is one chore's standing in a snapshot.
type ChoreProgress struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Target      int    `json:"target"`
	Done        int    `json:"done"`
	Outstanding int    `json:"outstanding"`
}

// Snapshot is everything the UI needs, computed by replaying the event log.
// EndedAt is set only when Status is terminal.
type Snapshot struct {
	Status             SessionStatus   `json:"status"`
	Chores             []ChoreProgress `json:"chores"`
	Done               int             `json:"done"`
	Outstanding        int             `json:"outstanding"`
	Missed             int             `json:"missed"`
	SessionSecondsLeft float64         `json:"session_seconds_left"`
	ActionSecondsLeft  float64         `json:"action_seconds_left"`
	ActionSecondsTotal float64         `json:"action_seconds_total"`
	Paused             bool            `json:"paused"`
	EndedAt            *time.Time      `json:"ended_at,omitempty"`
	AsOf               time.Time       `json:"as_of"`
}
