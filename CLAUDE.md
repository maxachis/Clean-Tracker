# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Clean Tracker** — desktop app for timed cleaning sessions. See `SPEC.md` for the full spec; its "Timer Model" section is the heart of the app.

Built with **Wails v2 + Go** and **Svelte 5**. All development happens in `wails/`.

## Build & Development Commands

```bash
# Go backend (from wails/)
cd wails && go build ./...        # compile
cd wails && go test ./...         # run all tests
cd wails && go vet ./...          # static analysis
cd wails && go mod tidy           # sync deps

# Frontend only
cd wails/frontend && npm run dev     # vite dev server
cd wails/frontend && npm run check   # svelte-check / tsc
```

`wails/main.go` embeds `frontend/dist`. A placeholder `dist/index.html` keeps the Go build compiling before the frontend is built; `npm run build` replaces it.

## Architecture

### Backend (`wails/`)
Flat layout (small project, single binary):

```
wails/
├── main.go          # Wails bootstrap, wires App into the runtime
├── app.go           # App struct — methods exposed to the frontend become IPC endpoints
├── store/           # SQLite persistence (modernc.org/sqlite, pure Go, no CGo)
├── model/           # Session, Chore, Event, Snapshot, Settings structs and typed status/kind constants
├── service/         # Replay (pure timer model) and the stateful session service
└── migrations/      # embedded SQL schema (//go:embed)
```

Exposed methods on `App` become callable from the frontend via `window.go.main.App.MethodName(...)`. Keep them thin: delegate to `service`, return typed structs. Methods with more than two parameters, or adjacent same-typed ones, take a request struct.

### Timer model
- `service.Replay` is a **pure function** (no DB, no `time.Now`): it rebuilds the whole execution state — misses, deadlines, status — from the session's length, start time, chores, event log and `Now`. Nothing about misses is stored.
- The stateful `service.Service` takes an injected clock (`func() time.Time`) so tests control time. It validates commands against a fresh replay, appends events, and persists `ended_at`/`end_status` once a replay yields a terminal status.
- Event kinds and session statuses are typed constants in `model`; never write bare string literals for them.

### Data
- SQLite file: `clean-tracker.db` under `os.UserConfigDir()/clean-tracker/`
- Timestamps are stored as RFC3339Nano UTC text.

## Go Code Style

- Format with `gofmt` / `goimports` — never hand-format
- Lint with `golangci-lint`; fix warnings rather than silencing
- Package names: short, lowercase, no underscores (`store`, not `data_store`)
- Wrap errors with context: `fmt.Errorf("load session %s: %w", id, err)`
- Return errors, don't panic (except unrecoverable startup config)
- Never ignore errors silently — if truly safe, `_ = f.Close() // best-effort` with a reason
- Named constants for every domain literal (statuses, limits, thresholds)

## Testing

- Tests live beside source as `*_test.go`
- Table-driven with `t.Run` subtests
- White-box (`package foo`) unless you specifically need black-box
- Use `t.TempDir()` for filesystem fixtures
- Prefer real dependencies over mocks — in-memory SQLite (`:memory:`) beats a mocked DB interface
- API/IPC-endpoint tests: include one with an in-memory DB and one against a real file-backed SQLite instance
- Each test directory gets a `TESTS.md` with a 1-2 sentence description per test; keep it current

## Cross-Platform Paths

- **Never hardcode user-data paths.** Use `os.UserConfigDir()` for the DB, `os.UserCacheDir()` for caches, `t.TempDir()` in tests
- Use `filepath.Join`, not string concatenation with `/`
- OS-specific logic goes in `_linux.go` / `_windows.go` files when it grows past a few branches

## Dependencies

- Manage via `go.mod`; run `go mod tidy` before committing
- Pin specific versions; avoid `latest` in imports
- Stdlib first, then well-maintained third-party
- SQLite: `modernc.org/sqlite` (pure Go — avoids CGo and keeps Windows cross-compile simple)

## TDD Workflow

Red → Green → Refactor. Write the failing test first, confirm it fails for the right reason, implement the minimum to pass, then clean up. Rerun `go test ./...` after every change.

## Devcontainer Safety

When running inside a devcontainer, killing processes carelessly will bring down the entire container.

- **NEVER** use broad process-killing commands: `pkill -f`, `killall`, `kill -9 -1`, `fuser -k`
- Free a port by identifying the exact PID, verifying it is not critical, then killing only that PID:
  ```bash
  lsof -ti :<PORT>
  ps -p <PID> -o pid,comm,args
  kill <PID>
  ```
- **NEVER kill PID 1** or any `vscode-server` process
- If a port is in use and you can't identify the process, ask the user

## Mistakes

- **[tooling]**: Wails CLI v2.12.0 (built with x/tools v0.30) crashes under Go 1.27 with `internal error: package "embed" without types`, even with `-skipbindings`. To run the app, build it directly: `cd wails/frontend && npm run build`, then `cd wails && go build -tags desktop,production,webkit2_41 -o <out> .` and run the binary. `wails dev` → that direct build.
