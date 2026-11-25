package service

import (
	"sort"
	"time"

	"github.com/maxachis/clean-tracker/wails/model"
)

// ReplayInput is every stored fact Replay needs, plus the current time.
// Events may arrive in any order; Replay orders them by time, then ID.
type ReplayInput struct {
	LengthSeconds int
	StartedAt     time.Time
	Chores        []model.Chore
	Events        []model.Event
	Now           time.Time
}

// Replay computes a session's full execution state from its stored facts
// (see SPEC.md, "Timer Model"). It is pure: no database, no clock.
func Replay(in ReplayInput) model.Snapshot {
	r := newReplayer(in)
	for _, e := range sortedEvents(in.Events) {
		if r.ended() {
			break
		}
		r.advanceTo(e.At)
		if r.ended() {
			break
		}
		r.apply(e)
	}
	r.advanceTo(in.Now)
	return r.snapshot(in.Now)
}

func sortedEvents(events []model.Event) []model.Event {
	sorted := append([]model.Event(nil), events...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if !sorted[i].At.Equal(sorted[j].At) {
			return sorted[i].At.Before(sorted[j].At)
		}
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}

// replayer is the mutable state of one replay. Times are active seconds
// (wall time since start, minus time spent paused) unless named wall.
type replayer struct {
	length      float64
	chores      []model.Chore
	outstanding map[string]int
	doneStack   []string // chore IDs of dones not yet undone, oldest first
	remaining   int

	active   float64
	wall     time.Time // wall time that `active` corresponds to
	paused   bool
	resetAt  float64
	deadline float64
	missed   int

	endStatus model.SessionStatus
	endedAt   time.Time
}

func newReplayer(in ReplayInput) *replayer {
	r := &replayer{
		length:      float64(in.LengthSeconds),
		chores:      in.Chores,
		outstanding: make(map[string]int, len(in.Chores)),
		wall:        in.StartedAt,
	}
	for _, c := range in.Chores {
		r.outstanding[c.ID] = c.TargetCount
		r.remaining += c.TargetCount
	}
	if r.remaining == 0 {
		r.finish(model.StatusCompleted, in.StartedAt)
		return r
	}
	r.reset()
	return r
}

func (r *replayer) ended() bool { return r.endStatus != "" }

func (r *replayer) finish(status model.SessionStatus, at time.Time) {
	r.endStatus = status
	r.endedAt = at
}

// reset starts a fresh countdown at the current active time from the
// current totals.
func (r *replayer) reset() {
	r.resetAt = r.active
	perAction := (r.length - r.active) / float64(r.remaining)
	if perAction < model.MinActionSeconds {
		perAction = model.MinActionSeconds
	}
	r.deadline = r.active + perAction
	if r.deadline > r.length {
		r.deadline = r.length
	}
}

// advanceTo moves the clocks forward to a wall time, counting each deadline
// passed along the way as a miss and ending the session if active time
// reaches the session length.
func (r *replayer) advanceTo(to time.Time) {
	if r.ended() || !to.After(r.wall) {
		return
	}
	if r.paused {
		r.wall = to
		return
	}
	target := r.active + to.Sub(r.wall).Seconds()
	wallAtLength := r.wall.Add(secondsToDuration(r.length - r.active))
	r.countMissesBefore(minFloat(target, r.length))
	if target >= r.length {
		r.finish(model.StatusTimeUp, wallAtLength)
		r.active = r.length
		return
	}
	r.active = target
	r.wall = to
}

// countMissesBefore records a miss for every deadline strictly before limit,
// each one starting a new countdown from the same totals.
func (r *replayer) countMissesBefore(limit float64) {
	for r.deadline < limit {
		r.missed++
		r.active = r.deadline
		r.reset()
	}
}

func (r *replayer) apply(e model.Event) {
	switch e.Kind {
	case model.EventDone:
		r.applyDone(e)
	case model.EventUndo:
		r.applyUndo()
	case model.EventPause:
		r.paused = true
	case model.EventResume:
		r.applyResume()
	case model.EventEnd:
		r.finish(model.StatusAbandoned, e.At)
	}
}

func (r *replayer) applyDone(e model.Event) {
	if r.outstanding[e.ChoreID] <= 0 {
		return
	}
	r.outstanding[e.ChoreID]--
	r.remaining--
	r.doneStack = append(r.doneStack, e.ChoreID)
	if r.remaining == 0 {
		r.finish(model.StatusCompleted, e.At)
		return
	}
	r.reset()
}

func (r *replayer) applyUndo() {
	last := len(r.doneStack) - 1
	if last < 0 {
		return
	}
	r.outstanding[r.doneStack[last]]++
	r.remaining++
	r.doneStack = r.doneStack[:last]
	r.reset()
}

func (r *replayer) applyResume() {
	if !r.paused {
		return
	}
	r.paused = false
	r.reset()
}

func (r *replayer) snapshot(now time.Time) model.Snapshot {
	snap := model.Snapshot{
		Status:             r.currentStatus(),
		Chores:             r.choreProgress(),
		Outstanding:        r.remaining,
		Missed:             r.missed,
		SessionSecondsLeft: r.length - r.active,
		Paused:             r.paused && !r.ended(),
		AsOf:               now,
	}
	for _, c := range snap.Chores {
		snap.Done += c.Done
	}
	if r.ended() {
		endedAt := r.endedAt
		snap.EndedAt = &endedAt
		return snap
	}
	snap.ActionSecondsLeft = r.deadline - r.active
	snap.ActionSecondsTotal = r.deadline - r.resetAt
	return snap
}

func (r *replayer) currentStatus() model.SessionStatus {
	switch {
	case r.ended():
		return r.endStatus
	case r.paused:
		return model.StatusPaused
	}
	return model.StatusActive
}

func (r *replayer) choreProgress() []model.ChoreProgress {
	progress := make([]model.ChoreProgress, 0, len(r.chores))
	for _, c := range r.chores {
		out := r.outstanding[c.ID]
		progress = append(progress, model.ChoreProgress{
			ID: c.ID, Name: c.Name, Target: c.TargetCount,
			Done: c.TargetCount - out, Outstanding: out,
		})
	}
	return progress
}

func secondsToDuration(sec float64) time.Duration {
	return time.Duration(sec * float64(time.Second))
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
