package model

const (
	DefaultLastSessionMinutes = 30
	DefaultMissSound          = true
	SecondsPerMinute          = 60

	// DefaultMinActionSeconds is the per-action minimum a fresh install uses.
	DefaultMinActionSeconds = 10
	// MinMinActionSeconds and MaxMinActionSeconds bound the setting. The lower
	// bound keeps repeated misses from shrinking the countdown toward the
	// session end without reaching it (infinitely many misses).
	MinMinActionSeconds = 1
	MaxMinActionSeconds = 600
)

// Settings are the user's saved preferences.
type Settings struct {
	LastSessionMinutes int  `json:"last_session_minutes"`
	MissSound          bool `json:"miss_sound"`
	MinActionSeconds   int  `json:"min_action_seconds"`
}

// UpdateSettingsRequest replaces the saved preferences.
type UpdateSettingsRequest struct {
	LastSessionMinutes int  `json:"last_session_minutes"`
	MissSound          bool `json:"miss_sound"`
	MinActionSeconds   int  `json:"min_action_seconds"`
}
