CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    length_seconds INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    started_at TEXT,
    ended_at TEXT,
    end_status TEXT
);

CREATE TABLE IF NOT EXISTS chores (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    target_count INTEGER NOT NULL CHECK(target_count >= 1),
    position INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_chores_session ON chores(session_id, position);

CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    chore_id TEXT REFERENCES chores(id) ON DELETE CASCADE,
    at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_events_session ON events(session_id, id);

CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    last_session_minutes INTEGER NOT NULL,
    miss_sound INTEGER NOT NULL
);
