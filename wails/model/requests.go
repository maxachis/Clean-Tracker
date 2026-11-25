package model

// AddChoreRequest adds a chore to the draft.
type AddChoreRequest struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// UpdateChoreRequest edits a draft chore's name and count.
type UpdateChoreRequest struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}
