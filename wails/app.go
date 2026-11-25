package main

import (
	"context"

	"github.com/maxachis/clean-tracker/wails/model"
	"github.com/maxachis/clean-tracker/wails/service"
)

// App is the root object exposed to the frontend. Every exported method
// becomes a callable IPC endpoint via `window.go.main.App.<Method>`. Keep
// them thin: delegate to the service and return typed structs.
type App struct {
	ctx     context.Context
	service *service.Service
}

func NewApp(svc *service.Service) *App {
	return &App{service: svc}
}

// startup is invoked by Wails when the window is ready.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetDraft returns the draft session, creating an empty one if none exists.
func (a *App) GetDraft() (model.Draft, error) {
	return a.service.GetDraft()
}

func (a *App) AddChore(req model.AddChoreRequest) (model.Draft, error) {
	return a.service.AddChore(req)
}

func (a *App) UpdateChore(req model.UpdateChoreRequest) (model.Draft, error) {
	return a.service.UpdateChore(req)
}

func (a *App) RemoveChore(id string) (model.Draft, error) {
	return a.service.RemoveChore(id)
}

func (a *App) ReorderChores(ids []string) (model.Draft, error) {
	return a.service.ReorderChores(ids)
}

func (a *App) SetDraftLength(minutes int) (model.Draft, error) {
	return a.service.SetDraftLength(minutes)
}

// RecentChoreNames returns distinct chore names from past sessions, newest first.
func (a *App) RecentChoreNames() ([]string, error) {
	return a.service.RecentChoreNames()
}

func (a *App) StartSession() (model.Snapshot, error) {
	return a.service.StartSession()
}

// GetRunningSession returns the running session, or nil when none is running.
func (a *App) GetRunningSession() (*model.SessionView, error) {
	return a.service.GetRunningSession()
}

func (a *App) LogAction(choreID string) (model.Snapshot, error) {
	return a.service.LogAction(choreID)
}

func (a *App) Undo() (model.Snapshot, error) {
	return a.service.Undo()
}

func (a *App) Pause() (model.Snapshot, error) {
	return a.service.Pause()
}

func (a *App) Resume() (model.Snapshot, error) {
	return a.service.Resume()
}

func (a *App) EndSession() (model.Snapshot, error) {
	return a.service.EndSession()
}

func (a *App) GetSummary(sessionID string) (model.Summary, error) {
	return a.service.GetSummary(sessionID)
}

// RepeatSession copies a past session's chores and length into the draft.
func (a *App) RepeatSession(sessionID string) (model.Draft, error) {
	return a.service.RepeatSession(sessionID)
}

func (a *App) GetSettings() (model.Settings, error) {
	return a.service.GetSettings()
}

func (a *App) UpdateSettings(req model.UpdateSettingsRequest) (model.Settings, error) {
	return a.service.UpdateSettings(req)
}
