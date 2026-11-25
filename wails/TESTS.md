# app (IPC endpoint) tests

- **TestApp_EveryEndpoint_InMemory**: Drives every App method through a full draft, run, end, summary, repeat and settings flow (including the per-action minimum) on an in-memory DB.
- **TestApp_EveryEndpoint_FileBacked**: Runs the same flow against a real on-disk SQLite file.
- **TestApp_RejectionsReachTheFrontendAsErrors**: Rule violations (empty start, undo with nothing running, blank chore name) come back as errors.
