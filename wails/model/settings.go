package model

const (
	DefaultLastSessionMinutes = 30
	DefaultMissSound          = true
	SecondsPerMinute          = 60
)

// Settings are the user's saved preferences.
type Settings struct {
	LastSessionMinutes int  `json:"last_session_minutes"`
	MissSound          bool `json:"miss_sound"`
}

// UpdateSettingsRequest replaces the saved preferences.
type UpdateSettingsRequest struct {
	LastSessionMinutes int  `json:"last_session_minutes"`
	MissSound          bool `json:"miss_sound"`
}
