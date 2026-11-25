package model

// EventKind names a fact recorded in a session's event log.
type EventKind string

const (
	EventDone   EventKind = "done"
	EventUndo   EventKind = "undo"
	EventPause  EventKind = "pause"
	EventResume EventKind = "resume"
	EventEnd    EventKind = "end"
)

// SessionStatus is where a session is in its life. Draft, active and paused
// are derived; completed, time_up and abandoned are also stored as end_status.
type SessionStatus string

const (
	StatusDraft     SessionStatus = "draft"
	StatusActive    SessionStatus = "active"
	StatusPaused    SessionStatus = "paused"
	StatusCompleted SessionStatus = "completed"
	StatusTimeUp    SessionStatus = "time_up"
	StatusAbandoned SessionStatus = "abandoned"
)

// IsTerminal reports whether a session with this status has ended.
func (s SessionStatus) IsTerminal() bool {
	switch s {
	case StatusCompleted, StatusTimeUp, StatusAbandoned:
		return true
	}
	return false
}
