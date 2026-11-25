# store tests

- **TestOpen_SeedsDefaultSettings**: A fresh in-memory store has the default last session length and miss sound.
- **TestUpdateSettings_PersistsFields**: Updated settings read back unchanged.
- **TestSession_RoundTripAndLifecycleQueries**: A new session is the draft, becomes the running session once started, and is in neither query after it ends with a stored status.
- **TestGetSession_NotFound**: Looking up a missing session wraps ErrNotFound.
- **TestUpdateSessionLength_PersistsAndRejectsMissing**: The draft length can be changed, and changing a missing session errors.
- **TestChores_CRUDAndReorder**: Chores can be reordered, edited and deleted, and list in position order.
- **TestInsertChore_RejectsZeroCount**: The database refuses a chore whose count is below one.
- **TestEvents_ListedInInsertionOrderWithChoreIDs**: Events come back in ID order with timestamps and the chore ID only on done events.
- **TestDeleteSession_CascadesToChoresAndEvents**: Deleting a session removes its chores and events, proving foreign keys are enforced.
- **TestRecentChoreNames_DistinctNewestFirstStartedOnly**: Recent names are distinct, ordered by newest session, capped by the limit, and ignore the unstarted draft.
- **TestOpen_FileBacked_PersistsAcrossReopen**: Data written to a real file survives closing and reopening the store.
- **TestDefaultDBPath_UsesConfigDir**: The default DB path is clean-tracker.db inside the clean-tracker config subdirectory.
